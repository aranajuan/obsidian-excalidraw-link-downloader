# Internal Architecture — obsidian-excalidraw-downloader

This document describes the internal design of the tool in detail. Its primary purpose is to serve as a reference for porting the functionality to an Obsidian plugin (TypeScript/Node.js).

---

## Table of contents

1. [High-level data flow](#1-high-level-data-flow)
2. [Package layout](#2-package-layout)
3. [Excalidraw URL format](#3-excalidraw-url-format)
4. [Encryption](#4-encryption)
5. [Static scene download (`#json=`)](#5-static-scene-download-json)
6. [Room download (`#room=`) — Socket.IO protocol](#6-room-download-room--socketio-protocol)
7. [Annotation injection](#7-annotation-injection)
8. [Idempotency and refresh logic](#8-idempotency-and-refresh-logic)
9. [Attachment folder resolution](#9-attachment-folder-resolution)
10. [Vault walker](#10-vault-walker)
11. [Error taxonomy](#11-error-taxonomy)
12. [Obsidian plugin porting notes](#12-obsidian-plugin-porting-notes)

---

## 1. High-level data flow

```
main.go
  │
  ├─ resolveAttachmentConfig()        reads .obsidian/app.json → AttachmentMode
  │
  ├─ vault.Walk()                     collects all .md file paths
  │
  └─ for each .md file:
       injector.ProcessFile()
         │
         ├─ excalidraw.ParseAll()     regex scan → []Link (unannotated URLs only)
         │
         ├─ for each Link:
         │    Processor.downloadTo()
         │      ├─ excalidraw.Download()      → static JSON path
         │      └─ excalidraw.DownloadRoom()  → WebSocket room path
         │
         └─ injectAll()              string replacement → updated .md content
              └─ injectOne()         per-URL injection with idempotency check
```

---

## 2. Package layout

### `main.go`

Entry point. Responsibilities:
- Parse CLI args (`vault-root`, `process-path`, `--refresh`)
- Read `.obsidian/app.json` to determine attachment mode (or prompt interactively)
- Set up dual logger: stdout + `_excalidraw-downloader.log`
- Invoke `vault.Walk` and `injector.ProcessFile` for each file
- Print summary counters

### `internal/vault/walker.go`

Single exported function: `Walk(root string, skipDirs ...string) ([]string, error)`

Walks the directory tree with `filepath.WalkDir`. Skips:
- Directories whose name starts with `.` (hidden: `.obsidian`, `.git`, etc.)
- Any directory whose name matches `skipDirs` (the attachment folder, to avoid scanning downloaded `.excalidraw` files)

Returns a flat slice of absolute paths to `.md` files (case-insensitive extension match).

### `internal/excalidraw/link.go`

URL parsing and idempotency pre-check.

- `ParseAll(content string, refresh bool) []Link`
  - Applies `urlRe` regex to find all Excalidraw URLs
  - Deduplicates by URL string
  - For each unique URL, checks the first occurrence to see if it is already annotated
  - In normal mode: skips URLs already followed by `([local copy`
  - In refresh mode: returns all URLs regardless of annotation state
  - Returns a `[]Link` slice; each entry has `URL`, `Kind` (`json`|`room`), `ID`, `Key`

### `internal/excalidraw/downloader.go`

Handles static (`#json=`) scenes.

- `Download(link Link, destDir string, force bool) (path string, cached bool, err error)`
  - Non-force: checks if `excalidraw-<id>.excalidraw` exists, returns `cached=true` if so
  - Force: skips cache check, downloads to `destFile + ".tmp"`, renames on success
  - GETs `https://json.excalidraw.com/api/v2/{ID}`
  - Response body is binary: `12-byte IV || ciphertext || 16-byte GCM tag`
  - Calls `Decrypt(key, data)` to get the plaintext JSON
  - Writes the plaintext to disk

### `internal/excalidraw/room_client.go`

Handles collaborative (`#room=`) scenes via Socket.IO WebSocket.

See section 6 for the full protocol description.

- `DownloadRoom(link Link, destDir string, force bool) (path string, cached bool, err error)`
  - Same force/cache pattern as `Download`
  - Connects to `wss://oss-collab.excalidraw.com/socket.io/?EIO=4&transport=websocket`
  - Exchanges Engine.IO and Socket.IO handshake messages
  - Joins the room and waits for a `client-broadcast` binary event
  - Decrypts the scene with `DecryptGCM`
  - Converts raw scene data to `.excalidraw` file format with `sceneToExcalidrawFile`

### `internal/excalidraw/crypto.go`

Two decrypt functions:

- `Decrypt(keyStr string, data []byte) ([]byte, error)`
  - For static scenes: IV is prepended to the ciphertext in the HTTP response body
  - Wire format: `[12-byte IV][ciphertext][16-byte GCM tag]`

- `DecryptGCM(keyStr string, ciphertext, iv []byte) ([]byte, error)`
  - For room broadcasts: IV and ciphertext arrive as separate Socket.IO binary attachments

Both decode `keyStr` from base64url (no padding) to get the raw AES key bytes (16 bytes = AES-128).

### `internal/injector/injector.go`

Orchestrates per-file processing and string injection.

Key types and functions described in sections 7, 8, 9.

---

## 3. Excalidraw URL format

```
https://excalidraw.com/#(json|room)=ID,KEY
```

| Component | Description |
|---|---|
| `json` or `room` | Scene type |
| `ID` | Alphanumeric scene/room identifier |
| `KEY` | Base64url (no padding) AES key; 128-bit = 16 raw bytes = ~22 base64url chars |

Regex used (`urlRe` in `link.go`):

```
https://excalidraw\.com/#(json|room)=([a-zA-Z0-9]+),([a-zA-Z0-9_\-]+)
```

Capture groups: `(1)` kind, `(2)` ID, `(3)` KEY.

---

## 4. Encryption

Excalidraw encrypts scene data client-side before sending it to its servers. The encryption key is embedded in the URL fragment (`#...`) so it is never sent to the server (browsers do not include the fragment in HTTP requests).

### Algorithm

**AES-128-GCM** (Galois/Counter Mode with Authentication)

- Key: 16 raw bytes, transmitted as base64url (no padding) in the URL
- Nonce/IV: 12 bytes
- Authentication tag: 16 bytes (appended by GCM)

### Static scene wire format (HTTP response body)

```
[12 bytes: IV][N bytes: ciphertext + 16-byte GCM tag]
```

### Room broadcast wire format (Socket.IO binary attachments)

Two separate binary frames:
- Attachment 0: ciphertext (includes GCM tag at the end)
- Attachment 1: IV (12 bytes)

### Decrypted output

For `#json=`: raw JSON matching the `.excalidraw` file format (can be written directly to disk).

For `#room=`: a JSON object with structure:
```json
{
  "type": "...",
  "payload": {
    "elements": [...]
  }
}
```

This is converted to the standard `.excalidraw` format by `sceneToExcalidrawFile`.

### `.excalidraw` file format

```json
{
  "type": "excalidraw",
  "version": 2,
  "source": "https://excalidraw.com",
  "elements": [...],
  "appState": { "viewBackgroundColor": "#ffffff" },
  "files": {}
}
```

---

## 5. Static scene download (`#json=`)

```
GET https://json.excalidraw.com/api/v2/{ID}
```

- Returns HTTP 200 with binary body (encrypted scene)
- Returns HTTP 404 if the scene was deleted
- No authentication required
- Response is stable as long as the scene exists (can be cached indefinitely)

Flow in `Download`:

```
1. Check if destFile exists → return cached (unless force=true)
2. GET https://json.excalidraw.com/api/v2/{ID}
3. ReadAll(body) → encrypted bytes
4. Decrypt(KEY, bytes) → plaintext JSON
5. WriteFile(tmpFile or destFile, plaintext)
6. If force: Rename(tmpFile, destFile)
```

---

## 6. Room download (`#room=`) — Socket.IO protocol

Room data is not stored on a server — it lives in the **browser's in-memory state and localStorage**. The server only relays real-time messages between connected clients.

### WebSocket endpoint

```
wss://oss-collab.excalidraw.com/socket.io/?EIO=4&transport=websocket
```

HTTP header required: `Origin: https://excalidraw.com`

### Protocol stack

```
WebSocket
  └── Engine.IO v4 (EIO=4)
        └── Socket.IO v4
```

### Handshake sequence

```
← "0{...}"          Engine.IO OPEN packet (server sends session info)
→ "40"              Socket.IO connect to default namespace "/"
← "40{...}"         Socket.IO connect acknowledgement
→ 42["join-room","ROOM_ID"]   join the Excalidraw room
```

### After joining

**Case A — another participant is already in the room:**

The server sends a `new-user` event to the existing participant. That participant's browser immediately broadcasts the current scene back to the new joiner.

```
← 45N-["client-broadcast",{placeholder},...] + N binary frames
```

**Case B — we are the first (and only) participant:**

```
← 42["first-in-room"]
```

The tool opens the Excalidraw URL in the default system browser (`open url` on macOS, `xdg-open` on Linux). The browser:
1. Loads excalidraw.com
2. Restores the drawing from localStorage
3. Connects to the same room
4. Triggers a `new-user` event back to the tool
5. Broadcasts the scene

The tool then receives the `client-broadcast` event. Initial deadline: 30 s. After the browser opens: 60 s from that moment.

### `client-broadcast` binary event

Socket.IO binary event wire format:

```
Text frame:   "45N-["client-broadcast",{placeholder_0},...,{placeholder_N-1}]"
Binary frame: one WebSocket binary message per attachment, in order
```

For `client-broadcast`: N=2
- Attachment 0 = ciphertext
- Attachment 1 = IV

### EIO ping/pong

The server sends `"2"` (ping); the client must respond with `"3"` (pong) to keep the connection alive during long waits.

### Empty canvas

If the decrypted scene has `payload.elements = []`, `ErrEmptyCanvas` is returned. This means:
- The room exists but the drawing was created on a different browser/machine
- That browser is not currently connected
- Solution: ask the author to open the room URL while running the tool

Empty canvas is **not retried** in normal mode (it's not a transient network error). In `--refresh` mode it is retried.

### Browser tab management (macOS)

After a successful room download where the browser was auto-opened, the tool closes the Excalidraw tab using AppleScript. Supported browsers: Chrome, Chromium, Edge, Brave, Arc, Safari.

On Linux and Windows no tab-close mechanism is implemented (would require browser-specific tooling).

---

## 7. Annotation injection

### `injectOne(content, url, annotation string, refresh bool) string`

Scans `content` for every occurrence of `url` and inserts `annotation` after it. Handles two URL embedding styles:

**Bare URL:**
```
text https://excalidraw.com/#room=abc,KEY text
→ text https://excalidraw.com/#room=abc,KEY ([local copy 24/03/2026](path)) text
```

**Markdown link:**
```
[title](https://excalidraw.com/#room=abc,KEY)
→ [title](https://excalidraw.com/#room=abc,KEY) ([local copy 24/03/2026](path))
```

Detection of markdown link: checks if the character immediately after the URL is `)`. If so, consumes it before checking for an existing annotation and then writes it back before the new annotation.

### Annotation state machine (per occurrence)

After consuming the optional `)`, the text immediately following (trimmed of spaces/tabs) is checked:

| Trimmed prefix | Action |
|---|---|
| `([local copy` | Normal: skip (already annotated). Refresh: strip old annotation, write new one. |
| `[local copy` | Old format without parens: strip old, write new (backward compat). |
| `(⚠` | Strip old warning, write new annotation (retry). |
| anything else | Write new annotation. |

### Annotation format

```
([local copy DD/MM/YYYY](relative/path/to/excalidraw-ID.excalidraw))
```

- Date: `time.Now().Format("02/01/2006")` → `DD/MM/YYYY`
- Path: `filepath.Rel(noteDir, destFile)` converted to forward slashes

### `indexAfterMarkdownLink(s string) int`

Helper used to skip over `[text](url)` constructs when stripping old annotations. Finds `](` then the next `)`.

---

## 8. Idempotency and refresh logic

### Normal mode idempotency

Three levels:

1. **File level** (`downloader.go`): `os.Stat(destFile)` — if the `.excalidraw` file exists, return it without a network request.

2. **Parse level** (`link.go`): `ParseAll` checks the first occurrence of each URL in the file content. If it is followed by `([local copy`, the URL is excluded from the returned list — no download attempt is made at all.

3. **Injection level** (`injector.go`): `injectOne` checks each occurrence before writing. Even if `ParseAll` somehow returned an already-annotated URL, the injection would be a no-op.

### Refresh mode

`--refresh` disables levels 1 and 2:

- `force=true` passed to `Download`/`DownloadRoom` → skips `os.Stat` check
- `refresh=true` passed to `ParseAll` → skips the `([local copy` skip
- `refresh=true` passed to `injectOne` → replaces `([local copy` annotation instead of skipping

### Refresh failure safety

When `force=true`, downloads go through a temp file:

```
Download → write to destFile.tmp
  ├── success → os.Rename(destFile.tmp, destFile)   ← atomic on same filesystem
  └── failure → os.Remove(destFile.tmp)             ← original destFile untouched
```

Additionally, `ProcessFile` tracks `hadLocalCopy[url]` before the download loop. If a refresh download fails for a URL that previously had a local copy:
- The URL is **not** added to the `annotations` map
- `injectOne` is therefore **never called** for that URL
- The original `([local copy DATE](path))` annotation is preserved in the file

---

## 9. Attachment folder resolution

Obsidian's `attachmentFolderPath` value in `.obsidian/app.json` maps to one of four modes:

| Value | Mode | `resourcesDirFor(notePath)` |
|---|---|---|
| `""` | `ModeVaultRoot` | `vaultPath` |
| `"/folder"` | `ModeFixedFolder` | `vaultPath/folder` |
| `"./"` or `"."` | `ModeSameAsNote` | `filepath.Dir(notePath)` |
| `"./folder"` | `ModeSubfolder` | `filepath.Dir(notePath)/folder` |

Any other format (plain name without prefix) is treated as `ModeFixedFolder`.

If `.obsidian/app.json` is absent or malformed, the user is prompted interactively.

The attachment directory is created with `os.MkdirAll` before any download.

---

## 10. Vault walker

`vault.Walk(root string, skipDirs ...string) ([]string, error)`

Uses `filepath.WalkDir` (Go's efficient directory traversal). On each entry:
- Skip if directory name starts with `.`
- Skip if directory name is in `skipDirs` (the attachment folder name is passed here to prevent scanning `.excalidraw` files)
- Collect if file extension is `.md` (case-insensitive)

`skipDirs` receives only the **base name** of the attachment folder (not a full path), so nested folders with the same name are also skipped. This is intentional: for `ModeSubfolder` the attachment folder appears inside each note's directory.

---

## 11. Error taxonomy

| Error | Source | Retry behaviour |
|---|---|---|
| HTTP non-200 | `Download` | Normal: once. `--refresh`: once. Room: retried indefinitely on transient errors. |
| Network timeout / connection refused | `Download`, `DownloadRoom` | Room links: retried every 5 s indefinitely. JSON links: no retry. |
| Decrypt failure | `Decrypt`, `DecryptGCM` | Not retried (corrupted data, wrong key — permanent). |
| `ErrEmptyCanvas` | `DownloadRoom` | Not retried in normal mode. Retried in `--refresh` mode. |
| File write failure | `Download`, `DownloadRoom` | Not retried. |
| `.md` read/write failure | `ProcessFile` | Not retried. Counted as error. |

Room links have indefinite retry in `downloadTo` because the room server is eventually consistent: the room may become populated at any moment while the tool is running.

---

## 12. Obsidian plugin porting notes

An Obsidian plugin would replace the CLI scaffolding (`main.go`, vault walker) with Obsidian's native APIs, while the core logic (URL parsing, crypto, download, injection) remains the same in concept.

### What maps directly

| CLI component | Obsidian equivalent |
|---|---|
| `vault.Walk` | `app.vault.getMarkdownFiles()` |
| `os.ReadFile` / `os.WriteFile` | `app.vault.read()` / `app.vault.modify()` |
| `os.MkdirAll` | `app.vault.createFolder()` |
| `filepath.Rel` | Manual relative path from `TFile.path` |
| `.obsidian/app.json` | `app.vault.getConfig("attachmentFolderPath")` |
| Log file | Obsidian's `console.log` or a Notice |
| `--refresh` flag | A button or command in the plugin UI |

### Crypto

The Web Crypto API (`crypto.subtle`) supports AES-GCM natively. The Go `Decrypt` function maps to:

```typescript
const keyBytes = base64urlDecode(keyStr); // no padding
const cryptoKey = await crypto.subtle.importKey(
  "raw", keyBytes, { name: "AES-GCM" }, false, ["decrypt"]
);
// Static scene: IV = data.slice(0, 12), ciphertext = data.slice(12)
const plaintext = await crypto.subtle.decrypt(
  { name: "AES-GCM", iv: data.slice(0, 12) },
  cryptoKey,
  data.slice(12)
);
```

### Static scene HTTP fetch

```typescript
const resp = await requestUrl(`https://json.excalidraw.com/api/v2/${id}`);
// resp.arrayBuffer → decrypt → JSON text → write to vault
```

Obsidian's `requestUrl` bypasses browser CORS restrictions, which matters for the Excalidraw API.

### Room download via WebSocket

WebSocket is available in Obsidian's Electron environment. The Socket.IO handshake is lightweight enough to implement from scratch (the protocol used is a small subset of Socket.IO v4 — no need for the full `socket.io-client` library).

Key steps:
1. `new WebSocket("wss://oss-collab.excalidraw.com/socket.io/?EIO=4&transport=websocket", [], { headers: { Origin: "https://excalidraw.com" } })`
2. Handle text vs. binary frames (binary frames arrive as `ArrayBuffer` in `onmessage`)
3. Reassemble the two binary attachments (ciphertext, IV)
4. Decrypt with `crypto.subtle`

The "open in browser" fallback for empty rooms uses `window.open(url)` in Obsidian's Electron context or `app.openWithDefaultApp(url)`.

### Annotation injection

Pure string manipulation — ports directly. The regex and the state machine in `injectOne` have no Go-specific dependencies.

TypeScript regex equivalent:
```typescript
const urlRe = /https:\/\/excalidraw\.com\/#(json|room)=([a-zA-Z0-9]+),([a-zA-Z0-9_\-]+)/g;
```

### Idempotency

Same logic applies. Use `String.prototype.indexOf` / `String.prototype.startsWith` for the annotation checks.

### `.excalidraw` file output

Room scenes need the same wrapping as `sceneToExcalidrawFile`:
```typescript
const file = {
  type: "excalidraw",
  version: 2,
  source: "https://excalidraw.com",
  elements: scene.payload.elements,
  appState: { viewBackgroundColor: "#ffffff" },
  files: {}
};
await vault.create(destPath, JSON.stringify(file, null, 2));
```

### Plugin UI suggestions

- Ribbon icon or command palette entry: **"Download Excalidraw links"**
- Settings tab: attachment folder override (in case auto-detection from `app.json` is insufficient)
- Modal or status bar progress for long-running room downloads
- **"Refresh"** toggle in the command or a separate command: **"Refresh all Excalidraw local copies"**
- Per-file command: **"Download Excalidraw links in current file"** — processes only the active note
- Notice on completion with the summary counters

### Key constraints to keep in mind

- Obsidian's mobile app does not support WebSocket to arbitrary servers (sandbox restrictions). Room download would be desktop-only.
- Static scene (`#json=`) download works on mobile via `requestUrl`.
- File paths in vault are always forward-slash, vault-root-relative (e.g., `folder/note.md`) — no leading slash.
- `app.vault.getConfig("attachmentFolderPath")` may return `undefined` if the user has never changed the setting (equivalent to `""`, i.e., vault root).
- Renaming/moving a note after annotation will break the relative path in the `([local copy ...](...))` link. A `rename` event handler could update annotations automatically.

# Excalidraw Link Downloader

Obsidian plugin that downloads Excalidraw shared scenes and collaborative rooms to local `.excalidraw` files, with automatic annotation so you can open them directly from your notes.

---

## Why?

Excalidraw links in your notes are fragile:
- Collaborative rooms expire when no one has them open
- Static scenes can vanish if the author deletes them

This plugin creates permanent local backups and annotates your notes so the drawings stay accessible forever.

---

## Installation

**Option 1: Community Plugins (Recommended)**
1. Open **Settings → Community Plugins**
2. Search for "Excalidraw Link Downloader"
3. Click Install

**Option 2: BRAT (Beta)**
1. Install the [BRAT](https://obsidian.md/plugins?id=obsidian42-brat) plugin
2. In BRAT settings, add: `https://github.com/aranajuan/obsidian-excalidraw-link-downloader`
3. Enable "Excalidraw Link Downloader"

---

## Usage

### Commands

Open the note with Excalidraw links, then run one of these from the command palette:

- **Download Excalidraw links in current file** — Finds and downloads all Excalidraw links in the active note
- **Refresh Excalidraw local copies in current file** — Re-downloads existing links to get the latest version

### How it works

When you run the download command, the plugin:

1. Scans the current note for Excalidraw URLs (`#json=` and `#room=` formats)
2. Downloads each drawing to your vault
3. Adds an annotation showing the local file path

Your notes transform from:

```
[diseño del sistema](https://excalidraw.com/#room=abc123,KEY)
```

To:

```
[diseño del sistema](https://excalidraw.com/#room=abc123,KEY) ([local copy 18/05/2026](_resources/excalidraw-abc123.excalidraw))
```

Click the local copy link to open the drawing directly in Obsidian (requires [Embedded Notes](https://github.com/obsidian-community/obsidian-embedded-notes) or similar plugin for `.excalidraw` rendering, or simply Ctrl+click to open as text).

### Error handling

If a download fails (expired room, network error), the annotation shows a warning:

```
https://excalidraw.com/#room=abc,KEY (⚠ download failed)
```

Retrying later may succeed if the room is reopened.

---

## Settings

In **Settings → Excalidraw Link Downloader**:

| Setting | Description |
|---------|-------------|
| Attachment folder override | Override where `.excalidraw` files are saved. Leave empty to use your vault's default attachment location. Format: empty = vault root, `/folder` = fixed folder, `./` = same as note, `./folder` = subfolder of note |

---

## Link types supported

### Static scenes (`#json=`)

```
https://excalidraw.com/#json=ID,KEY
```

Fetched from Excalidraw's static API. These persist until the author deletes them.

### Collaborative rooms (`#room=`)

```
https://excalidraw.com/#room=ID,KEY
```

Connected via WebSocket. The plugin opens your browser to restore the scene from localStorage, then downloads it. Works best when the room is still active.

---

## Idempotency

Running the download multiple times is safe:
- Already downloaded files are skipped
- Links already annotated are skipped
- Failed links are retried

Use **Refresh** to force re-download and update the date.

---

## Also available: CLI tool

This repository also includes a **Go CLI** for batch processing entire vaults:

```bash
go build -o excalidraw-downloader ./cli
./excalidraw-downloader ~/Documents/MyVault
```

See [`cli/README.md`](cli/README.md) for full CLI documentation.
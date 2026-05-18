# Release Readiness: Obsidian Plugin + Go CLI

## TL;DR

> **Quick Summary**: Fix all CRITICAL and HIGH severity issues in the Obsidian Excalidraw Link Downloader plugin and its companion Go CLI to prepare for public release on the Obsidian Community Plugin registry.
>
> **Deliverables**:
> - 5 CRITICAL plugin bug fixes (infinite retry, missing onunload, no abort signals, pinned dependency, safe API access)
> - 10 HIGH issue fixes (response size limit, Socket.IO validation, Spanish→English migration, versions.json, error visibility, popup blocker, notice leak, README annotation fix, screenshots, error state docs)
> - Go CLI fixes (refresh failure handling, go.mod, translations, test infrastructure)
> - Unit test suite for core modules (TypeScript + Go)
> - Obsidian community submission compliance (manifest, README, versions.json)
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 4 waves
> **Critical Path**: Task 1 → Task 8 → Task 14 → Task 19 → F1-F4

---

## Context

### Original Request
Deep audit of the Obsidian plugin for security, functionality, and release readiness. Fix everything needed before releasing to the public on the Obsidian community plugin registry.

### Interview Summary
**Key Discussions**:
- Full 5-area audit completed: plugin source, security/crypto, docs/release, Obsidian requirements, Go CLI consistency
- 6 CRITICAL + 10 HIGH issues found across plugin and CLI
- One false positive identified by Oracle: `platform: 'node'` in esbuild is correct (room-client.ts requires `ws` Node.js package)
- Spanish→English translation requires migration strategy to preserve idempotency

**Research Findings**:
- Obsidian community plugin submission requires: LICENSE, README, manifest.json (description ≤250 chars, ends with `.`), main.js as GitHub release asset, `onunload()` cleanup, `normalizePath()` for paths, no `innerHTML` with user input
- AES-128-GCM crypto implementation is correct (Web Crypto API, proper IV/tag handling)
- No XSS vectors in injector (string-based, no DOM)
- Regex is ReDoS-safe (no nested quantifiers)
- `isDesktopOnly: true` is correct (uses `window.open` + `ws`)

### Metis Review
**Identified Gaps** (addressed):
- esbuild `platform:'node'` is correct — NOT a bug. Changed to "document, don't change"
- Spanish→English translation breaks idempotency — need migration logic for old Spanish annotations
- `versions.json` format needs verification against Obsidian standards
- WebSocket room interactions are hard to test without a real server — use mocks

---

## Work Objectives

### Core Objective
Fix all blocking bugs and submission requirements so the plugin can be published on the Obsidian Community Plugin registry and the Go CLI can be released as a companion tool.

### Concrete Deliverables
- Plugin passes Obsidian community submission review
- All CRITICAL bugs fixed (infinite retry, no cleanup, no abort, pinned deps, safe API)
- All HIGH bugs fixed (size limits, validation, translations, docs, notice lifecycle)
- Go CLI aligned with plugin behavior (refresh failure handling, translations)
- Unit tests for core logic modules
- README with correct examples, screenshots, error state docs

### Definition of Done
- [ ] `npm run build` succeeds
- [ ] `cd cli && go build ./...` succeeds
- [ ] `cd cli && go test ./...` passes
- [ ] Plugin loads in Obsidian test vault without errors
- [ ] All 5 CRITICAL and 10 HIGH issues verified as fixed
- [ ] README annotation format matches actual code output
- [ ] `versions.json` exists with correct format
- [ ] No Spanish strings in user-facing output (plugin or CLI)

### Must Have
- Infinite retry loop capped at maximum attempts
- `onunload()` method with proper cleanup (AbortController)
- Request size limits preventing OOM
- Socket.IO broadcast validation using JSON.parse
- Spanish→English translation with backward-compatible migration
- `versions.json` for Obsidian submission
- Correct README annotation format
- Pinned `obsidian` dependency
- Safe replacement for `(this.app.vault as any).getConfig()`
- Go CLI refresh failure annotation matching TS behavior
- Unit tests for link parsing, crypto, and injector state machine

### Must NOT Have (Guardrails)
- Do NOT change `platform: 'node'` to `'browser'` in esbuild config — it's correct for `ws`
- Do NOT change the annotation format `([[...|local copy DATE]])` — existing users depend on it
- Do NOT add i18n library — string replacement is sufficient
- Do NOT add CI/CD pipeline — out of scope
- Do NOT add new features — only bug fixes and release requirements
- Do NOT change the WebSocket protocol handling — it works, just needs bounds
- Do NOT modify crypto implementation — it's correct
- Do NOT touch MEDIUM/LOW severity issues
- Do NOT change Go module path — it's a CLI, not a library
- Do NOT add configuration for retry limits — hardcode reasonable defaults (max 5 for room, 3 for JSON)

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO (plugin has no test framework)
- **Automated tests**: YES (tests-after)
- **Framework**: 
  - Plugin: Vitest (lightweight, ESM-native, works well with TypeScript)
  - CLI: Go standard `testing` package (table-driven tests)
- **If TDD**: Tests written AFTER bug fixes as regression tests

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.omo/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Plugin logic**: Use Bash (node/vitest) — import modules, call functions, assert output
- **Go CLI**: Use Bash (go test) — run tests, verify pass/fail
- **Obsidian plugin**: Use Bash — manifest validation, build success, file existence
- **Documentation**: Use Bash — grep for Spanish strings, verify annotation format

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — foundation + independent fixes):
├── Task 1: Add onunload() + AbortController to plugin [deep]
├── Task 2: Cap infinite retry loop + add max attempts [quick]
├── Task 3: Add response size limit to HTTP and WS [quick]
├── Task 4: Fix Socket.IO broadcast validation with JSON.parse [quick]
├── Task 5: Pin obsidian dependency + add comment to esbuild config [quick]
├── Task 6: Replace (this.app.vault as any).getConfig with safe API [quick]
└── Task 7: Fix Notice lifecycle (duration 0 leak) [quick]

Wave 2 (After Wave 1 — depends on abort controller + retry fix):
├── Task 8: Spanish→English translation with migration [deep]
├── Task 9: Fix window.open popup blocker issue [unspecified-high]
├── Task 10: Surface errors via Notice (not just console) [quick]
├── Task 11: Fix unsafe .buffer as ArrayBuffer in writeBinary [quick]
├── Task 12: Add versions.json for Obsidian submission [quick]
├── Task 13: Fix Go CLI: refresh failure annotation + go.mod [quick]
└── Task 14: Fix Go CLI: Spanish→English prompts and annotations [quick]

Wave 3 (After Wave 2 — docs + tests, mostly independent):
├── Task 15: Fix README annotation format + error state docs [writing]
├── Task 16: Add screenshots/demo to README [visual-engineering]
├── Task 17: Add Go CLI --yes flag + closeBrowserTab warning [quick]
├── Task 18: Set up Vitest for plugin + write core tests [deep]
├── Task 19: Write Go unit tests for core modules [deep]
└── Task 20: Final Obsidian submission compliance check [quick]

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
→ Present results → Get explicit user okay

Critical Path: Task 1 → Task 8 → Task 15 → Task 20 → F1-F4 → user okay
Parallel Speedup: ~65% faster than sequential
Max Concurrent: 7 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks |
|------|-----------|--------|
| 1 | - | 8, 9 |
| 2 | - | 8 |
| 3 | - | - |
| 4 | - | - |
| 5 | - | - |
| 6 | - | - |
| 7 | - | - |
| 8 | 1, 2 | 15 |
| 9 | 1 | - |
| 10 | - | - |
| 11 | - | - |
| 12 | - | - |
| 13 | - | 14 |
| 14 | 13 | - |
| 15 | 8 | 20 |
| 16 | 15 | - |
| 17 | - | - |
| 18 | 1, 2, 4, 11 | 20 |
| 19 | 13 | 20 |
| 20 | 15, 18, 19 | F1-F4 |

### Agent Dispatch Summary

- **Wave 1**: 7 tasks — T1→`deep`, T2-T7→`quick`
- **Wave 2**: 7 tasks — T8→`deep`, T9→`unspecified-high`, T10-T14→`quick`
- **Wave 3**: 6 tasks — T15→`writing`, T16→`visual-engineering`, T17→`quick`, T18→`deep`, T19→`deep`, T20→`quick`
- **FINAL**: 4 tasks — F1→`oracle`, F2→`unspecified-high`, F3→`unspecified-high`, F4→`deep`

---

## TODOs

- [x] 1. Add onunload() + AbortController to plugin

  **What to do**:
  - Add a class property `private abortController = new AbortController()` to `ExcalidrawDownloaderPlugin` in `src/main.ts`
  - Implement `onunload()` method that calls `this.abortController.abort()` and hides any active `progressNotice`
  - Thread the `AbortSignal` through `download()` → `downloadRoomWithRetry()` → `downloadRoom()` in `src/downloader.ts`
  - In `downloadRoomWithRetry`, check `signal.aborted` at the start of each retry iteration and between await points
  - In `room-client.ts`, accept `signal: AbortSignal` parameter; add `signal.addEventListener('abort', () => ws.close())` to the WebSocket connection
  - Clear any pending `setTimeout` timers in `downloadRoom` when signal fires
  - Pass `signal` from `main.ts` `processFile()` down through all call chains

  **Must NOT do**:
  - Do NOT add AbortSignal to `requestUrl()` calls (Obsidian's API doesn't support it — use `Promise.race` with signal instead)
  - Do NOT change the WebSocket protocol or message format
  - Do NOT add configuration for abort timeouts

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires understanding async flow across 3 files, threading a signal through multiple call chains
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2-7)
  - **Blocks**: Tasks 8, 9
  - **Blocked By**: None

  **References**:
  - `src/main.ts:33-148` — Plugin class, onload, processFile. Add onunload here
  - `src/downloader.ts:79-86` — The infinite retry loop that needs abort signal check
  - `src/room-client.ts:1-80` — WebSocket connection that needs signal-based cleanup
  - `src/room-client.ts:55-76` — setTimeout timers that need clearing on abort

  **Acceptance Criteria**:
  - [ ] `onunload()` method exists in plugin class and calls `abortController.abort()`
  - [ ] `AbortSignal` is threaded from `processFile` through `download` to `downloadRoomWithRetry` to `downloadRoom`
  - [ ] WebSocket connection closes when `signal.aborted` becomes true
  - [ ] Pending `setTimeout` timers are cleared on abort
  - [ ] Plugin compiles without TypeScript errors: `npx tsc --noEmit`

  **QA Scenarios**:
  ```
  Scenario: Plugin loads and unloads cleanly
    Tool: Bash
    Preconditions: Plugin built with npm run build
    Steps:
      1. Copy main.js, manifest.json to test vault .obsidian/plugins/excalidraw-link-downloader/
      2. Open Obsidian with test vault
      3. Enable plugin via settings
      4. Disable plugin via settings
      5. Check Obsidian developer console for errors
    Expected Result: No errors in console, no lingering WebSocket connections, no hanging notices
    Failure Indicators: "WebSocket" errors, "uncaught promise" errors, persistent notices
    Evidence: .omo/evidence/task-1-load-unload.txt

  Scenario: Abort stops active download
    Tool: Bash
    Preconditions: Plugin installed, note with #room= link open
    Steps:
      1. Start download via command palette
      2. Immediately disable plugin
      3. Check that no retry loop continues in console
    Expected Result: Download stops, no continued retry attempts
    Failure Indicators: Continued "intento N" messages in console after disable
    Evidence: .omo/evidence/task-1-abort-download.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix(plugin): add onunload with AbortController for cleanup`
  - Files: `src/main.ts`, `src/downloader.ts`, `src/room-client.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 2. Cap infinite retry loop with maximum attempts

  **What to do**:
  - In `src/downloader.ts`, change `downloadRoomWithRetry` from `for (let attempt = 2; ; attempt++)` to `for (let attempt = 2; attempt <= MAX_ROOM_RETRIES; attempt++)`
  - Add `const MAX_ROOM_RETRIES = 5` constant at top of file
  - After the loop exits without return, throw a `RoomError` with message "room download failed after {MAX_ROOM_RETRIES} attempts"
  - In `src/downloader.ts`, add max retry count for JSON downloads too (currently `downloadJson` is called once in `download()` — add 3 retries with exponential backoff: 1s, 2s, 4s)
  - Ensure error messages are user-facing (will surface via Notice, not just console)

  **Must NOT do**:
  - Do NOT add configuration for retry limits — hardcode `MAX_ROOM_RETRIES = 5`
  - Do NOT change the retry interval (3s for room is fine)
  - Do NOT remove the retry mechanism entirely — rooms need retries for initialization

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small, focused change to one file
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3-7)
  - **Blocks**: Task 8
  - **Blocked By**: None

  **References**:
  - `src/downloader.ts:79-86` — Current infinite loop `for (let attempt = 2; ; attempt++)`
  - `src/downloader.ts:88-96` — `downloadJson` function that has no retry
  - `src/downloader.ts:67-77` — RoomError class and ErrEmptyCanvas

  **Acceptance Criteria**:
  - [ ] `downloadRoomWithRetry` has max 5 attempts
  - [ ] After max attempts, throws descriptive error
  - [ ] `downloadJson` has 3 retries with backoff
  - [ ] TypeScript compiles: `npx tsc --noEmit`

  **QA Scenarios**:
  ```
  Scenario: Room retry stops after 5 attempts
    Tool: Bash (node)
    Preconditions: Test with invalid room ID
    Steps:
      1. Create a vault with a #room=invalidID,invalidKey link
      2. Run download command
      3. Observe Notices
    Expected Result: Notice shows "attempt 5" then stops, does not continue indefinitely
    Failure Indicators: Notices continue past attempt 5
    Evidence: .omo/evidence/task-2-max-retries.txt

  Scenario: JSON download retries on transient failure
    Tool: Bash (node)
    Preconditions: Network available, valid #json= link
    Steps:
      1. Download valid Excalidraw scene
      2. Verify completion
    Expected Result: Download succeeds within 3 attempts for valid URL
    Failure Indicators: Download fails without retry
    Evidence: .omo/evidence/task-2-json-retry.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix(plugin): cap room retry at 5 attempts, add JSON retry with backoff`
  - Files: `src/downloader.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 3. Add response size limit to HTTP and WS downloads

  **What to do**:
  - In `src/downloader.ts`, after `requestUrl` returns, check `resp.arrayBuffer.byteLength` before passing to `decrypt()`
  - Add `const MAX_DOWNLOAD_BYTES = 50 * 1024 * 1024` (50 MB) constant
  - If response exceeds limit, throw `Error(response too large: ${byteLength} bytes, max ${MAX_DOWNLOAD_BYTES})`
  - In `src/room-client.ts`, add similar size check when accumulating binary frames: if total size exceeds `MAX_DOWNLOAD_BYTES`, reject promise
  - In Go CLI `cli/internal/excalidraw/downloader.go`, add `io.LimitReader` wrapping `resp.Body` with same 50MB limit
  - In Go CLI `cli/internal/excalidraw/room_client.go`, add size check when accumulating broadcast data

  **Must NOT do**:
  - Do NOT change the actual download protocol
  - Do NOT add configuration for size limits
  - Do NOT skip the check for "small" responses — always check

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple size check addition, same pattern in 4 locations
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-2, 4-7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `src/downloader.ts:89-96` — `downloadJson` function where size check goes after line 91
  - `src/room-client.ts:95-120` — Binary frame accumulation where size check goes
  - `cli/internal/excalidraw/downloader.go:39-49` — Go HTTP download
  - `cli/internal/excalidraw/room_client.go:90-115` — Go WS frame accumulation

  **Acceptance Criteria**:
  - [ ] TypeScript: response size checked before `decrypt()` in `downloadJson`
  - [ ] TypeScript: total binary frame size checked in `room-client.ts`
  - [ ] Go: `io.LimitReader` wraps response body
  - [ ] Go: broadcast data size checked in room client
  - [ ] Both languages compile: `npx tsc --noEmit` and `cd cli && go build ./...`

  **QA Scenarios**:
  ```
  Scenario: Large response is rejected
    Tool: Bash (node)
    Preconditions: Plugin loaded
    Steps:
      1. Mock or simulate a response >50MB (unit test with mocked requestUrl)
      2. Verify error is thrown, not OOM crash
    Expected Result: Error message "response too large: X bytes, max 52428800 bytes"
    Failure Indicators: No error thrown, or crash/out-of-memory
    Evidence: .omo/evidence/task-3-size-limit.txt

  Scenario: Normal downloads still work
    Tool: Bash
    Preconditions: Valid #json= link in vault
    Steps:
      1. Download valid Excalidraw scene
      2. Verify file saved and annotation injected
    Expected Result: Download succeeds, file <50MB saved correctly
    Failure Indicators: Download fails with size error for normal-sized scenes
    Evidence: .omo/evidence/task-3-normal-download.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix: add 50MB response size limit to prevent OOM`
  - Files: `src/downloader.ts`, `src/room-client.ts`, `cli/internal/excalidraw/downloader.go`, `cli/internal/excalidraw/room_client.go`
  - Pre-commit: `npx tsc --noEmit && cd cli && go build ./...`

- [x] 4. Fix Socket.IO broadcast validation with JSON.parse

  **What to do**:
  - In `src/room-client.ts`, replace `isBroadcast = text.slice(dashIdx + 1).includes('"client-broadcast"')` with proper JSON parsing:
    ```typescript
    try {
      const parts = JSON.parse(text.slice(dashIdx + 1));
      isBroadcast = Array.isArray(parts) && parts[0] === 'client-broadcast';
    } catch { isBroadcast = false; }
    ```
  - In Go CLI `cli/internal/excalidraw/room_client.go`, make equivalent fix: parse the Socket.IO packet as JSON array and check index 0 equals "client-broadcast" instead of using `strings.Contains`

  **Must NOT do**:
  - Do NOT change the Socket.IO protocol handling beyond this fix
  - Do NOT add logging for every packet

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small, targeted fix in 2 files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-3, 5-7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `src/room-client.ts:143` — Current `includes('"client-broadcast"')` check
  - `cli/internal/excalidraw/room_client.go` — Go equivalent of the broadcast detection

  **Acceptance Criteria**:
  - [ ] TypeScript uses `JSON.parse` + array check instead of `includes`
  - [ ] Go uses `json.Unmarshal` + array index check instead of `strings.Contains`
  - [ ] Both compile without errors
  - [ ] Malformed Socket.IO packets are caught and handled gracefully (no crash)

  **QA Scenarios**:
  ```
  Scenario: Malformed packet doesn't crash
    Tool: Bash (node - unit test)
    Preconditions: None
    Steps:
      1. Test with text containing "client-broadcast" in data context (not event name)
      2. Verify it's correctly identified as non-broadcast
    Expected Result: Packet with "client-broadcast" in data (not event name) is not treated as broadcast
    Failure Indicators: False positive broadcast detection
    Evidence: .omo/evidence/task-4-socketio-validation.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix: validate Socket.IO broadcast with JSON.parse instead of includes`
  - Files: `src/room-client.ts`, `cli/internal/excalidraw/room_client.go`
  - Pre-commit: `npx tsc --noEmit && cd cli && go build ./...`

- [x] 5. Pin obsidian dependency + document esbuild platform choice

  **What to do**:
  - In `package.json`, change `"obsidian": "latest"` to `"obsidian": "^1.4.0"` (matching minAppVersion in manifest)
  - In `esbuild.config.mjs`, add a comment above `platform: 'node'` explaining why:
    ```javascript
    // platform: 'node' is correct for this plugin because room-client.ts
    // imports 'ws' (Node.js WebSocket library) for custom Origin headers.
    // Browser WebSocket API doesn't support custom headers.
    // Do NOT change to 'browser' — it will break room downloads.
    ```
  - In `package.json`, add `description`, `keywords`, `license`, `author` fields:
    ```json
    "description": "Downloads Excalidraw shared scenes and rooms to local .excalidraw files",
    "license": "MIT",
    "author": "Your Name"
    ```

  **Must NOT do**:
  - Do NOT change `platform: 'node'` to `'browser'`
  - Do NOT change `target: 'node16'` — it's correct for Electron
  - Do NOT upgrade other dependencies

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple config changes with explanatory comment
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-4, 6-7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `package.json:16-17` — `"obsidian": "latest"` to pin
  - `esbuild.config.mjs:10-11` — `platform: 'node'` to document
  - `manifest.json:3` — `"minAppVersion": "1.4.0"` to match

  **Acceptance Criteria**:
  - [ ] `package.json` has `"obsidian": "^1.4.0"` (not `"latest"`)
  - [ ] `esbuild.config.mjs` has comment explaining `platform: 'node'`
  - [ ] `package.json` has `description`, `license`, `author` fields
  - [ ] Plugin still builds: `npm run build`

  **QA Scenarios**:
  ```
  Scenario: Build succeeds with pinned dependency
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `rm -rf node_modules package-lock.json`
      2. Run `npm install`
      3. Run `npm run build`
    Expected Result: Build succeeds, main.js produced
    Failure Indicators: npm install fails, build fails
    Evidence: .omo/evidence/task-5-pinned-deps.txt

  Scenario: esbuild comment exists
    Tool: Bash (grep)
    Steps:
      1. grep "platform: 'node'" esbuild.config.mjs
      2. Verify comment above it explains ws dependency
    Expected Result: Comment found explaining why platform:node is needed
    Failure Indicators: No comment, or comment says "change to browser"
    Evidence: .omo/evidence/task-5-esbuild-comment.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix(deps): pin obsidian dep, document esbuild platform choice`
  - Files: `package.json`, `esbuild.config.mjs`
  - Pre-commit: `npm run build`

- [x] 6. Replace (this.app.vault as any).getConfig with safe API

  **What to do**:
  - In `src/main.ts`, replace `(this.app.vault as any).getConfig?.('attachmentFolderPath')` with `this.app.getConfig('attachmentFolderPath')`
  - The `getConfig` method exists on `App` (not `Vault`) in Obsidian 1.4+. This is the documented public API.
  - Add a fallback: if `this.app.getConfig` is undefined (older Obsidian), return empty string (vault root)
  - Remove the `as any` cast entirely
  - Also check `src/main.ts:68` — the `parseAttachmentFolderPath` function that processes this — ensure it handles `undefined` return

  **Must NOT do**:
  - Do NOT cast to `any` anywhere in the fix
  - Do NOT add a settings field to replicate this — just use the public API
  - Do NOT remove the `attachmentFolderOverride` settings functionality

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-line fix with fallback
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-5, 7)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `src/main.ts:66` — Current `(this.app.vault as any).getConfig?.('attachmentFolderPath')`
  - `src/main.ts:61-77` — `getAttachmentFolder()` function where fix goes
  - Obsidian API: `App.getConfig(key: string): string | null` — public since 1.4

  **Acceptance Criteria**:
  - [ ] No `as any` in `src/main.ts` (search confirms)
  - [ ] `this.app.getConfig('attachmentFolderPath')` is used instead
  - [ ] Fallback handles `undefined`/`null` return
  - [ ] TypeScript compiles: `npx tsc --noEmit`

  **QA Scenarios**:
  ```
  Scenario: Attachment folder resolved correctly
    Tool: Bash (node - unit test or manual)
    Preconditions: Plugin loaded in Obsidian
    Steps:
      1. Set attachment folder in Obsidian settings to "_resources"
      2. Run download command on a note with #json= link
      3. Verify file is saved to "_resources/" folder
    Expected Result: File saved to correct attachment folder
    Failure Indicators: File saved to vault root, or error about getConfig
    Evidence: .omo/evidence/task-6-getConfig.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix(plugin): use public App.getConfig instead of vault as any`
  - Files: `src/main.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 7. Fix Notice lifecycle (duration 0 leak)

  **What to do**:
  - In `src/main.ts`, change `new Notice(msg, 0)` to `new Notice(msg, 10_000)` (10 seconds auto-dismiss)
  - Ensure `progressNotice?.hide()` is still called on completion
  - Add try/finally around the download processing loop so `progressNotice?.hide()` is always called even on unexpected errors
  - In error catch blocks, ensure Notice is also hidden before showing error Notice

  **Must NOT do**:
  - Do NOT remove the progress notice entirely — users need feedback
  - Do NOT increase duration beyond 10 seconds for progress messages
  - Do NOT change error notice durations (those are already default ~5s)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small fix to one file, straightforward lifecycle change
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-6)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `src/main.ts:98` — `new Notice(msg, 0)` duration-0 notice
  - `src/main.ts:100-110` — Download processing loop and notice lifecycle

  **Acceptance Criteria**:
  - [ ] No `new Notice(msg, 0)` in codebase
  - [ ] Progress notice has auto-dismiss duration (10 seconds)
  - [ ] `progressNotice?.hide()` called in try/finally block
  - [ ] Notice hidden before showing error notices

  **QA Scenarios**:
  ```
  Scenario: Notice auto-dismisses after duration
    Tool: Bash
    Preconditions: Plugin loaded in Obsidian
    Steps:
      1. Download a valid #json= link
      2. Observe progress notice appears then auto-dismisses within 10s
      3. Verify no lingering notices on screen
    Expected Result: Notice appears during download, disappears after completion or 10s
    Failure Indicators: Notice persists indefinitely after download completes
    Evidence: .omo/evidence/task-7-notice-lifecycle.txt
  ```

  **Commit**: YES (groups with Wave 1)
  - Message: `fix(plugin): auto-dismiss progress notice, ensure cleanup on error`
  - Files: `src/main.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 8. Spanish→English translation with backward-compatible migration

  **What to do**:
  Plugin strings to translate:
  - `src/injector.ts:59` — `ANNOTATION_EMPTY_CANVAS`: Change from `(⚠ canvas vacío — pedile al autor que abra la sala cuando corras el downloader)` to `(⚠ empty canvas — ask the author to open the room, then re-run the downloader)`
  - `src/downloader.ts:74` — `notify?.('Excalidraw: sala vacía — abriendo browser, reconectando en 10s…')` → `notify?.('Excalidraw: empty room — opening browser, reconnecting in 10s…')`
  - `src/downloader.ts:83` — `notify?.('Excalidraw: intento ${attempt} — reconectando en 3s…')` → `notify?.('Excalidraw: attempt ${attempt} — reconnecting in 3s…')`
  - `src/downloader.ts:96` — `notify?.('Excalidraw: descargando escena…')` → `notify?.('Excalidraw: downloading scene…')`
  - `src/room-client.ts:154` — Error message with Spanish text → English equivalent

  Migration logic for backward compatibility:
  - In `src/injector.ts`, update `hasExistingLocalCopy()` and `injectOne()` to recognize BOTH old Spanish annotations and new English annotations
  - Add detection of old Spanish pattern: `canvas vacío` alongside new `empty canvas`
  - When a Spanish annotation is found during re-run, replace it with the English version (migration)
  - In `src/link.ts`, ensure `parseAll` idempotency check works with both Spanish and English annotations

  **Must NOT do**:
  - Do NOT add an i18n system — just replace strings
  - Do NOT change the annotation format `([[...|local copy DATE]])` — only the error strings inside
  - Do NOT forget to handle migration — existing users have Spanish annotations in their notes

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires understanding idempotency + migration logic across multiple files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (depends on Wave 1 abort signal for retries)
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 1, 2

  **References**:
  - `src/injector.ts:57-59` — Constants with Spanish strings
  - `src/injector.ts:65-71` — `hasExistingLocalCopy` that needs migration awareness
  - `src/injector.ts:96-183` — `injectOne` state machine that handles annotations
  - `src/downloader.ts:74,83,96` — Spanish notice messages
  - `src/room-client.ts:154` — Spanish error message

  **Acceptance Criteria**:
  - [ ] All Spanish strings in user-facing output replaced with English
  - [ ] `hasExistingLocalCopy` recognizes old Spanish annotation pattern
  - [ ] `injectOne` migrates Spanish annotations to English on re-run
  - [ ] No Spanish strings when searching: `grep -r "vacío\|sala\|intento\|pedile" src/` returns empty
  - [ ] TypeScript compiles: `npx tsc --noEmit`

  **QA Scenarios**:
  ```
  Scenario: New annotation is in English
    Tool: Bash
    Preconditions: Plugin loaded, note with #json= link for empty room
    Steps:
      1. Download an empty room
      2. Check annotation injected next to link
    Expected Result: Annotation contains "empty canvas" (English), not "canvas vacío"
    Failure Indicators: Annotation contains Spanish text
    Evidence: .omo/evidence/task-8-english-annotation.txt

  Scenario: Old Spanish annotation is migrated
    Tool: Bash
    Preconditions: Note with old Spanish annotation "(⚠ canvas vacío — pedile al autor...)"
    Steps:
      1. Run download/refresh on the note
      2. Check that old Spanish annotation is replaced with English
    Expected Result: Spanish annotation replaced with "empty canvas" English version
    Failure Indicators: Old Spanish annotation remains, or duplicate annotation appears
    Evidence: .omo/evidence/task-8-migration.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `fix: translate Spanish strings to English with backward-compatible migration`
  - Files: `src/injector.ts`, `src/downloader.ts`, `src/room-client.ts`
  - Pre-commit: `npx tsc --noEmit && grep -r "vacío\|sala\|intento\|pedile" src/`

- [x] 9. Fix window.open popup blocker issue

  **What to do**:
  - In `src/downloader.ts:75`, replace `window.open(link.url)` with a user-facing Notice that includes the URL:
    ```typescript
    new Notice('Excalidraw: Opening browser to restore room data. If blocked by popup blocker, open this URL manually: ' + link.url, 15000);
    window.open(link.url); // Attempt to open, may be blocked
    ```
  - This way, if the popup is blocked, the user still has the URL in a 15-second Notice to copy manually
  - Do NOT remove the `window.open` call — it still works in Electron most of the time

  **Must NOT do**:
  - Do NOT switch to `require('electron').shell.openExternal` — that's more invasive
  - Do NOT make the Notice dismissable-only — let it auto-dismiss after 15s
  - Do NOT add clipboard operations

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Small change but needs understanding of Electron popup behavior
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (after Wave 1)
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: Task 1 (for abort signal integration context)

  **References**:
  - `src/downloader.ts:75` — Current `window.open(link.url)` call
  - `src/downloader.ts:70-77` — Room download retry block where window.open is called

  **Acceptance Criteria**:
  - [ ] `window.open` call is still present
  - [ ] Notice shown with URL before window.open call
  - [ ] Notice duration is 15 seconds
  - [ ] TypeScript compiles

  **QA Scenarios**:
  ```
  Scenario: Popup blocked notice appears
    Tool: Bash (visual test in Obsidian)
    Steps:
      1. Open note with #room= link for an empty room
      2. Run download command
      3. Observe Notice with URL
    Expected Result: Notice appears with "Opening browser..." + URL, stays for 15 seconds
    Failure Indicators: No notice appears, or notice disappears too quickly
    Evidence: .omo/evidence/task-9-popup-notice.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `fix(plugin): show URL in Notice before window.open for popup blocker fallback`
  - Files: `src/downloader.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 10. Surface errors via Notice (not just console)

  **What to do**:
  - In `src/main.ts`, change `console.warn` calls to also show user-facing Notices:
    - Line 115: `console.warn(...)` → add `new Notice('Excalidraw: refresh failed for link — see console for details', 5000)`
    - Line 119: `console.error(...)` → add `new Notice('Excalidraw: download failed for link — see console for details', 5000)`
  - Keep the `console.warn/error` calls for debugging — just add Notices on top

  **Must NOT do**:
  - Do NOT remove console.warn/error — they're useful for debugging
  - Do NOT show full error details in Notices — just human-readable summary + "see console"
  - Do NOT add a logging framework

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two-line change, straightforward
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `src/main.ts:115` — `console.warn` for refresh failure
  - `src/main.ts:119` — `console.error` for download failure

  **Acceptance Criteria**:
  - [ ] Both error paths show a Notice to the user
  - [ ] Both Notice messages are in English
  - [ ] console.warn/error calls are preserved
  - [ ] TypeScript compiles

  **QA Scenarios**:
  ```
  Scenario: Failed download shows user-visible Notice
    Tool: Bash (visual in Obsidian)
    Steps:
      1. Create note with invalid #json= link
      2. Run download command
      3. Observe Notice
    Expected Result: User-visible Notice appears saying "download failed"
    Failure Indicators: No Notice appears (only console error)
    Evidence: .omo/evidence/task-10-error-notice.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `fix(plugin): show user-facing Notice for download failures alongside console error`
  - Files: `src/main.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 11. Fix unsafe .buffer as ArrayBuffer in writeBinary

  **What to do**:
  - In `src/downloader.ts`, replace `data.buffer as ArrayBuffer` with a safe slice:
    ```typescript
    await app.vault.adapter.writeBinary(
      destPath,
      data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength)
    );
    ```
  - `Uint8Array.buffer` may be larger than the actual data (if it's a view into a larger buffer). Slicing ensures we only write the actual bytes.

  **Must NOT do**:
  - Do NOT change the file writing approach
  - Do NOT add a new utility function — inline the slice is sufficient

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-line fix, well-understood issue
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `src/downloader.ts:48` — Current `data.buffer as ArrayBuffer`

  **Acceptance Criteria**:
  - [ ] `data.buffer as ArrayBuffer` replaced with `data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength)`
  - [ ] No `as ArrayBuffer` unsafe casts in codebase
  - [ ] TypeScript compiles

  **QA Scenarios**:
  ```
  Scenario: Downloaded file has correct size
    Tool: Bash
    Preconditions: Valid #json= link
    Steps:
      1. Download a valid Excalidraw scene
      2. Check file size matches expected (not padded)
    Expected Result: File size exactly matches decrypted data length
    Failure Indicators: File larger than expected (includes padding bytes)
    Evidence: .omo/evidence/task-11-arraybuffer.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `fix(plugin): safely slice Uint8Array buffer to avoid writing extra bytes`
  - Files: `src/downloader.ts`
  - Pre-commit: `npx tsc --noEmit`

- [x] 12. Add versions.json for Obsidian submission

  **What to do**:
  - Create `versions.json` in repository root with format:
    ```json
    {
      "1.0.0": "1.4.0"
    }
    ```
  - This maps plugin version to minimum Obsidian version, matching `manifest.json`'s `minAppVersion`

  **Must NOT do**:
  - Do NOT add versions beyond 1.0.0 — this is the initial release
  - Do NOT change manifest.json version

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Single-file creation, simple format
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `manifest.json` — Current version `1.0.0`, minAppVersion `1.4.0`
  - Obsidian submission requirements: versions.json format

  **Acceptance Criteria**:
  - [ ] `versions.json` exists in repository root
  - [ ] Contains `{"1.0.0": "1.4.0"}`
  - [ ] Version matches `manifest.json` version field

  **QA Scenarios**:
  ```
  Scenario: versions.json is valid
    Tool: Bash
    Steps:
      1. cat versions.json
      2. jq '.' versions.json
      3. Verify manifest.json version matches
    Expected Result: Valid JSON with correct version mapping
    Failure Indicators: Missing file, invalid JSON, version mismatch
    Evidence: .omo/evidence/task-12-versions-json.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `chore: add versions.json for Obsidian community plugin submission`
  - Files: `versions.json`
  - Pre-commit: None

- [x] 13. Fix Go CLI: refresh failure annotation + go.mod

  **What to do**:
  - In `cli/internal/injector/injector.go`, add `ANNOTATION_REFRESH_FAILED = "(⚠ refresh failed)"` constant
  - Update `processFile` (or `injectOne` equivalent) to handle refresh failure: when a refresh download fails but a local copy exists, keep the old `([[...]])` annotation AND append `(⚠ refresh failed)` — matching the TS plugin behavior
  - Currently the Go code silently keeps the old annotation with no error marker — this diverges from the TypeScript plugin
  - In `cli/go.mod`, run `go mod tidy` equivalent: remove `// indirect` from `gorilla/websocket` line since it's a direct dependency

  **Must NOT do**:
  - Do NOT change the annotation format — use same format as TS plugin
  - Do NOT add new error types — reuse existing error handling

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Two small targeted changes in Go files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 14
  - **Blocked By**: None

  **References**:
  - `cli/internal/injector/injector.go:107-111` — Current silent keep of old annotation on refresh failure
  - `cli/go.mod:5` — `// indirect` annotation on gorilla/websocket
  - `src/injector.ts:125-141` — TS plugin refresh failure handling (reference implementation)

  **Acceptance Criteria**:
  - [ ] `ANNOTATION_REFRESH_FAILED` constant exists in Go injector
  - [ ] Refresh failure appends "(⚠ refresh failed)" to existing annotation
  - [ ] `go.mod` no longer has `// indirect` on gorilla/websocket
  - [ ] `cd cli && go build ./...` succeeds
  - [ ] `cd cli && go mod tidy` shows no changes needed

  **QA Scenarios**:
  ```
  Scenario: Refresh failure shows annotation in Go CLI
    Tool: Bash
    Preconditions: Go CLI built
    Steps:
      1. Create a vault note with an already-downloaded excalidraw link
      2. Run CLI with --refresh on a link that fails to download
      3. Check note content for "(⚠ refresh failed)" annotation
    Expected Result: Annotation contains "(⚠ refresh failed)" appended to existing link
    Failure Indicators: No annotation, or old annotation without error marker
    Evidence: .omo/evidence/task-13-go-refresh.txt

  Scenario: go.mod is correct
    Tool: Bash
    Steps:
      1. cd cli && go mod tidy
      2. git diff go.mod
    Expected Result: No changes needed, gorilla/websocket listed as direct dependency
    Failure Indicators: go.mod changes after tidy
    Evidence: .omo/evidence/task-13-go-mod.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `fix(cli): add refresh failure annotation, fix go.mod indirect annotation`
  - Files: `cli/internal/injector/injector.go`, `cli/go.mod`
  - Pre-commit: `cd cli && go build ./... && go mod tidy`

- [x] 14. Fix Go CLI: Spanish→English prompts and annotations

  **What to do**:
  - In `cli/internal/injector/injector.go`, change `annotationEmptyCanvas` from Spanish to English:
    - `(⚠ canvas vacío — pedile al autor que abra la sala cuando corras el downloader)` → `(⚠ empty canvas — ask the author to open the room, then re-run the downloader)`
  - In `cli/main.go`, translate `promptAttachmentConfig` from Spanish to English:
    - "Selecciona el modo de carpeta de attachments" → "Select attachment folder mode"
    - "Carpeta fija" → "Fixed folder"
    - "Misma carpeta que la nota" → "Same folder as note"
    - "Subcarpeta de la nota" → "Subfolder of note"
    - "Raíz del vault" → "Vault root"
    - "Selección [1-4]:" → "Selection [1-4]:"
    - "Opción inválida, usando carpeta fija '_resources'." → "Invalid selection, using fixed folder '_resources'."
  - In `cli/internal/excalidraw/room_client.go`, translate any remaining Spanish strings
  - Add backward-compatible migration: recognize old Spanish annotations in `hasExistingLocalCopy` / `injectOne` equivalent in Go

  **Must NOT do**:
  - Do NOT add i18n system — just replace strings
  - Do NOT change CLI flags or command structure
  - Do NOT forget migration for existing annotations

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: String replacement across a few files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (after Task 13)
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: Task 13

  **References**:
  - `cli/internal/injector/injector.go` — Spanish annotationEmptyCanvas constant
  - `cli/main.go:206-248` — Spanish promptAttachmentConfig
  - `cli/internal/excalidraw/room_client.go` — Any remaining Spanish strings

  **Acceptance Criteria**:
  - [ ] No Spanish strings in Go CLI output: `grep -r "vacío\|sala\|intento\|pedile\|Selecciona\|Opción\|Carpeta\|Raíz" cli/` returns empty
  - [ ] Go CLI compiles: `cd cli && go build ./...`
  - [ ] Old Spanish annotations recognized and migrated on re-run

  **QA Scenarios**:
  ```
  Scenario: CLI prompts in English
    Tool: Bash
    Preconditions: Go CLI built
    Steps:
      1. Run CLI without attachment config
      2. Check prompt text
    Expected Result: Prompts in English ("Select attachment folder mode")
    Failure Indicators: Spanish text in prompts
    Evidence: .omo/evidence/task-14-cli-english.txt

  Scenario: Old Spanish annotations migrated
    Tool: Bash
    Preconditions: Vault note with Spanish annotation
    Steps:
      1. Run CLI with --refresh on note with old Spanish annotation
      2. Check annotation is replaced with English
    Expected Result: Spanish replaced with English
    Failure Indicators: Old Spanish annotation remains or duplicates
    Evidence: .omo/evidence/task-14-cli-migration.txt
  ```

  **Commit**: YES (groups with Wave 2)
  - Message: `fix(cli): translate Spanish strings to English with migration support`
  - Files: `cli/internal/injector/injector.go`, `cli/main.go`, `cli/internal/excalidraw/room_client.go`
  - Pre-commit: `cd cli && go build ./...`

- [x] 15. Fix README annotation format + error state docs

  **What to do**:
  - In `README.md`, fix the annotation example on line 57 from markdown link format:
    `([local copy 18/05/2026](_resources/excalidraw-abc123.excalidraw))`
    to wikilink format:
    `([[excalidraw-abc123.excalidraw|local copy 18/05/2026]])`
  - Add a **Desktop Only** notice at the top: "⚠ This plugin is desktop-only — it uses WebSocket connections and `window.open()` which are not available on mobile."
  - Update the **Error handling** section to document ALL three annotation types:
    1. `([[excalidraw-ID.excalidraw|local copy DD/MM/YYYY]])` — successful download
    2. `(⚠ download failed)` — download failed, no local copy
    3. `(⚠ refresh failed)` — refresh failed, old local copy preserved (currently NOT documented)
    4. `(⚠ empty canvas — ask the author to open the room, then re-run the downloader)` — room was empty (currently in Spanish only, must be English after Task 8)
  - Add a **Known Limitations** section:
    - Desktop only (Electron required)
    - Room downloads require opening a browser tab
    - Rate limits may apply from Excalidraw's servers
  - Update **Community Plugins** installation section to clarify: "Once approved, you can install from Community Plugins. Until then, use BRAT or manual installation."

  **Must NOT do**:
  - Do NOT change the actual annotation format in code — only fix the README examples
  - Do NOT add screenshots (that's Task 16)
  - Do NOT add a CHANGELOG (can be a separate task)

  **Recommended Agent Profile**:
  - **Category**: `writing`
    - Reason: Documentation-focused task requiring clear technical writing
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (needs Task 8 results for accurate annotation text)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 20
  - **Blocked By**: Task 8 (Spanish→English translation must be done first so README matches new strings)

  **References**:
  - `README.md:57` — Wrong annotation format example
  - `README.md:62-71` — Error handling section missing 2 annotation types
  - `src/injector.ts:57-59` — Actual annotation constants (updated English versions after Task 8)
  - `manifest.json:9` — `"isDesktopOnly": true` needs to be documented

  **Acceptance Criteria**:
  - [ ] README annotation example shows `([[excalidraw-abc123.excalidraw|local copy 18/05/2026]])` (wikilink format)
  - [ ] Desktop-only notice present in README
  - [ ] All 4 error annotation types documented
  - [ ] Error descriptions are in English
  - [ ] Community plugins installation section clarifies pre-approval status

  **QA Scenarios**:
  ```
  Scenario: README annotation format matches code
    Tool: Bash
    Steps:
      1. grep -n "local copy" README.md
      2. Verify format uses wikilink `([[...|...]])` not markdown `([...](...))`
    Expected Result: All annotation examples use wikilink format
    Failure Indicators: Any markdown link format for annotations
    Evidence: .omo/evidence/task-15-readme-annotations.txt

  Scenario: All error states documented
    Tool: Bash
    Steps:
      1. grep -c "empty canvas\|refresh failed\|download failed" README.md
    Expected Result: Count >= 3 (all three error states mentioned)
    Failure Indicators: Missing error state documentation
    Evidence: .omo/evidence/task-15-error-states.txt
  ```

  **Commit**: YES (groups with Wave 3 — docs)
  - Message: `docs: fix README annotation format, add error states, desktop-only notice`
  - Files: `README.md`
  - Pre-commit: None

- [x] 16. Add screenshots/demo to README

  **What to do**:
  - Create a screenshot or GIF showing the before/after transformation:
    - Before: A note with `https://excalidraw.com/#json=abc123,key123`
    - After: The same note with `([[excalidraw-abc123.excalidraw|local copy 18/05/2026]])` annotation
  - Create a screenshot of the settings tab (attachment folder override)
  - Add these images to a `screenshots/` directory in the repo
  - Update README.md to include the images with relative links
  - Place the before/after image near the top of the README (after the description, before installation)

  **Must NOT do**:
  - Do NOT use external image hosting — images must be in the repo
  - Do NOT use large GIFs (>2MB) — keep it lightweight for the community store
  - Do NOT show real user data or vault contents in screenshots

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Requires creating visual assets and embedding in README
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: Task 15 (README structure should be settled)

  **References**:
  - `README.md` — Current README structure
  - `src/main.ts` — Plugin commands and settings for screenshot context
  - Obsidian community plugin listing — README is rendered in the store

  **Acceptance Criteria**:
  - [ ] `screenshots/` directory exists with at least 2 images
  - [ ] README includes image links using relative paths
  - [ ] Before/after image shows the annotation transformation
  - [ ] Settings screenshot shows the attachment folder override

  **QA Scenarios**:
  ```
  Scenario: Screenshots exist and render in README
    Tool: Bash
    Steps:
      1. ls screenshots/
      2. grep -n "screenshots/" README.md
      3. Verify image files are <2MB each
    Expected Result: At least 2 images in screenshots/, README links to them
    Failure Indicators: Missing screenshots directory, no image links in README
    Evidence: .omo/evidence/task-16-screenshots.txt
  ```

  **Commit**: YES (groups with Wave 3 — docs)
  - Message: `docs: add before/after and settings screenshots to README`
  - Files: `README.md`, `screenshots/`
  - Pre-commit: None

- [x] 17. Add Go CLI --yes flag + closeBrowserTab warning

  **What to do**:
  - In `cli/main.go`, add `--yes` / `-y` flag that skips the confirmation prompt:
    ```go
    yesFlag := flag.Bool("yes", false, "Skip confirmation prompt (for scripting)")
    // or shorter: flag.BoolVarP(&yes, "yes", "y", false, "...")
    ```
  - When `--yes` is set, skip `promptAttachmentConfig` if attachment folder is already configured, or use default `_resources`
  - In `cli/internal/excalidraw/room_client.go`, add a log message when `closeBrowserTab` is a no-op (Linux/Windows):
    ```go
    log.Println("Note: Browser tab could not be closed automatically. Please close it manually.")
    ```
  - Fix `openInBrowser` on Windows to add empty title argument: `exec.Command("cmd", "/c", "start", "", url)`

  **Must NOT do**:
  - Do NOT change the default behavior — `--yes` must be opt-in
  - Do NOT remove the confirmation prompt (it's useful for interactive use)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small targeted changes across 2 Go files
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `cli/main.go` — Current flag parsing and confirmation prompt
  - `cli/internal/excalidraw/room_client.go:155-201` — `closeBrowserTab` function
  - `cli/internal/excalidraw/room_client.go:145` — `openInBrowser` with Windows `start` command

  **Acceptance Criteria**:
  - [ ] `--yes` / `-y` flag exists and skips confirmation
  - [ ] Default behavior unchanged (confirmation still required without flag)
  - [ ] Linux/Windows logs message about unclosed tab
  - [ ] Windows `start` command includes empty title argument
  - [ ] `cd cli && go build ./...` succeeds

  **QA Scenarios**:
  ```
  Scenario: --yes flag skips confirmation
    Tool: Bash
    Preconditions: Go CLI built
    Steps:
      1. Run CLI with --yes flag
      2. Verify no interactive prompt appears
    Expected Result: No confirmation prompt, processing starts immediately
    Failure Indicators: Prompt still appears
    Evidence: .omo/evidence/task-17-yes-flag.txt

  Scenario: Without --yes flag still prompts
    Tool: Bash
    Steps:
      1. Run CLI without --yes flag
      2. Verify confirmation prompt appears
    Expected Result: Prompt appears asking for confirmation
    Failure Indicators: No prompt, processing starts immediately
    Evidence: .omo/evidence/task-17-no-yes.txt
  ```

  **Commit**: YES (groups with Wave 3)
  - Message: `feat(cli): add --yes flag for non-interactive mode, warn on unclosed browser tab`
  - Files: `cli/main.go`, `cli/internal/excalidraw/room_client.go`
  - Pre-commit: `cd cli && go build ./...`

- [x] 18. Set up Vitest for plugin + write core tests

  **What to do**:
  - Install Vitest: `npm install -D vitest`
  - Add `test` script to `package.json`: `"test": "vitest run"`
  - Create `src/__tests__/` directory
  - Write tests for the most critical pure functions:

  **Test file 1: `src/__tests__/link.test.ts`**
  - `parseAll()` with valid `#json=` URL → extracts kind, id, key
  - `parseAll()` with valid `#room=` URL → extracts kind, id, key
  - `parseAll()` with no URL → returns empty array
  - `parseAll()` with multiple URLs → deduplicates (seen map)
  - `parseAll()` with already-annotated URL (has `([[` prefix) → marks as `refresh: true`
  - `parseAll()` with URL containing special characters in key (`_`, `-`) → extracts correctly
  - `parseAll()` edge case: empty content → returns empty array

  **Test file 2: `src/__tests__/crypto.test.ts`**
  - `decrypt()` with valid ciphertext → returns decrypted Uint8Array
  - `decrypt()` with too-short input (< 28 bytes) → throws error
  - `decryptGCM()` with valid separate IV + ciphertext → returns decrypted data
  - `decrypt()` with invalid key → throws error

  **Test file 3: `src/__tests__/injector.test.ts`**
  - `injectOne()` with fresh URL → returns content with `([[excalidraw-123.excalidraw|local copy DATE]])` annotation
  - `injectOne()` with already-annotated URL → no duplicate annotation
  - `injectOne()` with failed download → `(⚠ download failed)` annotation
  - `injectOne()` with refresh failure + existing local copy → `(⚠ refresh failed)` annotation
  - `hasExistingLocalCopy()` returns true for annotated links
  - `hasExistingLocalCopy()` returns false for unprocessed links
  - Migration: old Spanish `canvas vacío` annotation recognized

  **Must NOT do**:
  - Do NOT test WebSocket interactions — too complex to mock
  - Do NOT add integration tests with real Excalidraw API
  - Do NOT mock Obsidian's `App` unless necessary for injector tests
  - Do NOT aim for 100% coverage — focus on critical paths

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Setting up test infrastructure + writing meaningful tests for state machine logic
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (after Wave 1 completes)
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 20
  - **Blocked By**: Tasks 1, 2, 4, 11 (bugs that affect test expectations)

  **References**:
  - `src/link.ts` — Pure function `parseAll`, easy to test
  - `src/crypto.ts` — `decrypt` and `decryptGCM` functions
  - `src/injector.ts:44-183` — `injectOne` state machine (most complex, most important to test)
  - `src/injector.ts:57-59` — Annotation constants (will be English after Task 8)
  - `vitest` docs for configuration

  **Acceptance Criteria**:
  - [ ] `vitest` installed as devDependency
  - [ ] `"test": "vitest run"` script in package.json
  - [ ] `src/__tests__/link.test.ts` exists with 6+ test cases
  - [ ] `src/__tests__/crypto.test.ts` exists with 4+ test cases
  - [ ] `src/__tests__/injector.test.ts` exists with 7+ test cases
  - [ ] All tests pass: `npx vitest run`
  - [ ] No `any` types or `@ts-ignore` in test files

  **QA Scenarios**:
  ```
  Scenario: All tests pass
    Tool: Bash
    Preconditions: vitest installed, test files written
    Steps:
      1. npx vitest run
      2. Check exit code
    Expected Result: All tests pass, exit code 0
    Failure Indicators: Any test failures, import errors
    Evidence: .omo/evidence/task-18-vitest.txt

  Scenario: Test coverage of critical paths
    Tool: Bash
    Steps:
      1. npx vitest run --coverage
      2. Check coverage for link.ts, crypto.ts, injector.ts
    Expected Result: Each tested module has >70% statement coverage
    Failure Indicators: Any module below 50% coverage
    Evidence: .omo/evidence/task-18-coverage.txt
  ```

  **Commit**: YES (groups with Wave 3 — tests)
  - Message: `test: add Vitest infrastructure and core module tests`
  - Files: `package.json`, `vitest.config.ts` (new), `src/__tests__/link.test.ts` (new), `src/__tests__/crypto.test.ts` (new), `src/__tests__/injector.test.ts` (new)
  - Pre-commit: `npx vitest run`

- [x] 19. Write Go unit tests for core modules

  **What to do**:
  - Create test files for the most critical Go packages:

  **Test file 1: `cli/internal/excalidraw/link_test.go`**
  - Test `ParseAll` with valid `#json=` URL
  - Test `ParseAll` with valid `#room=` URL
  - Test `ParseAll` with no URLs → empty slice
  - Test `ParseAll` deduplicates seen URLs
  - Test `ParseAll` with already-annotated URL → marks as refresh
  - Test `ParseAll` edge case: empty content

  **Test file 2: `cli/internal/excalidraw/crypto_test.go`**
  - Test `Decrypt` with valid ciphertext
  - Test `Decrypt` with too-short input
  - Test `DecryptGCM` with valid separate IV + ciphertext
  - Test `Decrypt` with invalid key

  **Test file 3: `cli/internal/injector/injector_test.go`**
  - Test `injectOne` fresh URL → annotation added
  - Test `injectOne` already-annotated URL → no duplicate
  - Test `injectOne` with download failure → `(⚠ download failed)`
  - Test `injectOne` with refresh failure + existing copy → `(⚠ refresh failed)`
  - Test `hasExistingLocalCopy` true for annotated links
  - Test `hasExistingLocalCopy` false for unprocessed links
  - Test migration: old Spanish `canvas vacío` annotation recognized

  **Test file 4: `cli/internal/vault/walker_test.go`**
  - Test `Walk` finds markdown files
  - Test `Walk` skips hidden directories (`.`)
  - Test `Walk` skips directories in skip list
  - Test `Walk` with empty directory → empty result

  **Must NOT do**:
  - Do NOT write integration tests requiring real Excalidraw server
  - Do NOT mock vault operations — use temp directories instead
  - Do NOT aim for 100% coverage — focus on state machine and parsing

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Multiple test files, includes state machine testing for injector
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 20
  - **Blocked By**: Task 13 (Go refresh failure annotation must be updated first)

  **References**:
  - `cli/internal/excalidraw/link.go` — ParseAll function
  - `cli/internal/excalidraw/crypto.go` — Decrypt and DecryptGCM
  - `cli/internal/injector/injector.go` — injectOne state machine
  - `cli/internal/vault/walker.go` — Walk function
  - Go testing patterns: table-driven tests, `testing.T`, `t.Errorf`

  **Acceptance Criteria**:
  - [ ] `cli/internal/excalidraw/link_test.go` exists with 6+ test cases
  - [ ] `cli/internal/excalidraw/crypto_test.go` exists with 4+ test cases
  - [ ] `cli/internal/injector/injector_test.go` exists with 7+ test cases
  - [ ] `cli/internal/vault/walker_test.go` exists with 4+ test cases
  - [ ] All tests pass: `cd cli && go test ./...`
  - [ ] `go vet ./...` passes with no warnings

  **QA Scenarios**:
  ```
  Scenario: All Go tests pass
    Tool: Bash
    Steps:
      1. cd cli && go test ./...
      2. Check exit code
    Expected Result: All tests pass, exit code 0
    Failure Indicators: Any test failures, compilation errors
    Evidence: .omo/evidence/task-19-go-tests.txt

  Scenario: Go vet passes
    Tool: Bash
    Steps:
      1. cd cli && go vet ./...
    Expected Result: No warnings or errors
    Failure Indicators: Any vet warnings
    Evidence: .omo/evidence/task-19-go-vet.txt
  ```

  **Commit**: YES (groups with Wave 3 — tests)
  - Message: `test(cli): add unit tests for link, crypto, injector, and vault packages`
  - Files: `cli/internal/excalidraw/link_test.go` (new), `cli/internal/excalidraw/crypto_test.go` (new), `cli/internal/injector/injector_test.go` (new), `cli/internal/vault/walker_test.go` (new)
  - Pre-commit: `cd cli && go test ./... && go vet ./...`

- [x] 20. Final Obsidian submission compliance check

  **What to do**:
  - Verify manifest.json compliance:
    - `id` doesn't contain "obsidian" or "plugin" → currently `excalidraw-link-downloader` ✅
    - `name` is filled → check
    - `version` follows semver → `1.0.0` ✅
    - `minAppVersion` matches → `1.4.0` ✅
    - `description` ≤ 250 chars and ends with `.` → verify
    - `author` is filled → verify
    - `isDesktopOnly` is `true` → currently ✅
    - Remove `fundingUrl` if not needed → currently not present ✅
  - Verify no sample code remnants:
    - No `MyPlugin` class names
    - No `MyPluginSettings` class names
    - No `SampleSettingTab` class names
    - No `console.log` in production code
  - Verify `normalizePath()` or equivalent is used for all user-defined paths (checking attachment folder)
  - Verify no `innerHTML`/`outerHTML` with user input
  - Verify all event listeners use `registerEvent()` or have `onunload()` cleanup (added in Task 1)
  - Verify no default hotkeys
  - Verify `main.js` builds successfully: `npm run build`
  - Verify `versions.json` exists with correct format
  - Verify `LICENSE` file exists and is filled out (not templated)
  - Verify README exists and describes purpose/usage

  **Must NOT do**:
  - Do NOT submit to Obsidian community plugins — just verify compliance
  - Do NOT change version numbers (that's for the actual release)
  - Do NOT add CI/CD pipeline

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Verification checklist, not implementation
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (last task, after all others)
  - **Blocks**: Final verification wave
  - **Blocked By**: Tasks 15, 18, 19

  **References**:
  - `manifest.json` — Plugin manifest
  - `README.md` — Documentation
  - `LICENSE` — License file
  - `versions.json` — Version mapping (created in Task 12)
  - `src/main.ts` — Plugin entry (for registerEvent check)
  - Obsidian community plugin submission requirements

  **Acceptance Criteria**:
  - [ ] `manifest.json` all fields valid: id, name, version, minAppVersion, description ≤250 chars ending with `.`, author, isDesktopOnly
  - [ ] No `MyPlugin`, `SampleSettingTab`, `MyPluginSettings` in source
  - [ ] No `console.log` in production code (only `console.warn`/`console.error`)
  - [ ] All user-defined paths use `normalizePath()` or similar
  - [ ] No `innerHTML`/`outerHTML` with user input
  - [ ] Event listeners cleaned up in `onunload()` (Task 1)
  - [ ] `npm run build` produces `main.js`
  - [ ] `versions.json` exists and is valid JSON
  - [ ] `LICENSE` file exists with actual author info
  - [ ] README exists with usage instructions

  **QA Scenarios**:
  ```
  Scenario: manifest.json is valid
    Tool: Bash
    Steps:
      1. cat manifest.json | jq '.id' | test "obsidian" not in value
      2. cat manifest.json | jq '.description' | test length ≤ 250, ends with "."
      3. cat manifest.json | jq '.version' | test matches semver
    Expected Result: All manifest checks pass
    Failure Indicators: Invalid id, description, or version
    Evidence: .omo/evidence/task-20-manifest.txt

  Scenario: No sample code remnants
    Tool: Bash
    Steps:
      1. grep -r "MyPlugin\|SampleSettingTab\|MyPluginSettings" src/
    Expected Result: No matches
    Failure Indicators: Sample code found
    Evidence: .omo/evidence/task-20-no-sample.txt

  Scenario: Build produces main.js
    Tool: Bash
    Steps:
      1. npm run build
      2. test -f main.js
    Expected Result: Build succeeds, main.js exists
    Failure Indicators: Build errors or missing main.js
    Evidence: .omo/evidence/task-20-build.txt
  ```

  **Commit**: YES (groups with Wave 3)
  - Message: `chore: verify Obsidian community plugin submission compliance`
  - Files: Manifest verification only, minor fixes if needed
  - Pre-commit: `npm run build`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .omo/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `tsc --noEmit` on plugin, `go vet ./...` on CLI. Review all changed files for: `as any`/`@ts-ignore`, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names. Run all tests: `npx vitest run` and `go test ./...`.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Build plugin (`npm run build`), load in Obsidian test vault. Run download command on a note with a `#json=` link. Verify: file appears in attachment folder, annotation appears next to link, Notice shows completion. Test refresh command. Test room download (requires real Excalidraw room). Verify no Spanish strings in UI output. Verify Go CLI compiles and runs. Save to `.omo/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT Have" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Wave 1** (all fixes): `fix(plugin): add lifecycle cleanup, cap retries, validate inputs, pin deps`
- **Wave 2** (translations + CLI): `fix: translate Spanish strings to English with migration, align CLI behavior`
- **Wave 3** (docs + tests): `docs: fix README and add screenshots, test: add unit tests for core modules`
- **Final**: `chore: add versions.json, verify submission compliance`

---

## Success Criteria

### Verification Commands
```bash
npm run build                                  # Expected: clean build
cd cli && go build ./...                      # Expected: clean build
cd cli && go test ./...                       # Expected: all tests pass
npx vitest run                                # Expected: all tests pass
grep -r "vacío\|sala\|intento\|pedile" src/   # Expected: no matches (no Spanish)
cat manifest.json | jq '.description' | tail -c 2  # Expected: "."
test -f versions.json && echo "OK"            # Expected: OK
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass (TypeScript + Go)
- [ ] Plugin builds without errors
- [ ] No Spanish strings in user-facing output
- [ ] README annotation format matches code output
- [ ] versions.json exists with correct format
- [ ] manifest.json description ends with period, ≤250 chars
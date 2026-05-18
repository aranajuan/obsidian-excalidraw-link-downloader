# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
cd cli && go build ./...      # compile Go CLI
cd cli && go run . <vault>    # run CLI against a vault
cd cli && go test ./...       # run Go tests
cd cli && go test ./internal/...  # test a specific package
cd obsidian-plugin && npm run build  # build Obsidian plugin
```

## Repository structure

```
├── cli/               # Go CLI for batch vault processing
├── src/               # Obsidian plugin TypeScript source
├── main.js            # Built plugin (for Obsidian)
├── manifest.json      # Plugin manifest
├── package.json       # Node dependencies
├── esbuild.config.mjs # Plugin build config
├── tsconfig.json      # TypeScript config
└── README.md          # Plugin documentation
```

## Obsidian plugin (TypeScript)

- `src/main.ts` — Entry point: registers commands, settings tab, orchestrates download/inject
- `src/link.ts` — `parseAll()`: regex finds Excalidraw URLs (`#json=` and `#room=` formats)
- `src/downloader.ts` — `download()`: fetches from API, decrypts WS messages, saves to vault
- `src/crypto.ts` — AES-128-GCM decryption
- `src/injector.ts` — `injectAll()`: adds `([local copy...])` annotations inline

## Go CLI (`cli/`)

- `main.go` — Entry point: validates args, sets up dual logger, creates `_resources/`, walks vault
- `internal/vault/walker.go` — Recursive `.md` file discovery
- `internal/excalidraw/link.go` — Regex parsing of Excalidraw URLs
- `internal/excalidraw/downloader.go` — API fetch and decryption
- `internal/excalidraw/crypto.go` — AES-128-GCM decrypt
- `internal/injector/injector.go` — Annotation injection

## Excalidraw API endpoints

- `#json=ID,KEY` → `https://json.excalidraw.com/api/v2/{ID}` (static shared scene)
- `#room=ID,KEY` → `https://room.excalidraw.com/api/v2/{ID}` (collaborative room)

## Idempotency

Running multiple times is safe — existing files are skipped, already-annotated URLs are skipped, failures are retried.

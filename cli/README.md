# Excalidraw Downloader CLI

Go CLI tool that walks an Obsidian vault, finds every Excalidraw shared link in `.md` files, downloads the drawing, and injects a local-copy annotation — so your notes keep working even when the remote room expires.

## Installation

```bash
git clone https://github.com/aranajuan/obsidian-excalidraw-link-downloader
cd obsidian-excalidraw-link-downloader
go build -o excalidraw-downloader ./cli
```

Requires Go 1.22+.

## Usage

```bash
./excalidraw-downloader [--refresh] <vault-root> [process-path]
```

| Argument | Description |
|----------|-------------|
| `vault-root` | Root directory of the Obsidian vault. Must contain `.obsidian/`. |
| `process-path` | Optional subtree to process. Defaults to `vault-root`. |

| Flag | Description |
|------|-------------|
| `--refresh` | Re-download all scenes, even those cached. Updates annotation dates. |

## Link types supported

- `#json=ID,KEY` — Static scenes fetched from `json.excalidraw.com`
- `#room=ID,KEY` — Collaborative rooms via WebSocket

## Annotation format

```
bare URL:
https://excalidraw.com/#room=abc,KEY ([local copy 24/03/2026](_resources/excalidraw-abc.excalidraw))

markdown link:
[my drawing](https://excalidraw.com/#room=abc,KEY) ([local copy 24/03/2026](_resources/excalidraw-abc.excalidraw))
```

On failure:
```
https://excalidraw.com/#room=abc,KEY (⚠ download failed)
```

## Idempotency

Running multiple times is safe — files already on disk are skipped, already-annotated URLs are skipped, failures are retried.

## `--refresh` mode

Re-downloads all scenes, updates dates, retries failures. If a re-download fails the old file and annotation are preserved.

## Attachment folder modes

Supports all four Obsidian attachment modes (vault root, fixed folder, same as note, subfolder of note).

## Log file

Written to `<vault-root>/_excalidraw-downloader.log`, appended on each run.

## Output summary

After processing:
```
=== done ===
  Downloaded:   12
  Cached:        3
  Empty:         1
  Unreachable:   2
```

See [INTERNALS.md](../INTERNALS.md) for detailed architecture.
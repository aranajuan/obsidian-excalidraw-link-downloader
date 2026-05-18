// Port of internal/excalidraw/downloader.go
// Uses Obsidian's requestUrl (bypasses CORS) for #json= links.
// Uses native WebSocket for #room= links.

import { requestUrl, App, Notice } from 'obsidian';
import { decrypt } from './crypto';
import { downloadRoom, ErrEmptyCanvas, RoomError } from './room-client';
import { apiUrl, Link } from './link';

export { ErrEmptyCanvas };

const MAX_ROOM_RETRIES = 5;
const MAX_DOWNLOAD_BYTES = 50 * 1024 * 1024; // 50 MB

export interface DownloadResult {
  /** Vault-relative path where the file was written. */
  destPath: string;
  /** True when the file already existed and no network request was made. */
  cached: boolean;
}

/**
 * Downloads link to destDir (vault-relative), returning the file path.
 * When force=true, skips the cache check and re-downloads.
 */
export async function download(
  link: Link,
  destDir: string,
  app: App,
  force: boolean,
  notify?: (msg: string) => void,
  signal?: AbortSignal,
): Promise<DownloadResult> {
  const filename = `excalidraw-${link.id}.excalidraw`;
  const destPath = destDir ? `${destDir}/${filename}` : filename;

  // Cache check
  if (!force && await app.vault.adapter.exists(destPath)) {
    return { destPath, cached: true };
  }

  const data = link.kind === 'room'
    ? await downloadRoomWithRetry(link, notify, signal)
    : await downloadJson(link, signal);

  // Ensure the destination directory exists
  if (destDir) {
    await ensureDir(destDir, app);
  }

  await app.vault.adapter.writeBinary(
    destPath,
    data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer
  );
  return { destPath, cached: false };
}

/**
 * Room download flow:
 * 1. Attempt 1: connect. If room is empty (first-in-room), close immediately.
 * 2. Open browser, wait 10s so the browser can load and join the room.
 * 3. Reconnect (attempts 2+): browser is now in room; server will trigger a
 *    re-broadcast when we re-join as a new participant.
 * Browser is only ever opened once (first iteration).
 */
async function downloadRoomWithRetry(
  link: Link,
  notify?: (msg: string) => void,
  signal?: AbortSignal,
): Promise<Uint8Array> {
  if (signal?.aborted) throw new RoomError("aborted", false);

  // Attempt 1
  try {
    return await downloadRoom(link, notify, signal);
  } catch (err) {
    if (!(err instanceof RoomError) || !err.firstInRoom) {
      throw err;
    }
    // Room was empty — fall through to open browser and retry.
  }

  new Notice('Excalidraw: Opening browser to restore room data. If blocked by popup blocker, open this URL manually: ' + link.url, 15_000);
  notify?.('Excalidraw: empty room — opening browser, reconnecting in 10s…');
  window.open(link.url);
  await delay(10_000, signal);

  // Attempts 2+: browser is in room, no need to open it again.
  for (let attempt = 2; attempt <= MAX_ROOM_RETRIES; attempt++) {
    if (signal?.aborted) throw new RoomError("aborted", false);
    try {
      return await downloadRoom(link, notify, signal);
    } catch (err) {
      notify?.(`Excalidraw: attempt ${attempt} — reconnecting in 3s…`);
      await delay(3_000, signal);
    }
  }
  throw new RoomError(`room download failed after ${MAX_ROOM_RETRIES} attempts`, false);
}

async function downloadJson(link: Link, signal?: AbortSignal): Promise<Uint8Array> {
  const url = apiUrl(link);
  const MAX_ATTEMPTS = 4;
  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    try {
      const fetchPromise = requestUrl({ url, method: 'GET' });
      const resp = signal
        ? await Promise.race([
            fetchPromise,
            new Promise<never>((_, reject) => {
              if (signal.aborted) {
                reject(new Error("aborted"));
                return;
              }
              signal.addEventListener('abort', () => reject(new Error("aborted")), { once: true });
            }),
          ])
        : await fetchPromise;
      if (resp.status !== 200) {
        throw new Error(`HTTP ${resp.status} fetching ${url}`);
      }
      if (resp.arrayBuffer.byteLength > MAX_DOWNLOAD_BYTES) {
        throw new Error(`response too large: ${resp.arrayBuffer.byteLength} bytes, max ${MAX_DOWNLOAD_BYTES}`);
      }
      return decrypt(link.key, resp.arrayBuffer);
    } catch (err) {
      if (attempt === MAX_ATTEMPTS) throw err;
      await delay(Math.pow(2, attempt - 1) * 1000, signal);
    }
  }
  throw new Error('unreachable');
}

function delay(ms: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(new RoomError("aborted", false));
      return;
    }
    const timer = setTimeout(resolve, ms);
    signal?.addEventListener('abort', () => {
      clearTimeout(timer);
      reject(new RoomError("aborted", false));
    }, { once: true });
  });
}

async function ensureDir(path: string, app: App): Promise<void> {
  if (await app.vault.adapter.exists(path)) return;
  await app.vault.createFolder(path);
}

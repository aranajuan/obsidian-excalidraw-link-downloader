// Port of internal/excalidraw/downloader.go
// Uses Obsidian's requestUrl (bypasses CORS) for #json= links.
// Uses native WebSocket for #room= links.

import { requestUrl } from 'obsidian';
import { App } from 'obsidian';
import { decrypt } from './crypto';
import { downloadRoom, ErrEmptyCanvas, RoomError } from './room-client';
import { apiUrl, Link } from './link';

export { ErrEmptyCanvas };

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
): Promise<DownloadResult> {
  const filename = `excalidraw-${link.id}.excalidraw`;
  const destPath = destDir ? `${destDir}/${filename}` : filename;

  // Cache check
  if (!force && await app.vault.adapter.exists(destPath)) {
    return { destPath, cached: true };
  }

  const data = link.kind === 'room'
    ? await downloadRoomWithRetry(link, notify)
    : await downloadJson(link);

  // Ensure the destination directory exists
  if (destDir) {
    await ensureDir(destDir, app);
  }

  await app.vault.adapter.writeBinary(destPath, data.buffer as ArrayBuffer);
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
): Promise<Uint8Array> {
  // Attempt 1
  try {
    return await downloadRoom(link, notify);
  } catch (err) {
    if (!(err instanceof RoomError) || !err.firstInRoom) {
      throw err;
    }
    // Room was empty — fall through to open browser and retry.
  }

  notify?.('Excalidraw: sala vacía — abriendo browser, reconectando en 10s…');
  window.open(link.url);
  await new Promise(r => setTimeout(r, 10_000));

  // Attempts 2+: browser is in room, no need to open it again.
  for (let attempt = 2; ; attempt++) {
    try {
      return await downloadRoom(link, notify);
    } catch (err) {
      notify?.(`Excalidraw: intento ${attempt} — reconectando en 3s…`);
      await new Promise(r => setTimeout(r, 3_000));
    }
  }
}

async function downloadJson(link: Link): Promise<Uint8Array> {
  const url = apiUrl(link);
  const resp = await requestUrl({ url, method: 'GET' });
  if (resp.status !== 200) {
    throw new Error(`HTTP ${resp.status} fetching ${url}`);
  }
  return decrypt(link.key, resp.arrayBuffer);
}

async function ensureDir(path: string, app: App): Promise<void> {
  if (await app.vault.adapter.exists(path)) return;
  await app.vault.createFolder(path);
}

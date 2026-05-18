// Port of internal/excalidraw/room_client.go
// Uses the 'ws' Node.js library (bundled) to support custom headers,
// specifically Origin: https://excalidraw.com which the room server requires.

import WebSocket from 'ws';
import { decryptGCM } from './crypto';
import type { Link } from './link';

const ROOM_WS_URL = 'wss://oss-collab.excalidraw.com/socket.io/?EIO=4&transport=websocket';

// Time to get a broadcast from an already-populated room.
const BROADCAST_TIMEOUT_MS = 30_000;
// After an empty broadcast: wait for the next one.
const EMPTY_RETRY_TIMEOUT_MS = 45_000;

export const ErrEmptyCanvas = new Error(
  "room broadcast has no elements — data may only exist on another participant's browser"
);

/**
 * Thrown when downloadRoom exits without receiving a usable broadcast.
 * `firstInRoom` = true means the room was empty when we joined (nobody was there).
 * `firstInRoom` = false means some other error (connection failure, timeout, etc.).
 */
export class RoomError extends Error {
  constructor(message: string, public readonly firstInRoom: boolean) {
    super(message);
    this.name = 'RoomError';
  }
}

/**
 * Connects to the Excalidraw room, waits for a client-broadcast, decrypts it
 * and returns the .excalidraw file bytes.
 *
 * If the room is empty (first-in-room), closes the connection immediately and
 * throws RoomError(firstInRoom=true) — the caller decides what to do next.
 *
 * Does NOT open a browser. That responsibility belongs to the caller.
 */
export function downloadRoom(link: Link, notify?: (msg: string) => void): Promise<Uint8Array> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(ROOM_WS_URL, {
      headers: { Origin: 'https://excalidraw.com' },
    });

    type State = 'eio-open' | 'sio-ack' | 'joined';
    let state: State = 'eio-open';

    let pendingBinaryCount = 0;
    let binFrames: Buffer[] = [];
    let isBroadcast = false;

    let timer: ReturnType<typeof setTimeout>;

    function resetTimer(ms: number) {
      clearTimeout(timer);
      timer = setTimeout(() => {
        ws.close();
        reject(new RoomError(`timed out after ${ms / 1000}s`, false));
      }, ms);
    }

    function done(result: Uint8Array | Error) {
      clearTimeout(timer);
      ws.close();
      result instanceof Uint8Array ? resolve(result) : reject(result);
    }

    resetTimer(BROADCAST_TIMEOUT_MS);

    ws.on('error', (err) => {
      done(new RoomError(`WebSocket error: ${err.message}`, false));
    });

    ws.on('message', async (data: Buffer | string, isBinary: boolean) => {
      try {
        // ── Binary frame ─────────────────────────────────────────────────
        if (isBinary) {
          binFrames.push(data as Buffer);
          if (binFrames.length < pendingBinaryCount) return;

          if (isBroadcast && binFrames.length >= 2) {
            const b0 = binFrames[0], b1 = binFrames[1];
            const ciphertext = b0.buffer.slice(b0.byteOffset, b0.byteOffset + b0.byteLength);
            const iv         = b1.buffer.slice(b1.byteOffset, b1.byteOffset + b1.byteLength);

            const plaintext = await decryptGCM(link.key, ciphertext, iv);
            const fileData  = buildExcalidrawFile(plaintext);

            if (fileData === null) {
              // Browser still initialising — reset timer and keep waiting.
              pendingBinaryCount = 0;
              binFrames = [];
              isBroadcast = false;
              notify?.('Excalidraw: broadcast vacío, esperando…');
              resetTimer(EMPTY_RETRY_TIMEOUT_MS);
              return;
            }

            done(fileData);
            return;
          }
          // Non-broadcast binary event — discard.
          pendingBinaryCount = 0;
          binFrames = [];
          isBroadcast = false;
          return;
        }

        // ── Text frame ───────────────────────────────────────────────────
        const text = (data as Buffer).toString();

        if (text === '2') { ws.send('3'); return; } // EIO ping → pong

        if (state === 'eio-open') {
          if (!text.startsWith('0')) {
            done(new RoomError(`unexpected EIO packet: ${text.slice(0, 60)}`, false));
            return;
          }
          state = 'sio-ack';
          ws.send('40');
          return;
        }

        if (state === 'sio-ack') {
          if (text.startsWith('40')) {
            state = 'joined';
            ws.send(`42["join-room","${link.id}"]`);
          }
          return;
        }

        // state === 'joined'

        // Socket.IO binary event header: "45N-[...]"
        if (text.startsWith('45')) {
          const dashIdx = text.indexOf('-');
          if (dashIdx > 2) {
            const n = parseInt(text.slice(2, dashIdx), 10);
            if (!isNaN(n) && n > 0) {
              pendingBinaryCount = n;
              isBroadcast = text.slice(dashIdx + 1).includes('"client-broadcast"');
              binFrames = [];
              return;
            }
          }
        }

        // Socket.IO text event: "42[...]"
        if (text.startsWith('42') && text.includes('"first-in-room"')) {
          // Room is empty — close immediately so the caller can open the browser
          // and wait before reconnecting.
          done(new RoomError('first-in-room: sala vacía al conectar', true));
        }
      } catch (err) {
        done(new RoomError(err instanceof Error ? err.message : String(err), false));
      }
    });
  });
}

function buildExcalidrawFile(plaintext: Uint8Array): Uint8Array | null {
  const scene = JSON.parse(new TextDecoder().decode(plaintext)) as {
    type: string;
    payload: { elements: unknown[] };
  };
  if (!scene.payload?.elements?.length) return null;

  return new TextEncoder().encode(JSON.stringify({
    type: 'excalidraw',
    version: 2,
    source: 'https://excalidraw.com',
    elements: scene.payload.elements,
    appState: { viewBackgroundColor: '#ffffff' },
    files: {},
  }, null, 2));
}

// Port of internal/excalidraw/crypto.go — uses Web Crypto API (crypto.subtle)

function base64urlDecode(s: string): Uint8Array {
  // base64url (no padding) → standard base64 → binary
  let b64 = s.replace(/-/g, '+').replace(/_/g, '/');
  while (b64.length % 4 !== 0) b64 += '=';
  const raw = atob(b64);
  const out = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  return out;
}

async function importKey(keyStr: string): Promise<CryptoKey> {
  const keyBytes = base64urlDecode(keyStr);
  return crypto.subtle.importKey('raw', keyBytes as BufferSource, { name: 'AES-GCM' }, false, ['decrypt']);
}

/**
 * Decrypts an Excalidraw static scene.
 * Wire format: 12-byte IV || ciphertext || 16-byte GCM tag
 */
export async function decrypt(keyStr: string, data: ArrayBuffer): Promise<Uint8Array> {
  if (data.byteLength < 12) throw new Error(`payload too short: ${data.byteLength} bytes`);
  const key = await importKey(keyStr);
  const iv = data.slice(0, 12);
  const ciphertext = data.slice(12);
  const plaintext = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext);
  return new Uint8Array(plaintext);
}

/**
 * Decrypts a room broadcast where IV and ciphertext arrive as separate binary frames.
 */
export async function decryptGCM(keyStr: string, ciphertext: ArrayBuffer, iv: ArrayBuffer): Promise<Uint8Array> {
  const key = await importKey(keyStr);
  const plaintext = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext);
  return new Uint8Array(plaintext);
}

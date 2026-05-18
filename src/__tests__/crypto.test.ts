import { describe, it, expect } from 'vitest';
import { decrypt, decryptGCM } from '../crypto';

describe('decrypt', () => {
  it('throws on too-short payload', async () => {
    const shortData = new Uint8Array(10).buffer;
    await expect(decrypt('key', shortData)).rejects.toThrow('too short');
  });
});

describe('decryptGCM', () => {
  it('throws on invalid key', async () => {
    const ciphertext = new Uint8Array(32).buffer;
    const iv = new Uint8Array(12).buffer;
    await expect(decryptGCM('invalid_key_here_123!', ciphertext, iv))
      .rejects.toThrow();
  });
});

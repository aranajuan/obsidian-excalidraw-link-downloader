import { describe, it, expect } from 'vitest';
import { parseAll } from '../link';

describe('parseAll', () => {
  it('extracts json link kind, id, key', () => {
    const links = parseAll('https://excalidraw.com/#json=abc123,key_abc-123', false);
    expect(links).toHaveLength(1);
    expect(links[0].kind).toBe('json');
    expect(links[0].id).toBe('abc123');
    expect(links[0].key).toBe('key_abc-123');
  });

  it('extracts room link kind, id, key', () => {
    const links = parseAll('https://excalidraw.com/#room=xyz789,key-xyz', false);
    expect(links).toHaveLength(1);
    expect(links[0].kind).toBe('room');
    expect(links[0].id).toBe('xyz789');
    expect(links[0].key).toBe('key-xyz');
  });

  it('returns empty array for content without links', () => {
    expect(parseAll('just some text', false)).toHaveLength(0);
  });

  it('deduplicates multiple same URLs', () => {
    const url = 'https://excalidraw.com/#json=abc,key';
    const links = parseAll(`${url}\n${url}`, false);
    expect(links).toHaveLength(1);
  });

  it('marks already-annotated links as refresh', () => {
    const links = parseAll('([[excalidraw-abc.excalidraw|local copy 01/01/2026]]) https://excalidraw.com/#json=abc,key', false);
    expect(links).toHaveLength(1);
    expect(links[0].refresh).toBe(true);
  });

  it('handles key with underscores and hyphens', () => {
    const links = parseAll('https://excalidraw.com/#json=abc,key_with-123', false);
    expect(links).toHaveLength(1);
    expect(links[0].key).toBe('key_with-123');
  });

  it('returns empty for empty content', () => {
    expect(parseAll('', false)).toHaveLength(0);
  });
});

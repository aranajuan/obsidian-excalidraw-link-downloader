import { describe, it, expect } from 'vitest';
import { injectOne, hasExistingLocalCopy, buildAnnotation, ANNOTATION_WARNING, ANNOTATION_REFRESH_FAILED, ANNOTATION_EMPTY_CANVAS } from '../injector';

describe('injectOne', () => {
  it('adds annotation for fresh URL', () => {
    const result = injectOne('text https://excalidraw.com/#json=abc,key more', 'https://excalidraw.com/#json=abc,key', '([[test.excalidraw|local copy 01/01/2026]])', false);
    expect(result).toContain('([[test.excalidraw|local copy 01/01/2026]])');
  });

  it('skips already-annotated URL in non-refresh mode', () => {
    const content = 'text https://excalidraw.com/#json=abc,key ([[existing.excalidraw|local copy 01/01/2026]])';
    const result = injectOne(content, 'https://excalidraw.com/#json=abc,key', '([[new.excalidraw|local copy 02/02/2026]])', false);
    expect(result).not.toContain('([[new.excalidraw');
    expect(result).toContain('([[existing.excalidraw');
  });

  it('uses download failed annotation', () => {
    const result = injectOne('bare https://excalidraw.com/#json=abc,key end', 'https://excalidraw.com/#json=abc,key', ANNOTATION_WARNING, false);
    expect(result).toContain(ANNOTATION_WARNING);
  });

  it('replaces old annotation in refresh mode', () => {
    const content = 'text https://excalidraw.com/#json=abc,key ([[old.excalidraw|local copy 01/01/2026]])';
    const result = injectOne(content, 'https://excalidraw.com/#json=abc,key', '([[new.excalidraw|local copy 02/02/2026]])', true);
    expect(result).toContain('([[new.excalidraw|local copy 02/02/2026]])');
    expect(result).not.toContain('([[old.excalidraw');
  });
});

describe('hasExistingLocalCopy', () => {
  it('returns true for annotated link', () => {
    expect(hasExistingLocalCopy('text https://excalidraw.com/#json=abc,key ([[f.excalidraw|local copy 01/01/2026]])', 'https://excalidraw.com/#json=abc,key')).toBe(true);
  });

  it('returns false for unprocessed link', () => {
    expect(hasExistingLocalCopy('text https://excalidraw.com/#json=abc,key end', 'https://excalidraw.com/#json=abc,key')).toBe(false);
  });
});

describe('buildAnnotation', () => {
  it('returns wikilink format', () => {
    const result = buildAnnotation('excalidraw-123.excalidraw');
    expect(result).toMatch(/^\(\[\[excalidraw-123\.excalidraw\|local copy \d{2}\/\d{2}\/\d{4}\]\]\)$/);
  });
});

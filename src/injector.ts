// Port of internal/injector/injector.go — pure string manipulation, no Obsidian API.

export type AttachmentMode = 'vault-root' | 'fixed-folder' | 'same-as-note' | 'subfolder';

export interface AttachmentConfig {
  mode: AttachmentMode;
  folderName: string; // used by fixed-folder and subfolder modes
}

/**
 * Parses Obsidian's attachmentFolderPath setting into a structured config.
 * Mirrors parseAttachmentFolderPath in main.go.
 */
export function parseAttachmentFolderPath(p: string | undefined): AttachmentConfig {
  if (!p) return { mode: 'vault-root', folderName: '' };

  if (p.startsWith('/')) {
    return { mode: 'fixed-folder', folderName: p.slice(1) };
  }
  if (p === './' || p === '.') {
    return { mode: 'same-as-note', folderName: '' };
  }
  if (p.startsWith('./')) {
    return { mode: 'subfolder', folderName: p.slice(2) };
  }
  // Plain name without prefix — treat as fixed folder.
  return { mode: 'fixed-folder', folderName: p };
}

/**
 * Returns the vault-relative destination directory for a given note.
 * noteDir is the vault-relative directory of the note (empty string = vault root).
 */
export function resolveDestDir(cfg: AttachmentConfig, noteDir: string, vaultRoot = ''): string {
  switch (cfg.mode) {
    case 'vault-root':    return vaultRoot;
    case 'fixed-folder':  return cfg.folderName;
    case 'same-as-note':  return noteDir;
    case 'subfolder':     return noteDir ? `${noteDir}/${cfg.folderName}` : cfg.folderName;
  }
}

// ── Annotation injection ──────────────────────────────────────────────────────

/**
 * Builds the annotation string for a successfully downloaded file.
 * Format: ([[excalidraw-ID.excalidraw|local copy DD/MM/YYYY]])
 */
export function buildAnnotation(filename: string): string {
  const now = new Date();
  const dd = String(now.getDate()).padStart(2, '0');
  const mm = String(now.getMonth() + 1).padStart(2, '0');
  const yyyy = now.getFullYear();
  return `([[${filename}|local copy ${dd}/${mm}/${yyyy}]])`;
}

export const ANNOTATION_WARNING       = '(⚠ download failed)';
export const ANNOTATION_REFRESH_FAILED = '(⚠ refresh failed)';
export const ANNOTATION_EMPTY_CANVAS  = '(⚠ empty canvas — ask the author to open the room, then re-run the downloader)';

/**
 * Checks if the first occurrence of url in content already has a local-copy annotation.
 * Used in refresh mode to decide whether to preserve the existing annotation on failure.
 */
export function hasExistingLocalCopy(content: string, url: string): boolean {
  const idx = content.indexOf(url);
  if (idx === -1) return false;
  let after = content.slice(idx + url.length);
  if (after.startsWith(')')) after = after.slice(1);
  return after.replace(/^[ \t]+/, '').startsWith('([[');
}

/**
 * Applies all url→annotation replacements to content.
 * URLs absent from the map are left untouched.
 */
export function injectAll(
  content: string,
  annotations: Map<string, string>,
  refresh: boolean,
): string {
  for (const [url, annotation] of annotations) {
    content = injectOne(content, url, annotation, refresh);
  }
  return content;
}

/**
 * Inserts annotation after every occurrence of url in content.
 * Handles bare URLs and markdown [text](url) links.
 * Is idempotent: skips occurrences already followed by ([[.
 * In refresh mode, replaces existing ([[...]])  annotations.
 *
 * Mirrors injectOne in injector.go exactly.
 */
export function injectOne(
  content: string,
  url: string,
  annotation: string,
  refresh: boolean,
): string {
  let result = '';
  let remaining = content;

  for (;;) {
    const idx = remaining.indexOf(url);
    if (idx === -1) {
      result += remaining;
      break;
    }

    result += remaining.slice(0, idx + url.length);
    let after = remaining.slice(idx + url.length);

    // Consume closing ) if inside a markdown link [text](url)
    let closeParen = '';
    if (after.startsWith(')')) {
      closeParen = ')';
      after = after.slice(1);
    }

    const trimmed = after.replace(/^[ \t]+/, '');

    if (trimmed.startsWith('([[')) {
      if (refresh && annotation === ANNOTATION_REFRESH_FAILED) {
        // Refresh failed but old local copy exists: keep ([[...]]) and append error marker.
        const end = trimmed.indexOf(']])');
        if (end >= 0) {
          const existing = trimmed.slice(0, end + 3); // ([[...|local copy DATE]])
          after = trimmed.slice(end + 3);
          const rest = after.replace(/^[ \t]+/, '');
          if (rest.startsWith(ANNOTATION_REFRESH_FAILED)) {
            // Already marked — leave as-is
            result += closeParen + ' ' + existing;
            after = rest.slice(ANNOTATION_REFRESH_FAILED.length);
          } else {
            result += closeParen + ' ' + existing + ' ' + ANNOTATION_REFRESH_FAILED;
          }
        } else {
          result += closeParen;
        }
      } else if (refresh) {
        // Successful refresh: replace ([[...]]) and remove any trailing (⚠ refresh failed).
        const end = trimmed.indexOf(']])');
        if (end >= 0) {
          after = trimmed.slice(end + 3);
          const rest = after.replace(/^[ \t]+/, '');
          if (rest.startsWith(ANNOTATION_REFRESH_FAILED)) {
            after = rest.slice(ANNOTATION_REFRESH_FAILED.length);
          }
        }
        result += closeParen + ' ' + annotation;
      } else {
        // Normal mode: already annotated — leave as-is
        result += closeParen;
      }
    } else if (trimmed.startsWith('([local copy')) {
      // Old markdown format — migrate to wikilink
      const inner = trimmed.slice(1); // strip leading (
      const end = indexAfterMarkdownLink(inner);
      if (end >= 0 && end < inner.length && inner[end] === ')') {
        after = inner.slice(end + 1);
      }
      result += closeParen + ' ' + annotation;
    } else if (trimmed.startsWith('[local copy')) {
      // Old bare format — migrate to wikilink
      const end = indexAfterMarkdownLink(trimmed);
      if (end >= 0) after = trimmed.slice(end);
      result += closeParen + ' ' + annotation;
    } else if (trimmed.startsWith('(⚠')) {
      // Previous attempt failed — replace warning with new annotation
      const end = trimmed.indexOf(')');
      if (end >= 0) after = trimmed.slice(end + 1);
      result += closeParen + ' ' + annotation;
    } else {
      result += closeParen + ' ' + annotation;
    }

    remaining = after;
  }

  return result;
}

/**
 * Returns the index just past the closing ')' of the first markdown link
 * "[text](url)" in s, or -1 if not found.
 * Mirrors indexAfterMarkdownLink in injector.go.
 */
function indexAfterMarkdownLink(s: string): number {
  const closeBracket = s.indexOf('](');
  if (closeBracket === -1) return -1;
  const rest = s.slice(closeBracket + 2);
  const closeParenIdx = rest.indexOf(')');
  if (closeParenIdx === -1) return -1;
  return closeBracket + 2 + closeParenIdx + 1;
}

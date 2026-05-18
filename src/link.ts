// Port of internal/excalidraw/link.go

export type Kind = 'json' | 'room';

export interface Link {
  url: string;
  id: string;
  key: string;
  kind: Kind;
}

export function apiUrl(link: Link): string {
  return link.kind === 'room'
    ? `https://room.excalidraw.com/api/v2/${link.id}`
    : `https://json.excalidraw.com/api/v2/${link.id}`;
}

// Same regex as urlRe in link.go
const URL_RE = /https?:\/\/excalidraw\.com\/#(json|room)=([a-zA-Z0-9]+),([a-zA-Z0-9_\-]+)/g;

/**
 * Returns all unique Excalidraw URLs that have at least one unannotated occurrence.
 * When refresh=true, already-annotated URLs are also returned.
 */
export function parseAll(content: string, refresh: boolean): Link[] {
  const seen = new Set<string>();
  const links: Link[] = [];

  URL_RE.lastIndex = 0;
  let match: RegExpExecArray | null;

  while ((match = URL_RE.exec(content)) !== null) {
    const url = match[0];
    if (seen.has(url)) continue;
    seen.add(url);

    // Idempotency check on the first occurrence (mirrors link.go)
    const afterPos = match.index + url.length;
    let offset = 0;
    if (afterPos < content.length && content[afterPos] === ')') {
      offset = 1; // closing paren of a markdown link
    }
    const trimmed = content.slice(afterPos + offset).replace(/^[ \t]+/, '');
    if (!refresh && trimmed.startsWith('([[')) {
      continue; // already annotated — skip
    }

    links.push({
      url,
      kind: match[1] as Kind,
      id: match[2],
      key: match[3],
    });
  }

  return links;
}

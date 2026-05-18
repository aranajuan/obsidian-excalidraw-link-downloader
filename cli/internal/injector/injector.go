package injector

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"obsidian-excalidraw-downloader/internal/excalidraw"
)

const (
	annotationWarning        = "(⚠ download failed)"
	annotationEmptyCanvas    = "(⚠ empty canvas — ask the author to open the room, then re-run the downloader)"
	annotationRefreshFailed  = "(⚠ refresh failed)"
)

// AttachmentMode mirrors Obsidian's "Default location for new attachments" setting.
type AttachmentMode int

const (
	// ModeVaultRoot saves attachments to the vault root (attachmentFolderPath = "").
	ModeVaultRoot AttachmentMode = iota
	// ModeFixedFolder saves attachments to a fixed folder in the vault (attachmentFolderPath = "/folder").
	ModeFixedFolder
	// ModeSameAsNote saves attachments next to each note (attachmentFolderPath = "./").
	ModeSameAsNote
	// ModeSubfolder saves attachments in a subfolder relative to each note (attachmentFolderPath = "./folder").
	ModeSubfolder
)

// Processor handles finding and annotating Excalidraw links in markdown files.
type Processor struct {
	vaultPath  string
	mode       AttachmentMode
	folderName string // used by ModeFixedFolder and ModeSubfolder
	refresh    bool   // when true, re-downloads existing files and updates annotations
	log        *log.Logger
}

func New(vaultPath string, mode AttachmentMode, folderName string, refresh bool, logger *log.Logger) *Processor {
	return &Processor{
		vaultPath:  vaultPath,
		mode:       mode,
		folderName: folderName,
		refresh:    refresh,
		log:        logger,
	}
}

// resourcesDirFor returns the attachment directory for a given note path.
func (p *Processor) resourcesDirFor(notePath string) string {
	switch p.mode {
	case ModeVaultRoot:
		return p.vaultPath
	case ModeFixedFolder:
		return filepath.Join(p.vaultPath, p.folderName)
	case ModeSameAsNote:
		return filepath.Dir(notePath)
	case ModeSubfolder:
		return filepath.Join(filepath.Dir(notePath), p.folderName)
	}
	return p.vaultPath
}

// ProcessFile finds Excalidraw links in a markdown file, downloads them, and
// injects annotations (local copy link or warning) next to each URL.
// Returns counts for: newly downloaded, already cached on disk, empty canvas,
// errors, and (in refresh mode) links that failed to refresh but kept their
// existing local copy unchanged.
func (p *Processor) ProcessFile(path string) (downloaded, cached, empty, errCount, refreshFailed int) {
	data, err := os.ReadFile(path)
	if err != nil {
		p.log.Printf("  ERROR reading %s: %v", relPath(p.vaultPath, path), err)
		return 0, 0, 0, 1, 0
	}

	content := string(data)
	links := excalidraw.ParseAll(content, p.refresh)
	if len(links) == 0 {
		return 0, 0, 0, 0, 0
	}

	p.log.Printf("FILE %s — %d link(s) found", relPath(p.vaultPath, path), len(links))

	resourcesDir := p.resourcesDirFor(path)
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		p.log.Printf("  ERROR creating resources dir %s: %v", resourcesDir, err)
		return 0, 0, 0, len(links), 0
	}

	// In refresh mode, track which URLs already had a local copy annotation so
	// that a failed re-download leaves those annotations untouched.
	hadLocalCopy := make(map[string]bool, len(links))
	if p.refresh {
		for _, link := range links {
			hadLocalCopy[link.URL] = hasExistingLocalCopy(content, link.URL)
		}
	}

	annotations := make(map[string]string, len(links))
	for _, link := range links {
		localFile, wasCached, dlErr := p.downloadTo(link, resourcesDir)
		if dlErr != nil {
			if p.refresh && hadLocalCopy[link.URL] {
				// New download failed but the old file is still intact —
				// keep existing annotation and append refresh failure marker.
				p.log.Printf("  REFRESH FAIL [%s] %s: %v — keeping existing copy", link.Kind, link.ID, dlErr)
				annotations[link.URL] = annotationRefreshFailed
				refreshFailed++
			} else if errors.Is(dlErr, excalidraw.ErrEmptyCanvas) {
				p.log.Printf("  EMPTY [room] %s: canvas empty — data may exist on another browser", link.ID)
				annotations[link.URL] = annotationEmptyCanvas
				empty++
			} else {
				p.log.Printf("  ERROR [%s] %s: %v", link.Kind, link.ID, dlErr)
				annotations[link.URL] = annotationWarning
				errCount++
			}
		} else {
			filename := filepath.Base(localFile)
			date := time.Now().Format("02/01/2006")
			annotation := fmt.Sprintf("([[%s|local copy %s]])", filename, date)
			annotations[link.URL] = annotation
			if wasCached {
				p.log.Printf("  CACHE [%s] %s → %s", link.Kind, link.ID, filename)
				cached++
			} else {
				p.log.Printf("  OK    [%s] %s → %s", link.Kind, link.ID, filename)
				downloaded++
			}
		}
	}

	newContent := injectAll(content, annotations, p.refresh)
	if newContent == content {
		return downloaded, cached, empty, errCount, refreshFailed
	}

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		p.log.Printf("  ERROR writing %s: %v", relPath(p.vaultPath, path), err)
		return downloaded, cached, empty, errCount + 1, refreshFailed
	}

	return downloaded, cached, empty, errCount, refreshFailed
}

// hasExistingLocalCopy reports whether the first occurrence of url in content
// is already followed by a ([local copy annotation.
func hasExistingLocalCopy(content, url string) bool {
	idx := strings.Index(content, url)
	if idx == -1 {
		return false
	}
	after := content[idx+len(url):]
	if len(after) > 0 && after[0] == ')' {
		after = after[1:]
	}
	return strings.HasPrefix(strings.TrimLeft(after, " \t"), "([[")
}

// downloadTo calls excalidraw.Download with retry logic:
//   - Room links: retry indefinitely on transient errors (timeout, connection).
//     ErrEmptyCanvas is not retried — it means the data is on another machine.
//   - Non-room links: single attempt, failures are permanent.
//
// Returns (localPath, cached, err) where cached is true when the file already
// existed on disk and no network request was made.
func (p *Processor) downloadTo(link excalidraw.Link, destDir string) (string, bool, error) {
	if link.Kind != excalidraw.Room {
		return excalidraw.Download(link, destDir, p.refresh)
	}

	for attempt := 1; ; attempt++ {
		localFile, cached, err := excalidraw.Download(link, destDir, p.refresh)
		if err == nil {
			return localFile, cached, nil
		}
		if errors.Is(err, excalidraw.ErrEmptyCanvas) {
			return "", false, err // not retryable: data is on another participant's browser
		}
		p.log.Printf("  RETRY [room] %s (attempt %d): %v — retrying in 5s...", link.ID, attempt, err)
		time.Sleep(5 * time.Second)
	}
}

// injectAll applies all URL→annotation pairs to content.
func injectAll(content string, annotations map[string]string, refresh bool) string {
	for url, annotation := range annotations {
		content = injectOne(content, url, annotation, refresh)
	}
	return content
}

// injectOne inserts annotation after each occurrence of url in content.
// Handles both bare URLs and markdown links [text](url).
// Is idempotent: skips occurrences that already have a ([local copy annotation,
// unless refresh is true — in that case the existing annotation is replaced.
func injectOne(content, url, annotation string, refresh bool) string {
	var b strings.Builder
	remaining := content

	for {
		idx := strings.Index(remaining, url)
		if idx == -1 {
			b.WriteString(remaining)
			break
		}

		b.WriteString(remaining[:idx+len(url)])
		after := remaining[idx+len(url):]

		// If the URL is inside a markdown link [text](URL), consume the closing ).
		closeParen := ""
		if len(after) > 0 && after[0] == ')' {
			closeParen = ")"
			after = after[1:]
		}

		// Check if already annotated right after the link.
		trimmed := strings.TrimLeft(after, " \t")
		if strings.HasPrefix(trimmed, "([[") {
			if refresh && annotation == annotationRefreshFailed {
				// Refresh failed but old local copy exists: keep ([[...]]) and append error marker.
				if end := strings.Index(trimmed, "]])"); end >= 0 {
					existing := trimmed[:end+3]
					after = trimmed[end+3:]
					rest := strings.TrimLeft(after, " \t")
					if strings.HasPrefix(rest, annotationRefreshFailed) {
						// Already marked — leave as-is.
						b.WriteString(closeParen)
						b.WriteString(" ")
						b.WriteString(existing)
					} else {
						b.WriteString(closeParen)
						b.WriteString(" ")
						b.WriteString(existing)
						b.WriteString(" ")
						b.WriteString(annotationRefreshFailed)
					}
				} else {
					b.WriteString(closeParen)
				}
			} else if refresh {
				// Replace existing wikilink annotation ([[file|local copy DATE]]).
				// Strip any trailing (⚠ refresh failed) if present.
				if end := strings.Index(trimmed, "]])"); end >= 0 {
					after = trimmed[end+3:]
				}
				rest := strings.TrimLeft(after, " \t")
				if strings.HasPrefix(rest, annotationRefreshFailed) {
					after = rest[len(annotationRefreshFailed):]
				}
				b.WriteString(closeParen)
				b.WriteString(" ")
				b.WriteString(annotation)
			} else {
				// Already annotated — leave as-is.
				b.WriteString(closeParen)
			}
		} else if strings.HasPrefix(trimmed, "([local copy") {
			// Old markdown format — migrate to wikilink.
			inner := trimmed[1:] // strip leading (
			if end := indexAfterMarkdownLink(inner); end >= 0 && end < len(inner) && inner[end] == ')' {
				after = inner[end+1:]
			}
			b.WriteString(closeParen)
			b.WriteString(" ")
			b.WriteString(annotation)
		} else if strings.HasPrefix(trimmed, "[local copy") {
			// Old bare format — migrate to wikilink.
			if endIdx := indexAfterMarkdownLink(trimmed); endIdx >= 0 {
				after = trimmed[endIdx:]
			}
			b.WriteString(closeParen)
			b.WriteString(" ")
			b.WriteString(annotation)
		} else if strings.HasPrefix(trimmed, "(⚠") {
			// Previous attempt failed — replace old warning with new annotation.
			if endIdx := strings.Index(trimmed, ")"); endIdx >= 0 {
				after = trimmed[endIdx+1:]
			}
			b.WriteString(closeParen)
			b.WriteString(" ")
			b.WriteString(annotation)
		} else {
			b.WriteString(closeParen)
			b.WriteString(" ")
			b.WriteString(annotation)
		}

		remaining = after
	}

	return b.String()
}

// indexAfterMarkdownLink returns the index just past the closing ')' of the
// first markdown link "[text](url)" in s, or -1 if not found.
func indexAfterMarkdownLink(s string) int {
	closeBracket := strings.Index(s, "](")
	if closeBracket == -1 {
		return -1
	}
	rest := s[closeBracket+2:]
	closeParenIdx := strings.Index(rest, ")")
	if closeParenIdx == -1 {
		return -1
	}
	return closeBracket + 2 + closeParenIdx + 1
}

func relPath(base, path string) string {
	r, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return r
}

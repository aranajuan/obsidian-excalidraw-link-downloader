package excalidraw

import (
	"regexp"
	"strings"
)

type Kind string

const (
	JSON Kind = "json"
	Room Kind = "room"
)

// Link is a parsed Excalidraw URL found in a markdown file.
type Link struct {
	URL  string
	ID   string
	Key  string
	Kind Kind
}

// APIURL returns the HTTP endpoint to fetch the scene data from.
func (l Link) APIURL() string {
	switch l.Kind {
	case Room:
		return "https://room.excalidraw.com/api/v2/" + l.ID
	default:
		return "https://json.excalidraw.com/api/v2/" + l.ID
	}
}

var urlRe = regexp.MustCompile(
	`https?://excalidraw\.com/#(json|room)=([a-zA-Z0-9]+),([a-zA-Z0-9_\-]+)`,
)

// ParseAll returns all unique Excalidraw URLs that have at least one
// unannotated occurrence in content.
// When refresh is true, URLs already annotated with ([local copy are also
// included so they can be re-downloaded and their annotation updated.
func ParseAll(content string, refresh bool) []Link {
	seen := map[string]bool{}
	var links []Link

	for _, m := range urlRe.FindAllStringSubmatchIndex(content, -1) {
		url := content[m[0]:m[1]]
		if seen[url] {
			continue
		}
		seen[url] = true

		// Check if this first occurrence is already annotated.
		// injectOne handles each occurrence independently, so this is just
		// used to avoid triggering a download for fully-processed files.
		afterPos := m[1]
		offset := 0
		if afterPos < len(content) && content[afterPos] == ')' {
			offset = 1 // closing paren of a markdown link
		}
		trimmed := strings.TrimLeft(content[afterPos+offset:], " \t")
		if !refresh && strings.HasPrefix(trimmed, "([[") {
			continue // already successfully downloaded
		}

		links = append(links, Link{
			URL:  url,
			Kind: Kind(content[m[2]:m[3]]),
			ID:   content[m[4]:m[5]],
			Key:  content[m[6]:m[7]],
		})
	}

	return links
}

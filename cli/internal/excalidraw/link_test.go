package excalidraw_test

import (
	"testing"

	"obsidian-excalidraw-downloader/internal/excalidraw"
)

func TestParseAll_JsonLink(t *testing.T) {
	links := excalidraw.ParseAll("https://excalidraw.com/#json=abc123,key_abc-123", false)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Kind != excalidraw.JSON {
		t.Errorf("expected json, got %s", links[0].Kind)
	}
	if links[0].ID != "abc123" {
		t.Errorf("expected abc123, got %s", links[0].ID)
	}
	if links[0].Key != "key_abc-123" {
		t.Errorf("expected key_abc-123, got %s", links[0].Key)
	}
}

func TestParseAll_RoomLink(t *testing.T) {
	links := excalidraw.ParseAll("https://excalidraw.com/#room=xyz789,key_xyz-789", false)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Kind != excalidraw.Room {
		t.Errorf("expected room, got %s", links[0].Kind)
	}
	if links[0].ID != "xyz789" {
		t.Errorf("expected xyz789, got %s", links[0].ID)
	}
	if links[0].Key != "key_xyz-789" {
		t.Errorf("expected key_xyz-789, got %s", links[0].Key)
	}
}

func TestParseAll_Empty(t *testing.T) {
	links := excalidraw.ParseAll("", false)
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestParseAll_Deduplicate(t *testing.T) {
	content := "https://excalidraw.com/#json=abc123,key_abc-123 and again https://excalidraw.com/#json=abc123,key_abc-123"
	links := excalidraw.ParseAll(content, false)
	if len(links) != 1 {
		t.Fatalf("expected 1 unique link, got %d", len(links))
	}
}

func TestParseAll_AnnotatedLink(t *testing.T) {
	content := "https://excalidraw.com/#json=abc123,key_abc-123 ([[excalidraw-abc123.excalidraw|local copy 01/01/2026]])"
	links := excalidraw.ParseAll(content, false)
	if len(links) != 0 {
		t.Fatalf("expected 0 links (already annotated), got %d", len(links))
	}
	links = excalidraw.ParseAll(content, true)
	if len(links) != 1 {
		t.Fatalf("expected 1 link in refresh mode, got %d", len(links))
	}
}

func TestParseAll_EmptyContent(t *testing.T) {
	links := excalidraw.ParseAll("just some text without any excalidraw links", false)
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

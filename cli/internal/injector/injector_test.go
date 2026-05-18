package injector

import (
	"strings"
	"testing"
)

func TestInjectOne_FreshURL(t *testing.T) {
	content := "Check out this drawing: https://excalidraw.com/#json=abc123,key_abc-123"
	annotation := "([[excalidraw-abc123.excalidraw|local copy 01/01/2026]])"
	result := injectOne(content, "https://excalidraw.com/#json=abc123,key_abc-123", annotation, false)
	if !strings.Contains(result, annotation) {
		t.Errorf("expected annotation to be injected, got: %s", result)
	}
}

func TestInjectOne_AlreadyAnnotated(t *testing.T) {
	annotation := "([[excalidraw-abc123.excalidraw|local copy 01/01/2026]])"
	content := "Check out this drawing: https://excalidraw.com/#json=abc123,key_abc-123 " + annotation
	result := injectOne(content, "https://excalidraw.com/#json=abc123,key_abc-123", annotation, false)
	if strings.Count(result, "local copy") != 1 {
		t.Errorf("expected no duplicate annotation, got: %s", result)
	}
}

func TestInjectOne_AnnotationWarning(t *testing.T) {
	content := "Failed link: https://excalidraw.com/#json=abc123,key_abc-123"
	result := injectOne(content, "https://excalidraw.com/#json=abc123,key_abc-123", annotationWarning, false)
	if !strings.Contains(result, annotationWarning) {
		t.Errorf("expected warning annotation, got: %s", result)
	}
}

func TestInjectOne_AnnotationRefreshFailed(t *testing.T) {
	existing := "([[excalidraw-abc123.excalidraw|local copy 01/01/2026]])"
	content := "Drawing: https://excalidraw.com/#json=abc123,key_abc-123 " + existing
	result := injectOne(content, "https://excalidraw.com/#json=abc123,key_abc-123", annotationRefreshFailed, true)
	if !strings.Contains(result, existing) {
		t.Errorf("expected existing annotation to be preserved, got: %s", result)
	}
	if !strings.Contains(result, annotationRefreshFailed) {
		t.Errorf("expected refresh-failed marker to be appended, got: %s", result)
	}
}

func TestHasExistingLocalCopy_True(t *testing.T) {
	content := "https://excalidraw.com/#json=abc123,key_abc-123 ([[excalidraw-abc123.excalidraw|local copy 01/01/2026]])"
	if !hasExistingLocalCopy(content, "https://excalidraw.com/#json=abc123,key_abc-123") {
		t.Error("expected hasExistingLocalCopy to be true")
	}
}

func TestHasExistingLocalCopy_False(t *testing.T) {
	content := "https://excalidraw.com/#json=abc123,key_abc-123"
	if hasExistingLocalCopy(content, "https://excalidraw.com/#json=abc123,key_abc-123") {
		t.Error("expected hasExistingLocalCopy to be false")
	}
}

func TestInjectOne_WikilinkFormat(t *testing.T) {
	content := "Drawing: https://excalidraw.com/#json=abc123,key_abc-123"
	annotation := "([[excalidraw-abc123.excalidraw|local copy 18/05/2026]])"
	result := injectOne(content, "https://excalidraw.com/#json=abc123,key_abc-123", annotation, false)
	if !strings.Contains(result, "([[excalidraw-abc123.excalidraw|local copy 18/05/2026]])") {
		t.Errorf("expected wikilink annotation format, got: %s", result)
	}
}

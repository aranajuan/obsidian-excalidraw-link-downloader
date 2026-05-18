package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"obsidian-excalidraw-downloader/internal/vault"
)

func TestWalk_FindsMdFiles(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "note1.md"), []byte("# Hello"), 0644)
	os.WriteFile(filepath.Join(tmp, "note2.md"), []byte("# World"), 0644)
	os.WriteFile(filepath.Join(tmp, "readme.txt"), []byte("text"), 0644)

	files, err := vault.Walk(tmp)
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 .md files, got %d", len(files))
	}
}

func TestWalk_SkipsHiddenDirs(t *testing.T) {
	tmp := t.TempDir()
	hiddenDir := filepath.Join(tmp, ".hidden")
	os.MkdirAll(hiddenDir, 0755)
	os.WriteFile(filepath.Join(hiddenDir, "secret.md"), []byte("# Secret"), 0644)
	os.WriteFile(filepath.Join(tmp, "visible.md"), []byte("# Visible"), 0644)

	files, err := vault.Walk(tmp)
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 .md file, got %d", len(files))
	}
	if filepath.Base(files[0]) != "visible.md" {
		t.Errorf("expected visible.md, got %s", files[0])
	}
}

func TestWalk_EmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	files, err := vault.Walk(tmp)
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

func TestWalk_SkipsDirsInSkipList(t *testing.T) {
	tmp := t.TempDir()
	skipDir := filepath.Join(tmp, "attachments")
	os.MkdirAll(skipDir, 0755)
	os.WriteFile(filepath.Join(skipDir, "img.md"), []byte("# Image"), 0644)
	os.WriteFile(filepath.Join(tmp, "note.md"), []byte("# Note"), 0644)

	files, err := vault.Walk(tmp, "attachments")
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 .md file, got %d", len(files))
	}
	if filepath.Base(files[0]) != "note.md" {
		t.Errorf("expected note.md, got %s", files[0])
	}
}

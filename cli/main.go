package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"obsidian-excalidraw-downloader/internal/injector"
	"obsidian-excalidraw-downloader/internal/vault"
)

func printHelp(bin string) {
	fmt.Printf(`Usage:
  %s [--refresh] [--yes] <vault-root> [process-path]

Arguments:
  vault-root     Root directory of the Obsidian vault (.obsidian/ must exist here).
  process-path   Subtree to process (default: vault-root). Useful to limit
                 the run to a specific folder inside the vault.

Flags:
  --refresh      Re-download all files, even those already cached. Updates the
                 date in every ([local copy DATE](...)) annotation. Also retries
                 links that previously failed. If a re-download fails, the
                 existing local file and annotation are kept unchanged.
  --yes, -y      Skip confirmation prompts (for scripting). When set, reads
                 attachment config from .obsidian/app.json if available, or
                 falls back to a '_resources' folder in the vault root.

What it does:
  Walks all .md files under process-path, finds Excalidraw URLs
  (#json= and #room= formats), downloads the drawings, and injects a
  ([local copy DATE](...)) annotation next to each link.

  If a previous run left a warning annotation (⚠), the link is retried
  and the warning is replaced with the result.

Link types:
  #json=ID,KEY   Static shared scene — fetched from json.excalidraw.com.
  #room=ID,KEY   Collaborative room — requires the drawing to be open in
                 a browser. If the room is empty, the tool auto-opens the
                 URL in your default browser and waits up to 60 s for the
                 scene to broadcast. The tab is closed automatically after
                 a successful download.

Output summary:
  Downloaded    Scenes fetched and saved for the first time.
  Cached        File already existed on disk — skipped.
  Empty         Room broadcast contained no elements.
  Unreachable   Download failed (HTTP error, timeout, decrypt error, etc.).

Files written:
  <vault-root>/_excalidraw-downloader.log   Full run log.

  The .excalidraw files are saved according to the vault's attachment folder
  setting (read from .obsidian/app.json, or prompted interactively):

    Vault root          <vault-root>/excalidraw-<id>.excalidraw
    Fixed folder        <vault-root>/<folder>/excalidraw-<id>.excalidraw
    Same as note        <note-dir>/excalidraw-<id>.excalidraw
    Subfolder of note   <note-dir>/<folder>/excalidraw-<id>.excalidraw

  The annotation injected into each .md file uses a path relative to the
  note, so it works regardless of which mode is configured.

Examples:
  %s ~/Documents/MyVault
  %s ~/Documents/MyVault ~/Documents/MyVault/Projects/Design

`, bin, bin, bin)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp(os.Args[0])
		if len(os.Args) < 2 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Extract --refresh and --yes flags, leaving positional args intact.
	refresh := false
	yes := false
	positional := os.Args[:1]
	for _, arg := range os.Args[1:] {
		if arg == "--refresh" {
			refresh = true
		} else if arg == "--yes" || arg == "-y" {
			yes = true
		} else {
			positional = append(positional, arg)
		}
	}

	if len(positional) < 2 {
		printHelp(os.Args[0])
		os.Exit(1)
	}

	vaultPath, err := filepath.Abs(positional[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid vault path: %v\n", err)
		os.Exit(1)
	}
	if info, err := os.Stat(vaultPath); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "vault path does not exist or is not a directory: %s\n", vaultPath)
		os.Exit(1)
	}

	processPath := vaultPath
	if len(positional) >= 3 {
		processPath, err = filepath.Abs(positional[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid process path: %v\n", err)
			os.Exit(1)
		}
		if info, err := os.Stat(processPath); err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "process path does not exist or is not a directory: %s\n", processPath)
			os.Exit(1)
		}
		rel, err := filepath.Rel(vaultPath, processPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			fmt.Fprintf(os.Stderr, "process path must be inside the vault: %s\n", processPath)
			os.Exit(1)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	var mode injector.AttachmentMode
	var folderName, description string
	if yes {
		mode, folderName, description = resolveAttachmentConfigNonInteractive(vaultPath)
	} else {
		mode, folderName, description = resolveAttachmentConfig(vaultPath, scanner)
	}

	fmt.Printf("Vault:            %s\n", vaultPath)
	if processPath != vaultPath {
		fmt.Printf("Processing:       %s\n", processPath)
	}
	fmt.Printf("Resources folder: %s\n", description)

	if !yes {
		fmt.Printf("Proceed? [y/N] ")
		scanner.Scan()
		if answer := strings.TrimSpace(strings.ToLower(scanner.Text())); answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	logPath := filepath.Join(vaultPath, "_excalidraw-downloader.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)
	logger.Printf("=== excalidraw-downloader started | vault: %s ===", vaultPath)
	logger.Printf("resources folder: %s", description)

	files, err := vault.Walk(processPath, folderName)
	if err != nil {
		logger.Fatalf("error walking files: %v", err)
	}
	logger.Printf("found %d markdown file(s)", len(files))

	proc := injector.New(vaultPath, mode, folderName, refresh, logger)
	var totalDownloaded, totalCached, totalEmpty, totalErrors, totalRefreshFailed int
	for _, f := range files {
		d, c, e, er, rf := proc.ProcessFile(f)
		totalDownloaded += d
		totalCached += c
		totalEmpty += e
		totalErrors += er
		totalRefreshFailed += rf
	}

	logger.Printf("=== done ===")
	logger.Printf("  Downloaded:   %d", totalDownloaded)
	logger.Printf("  Cached:       %d", totalCached)
	logger.Printf("  Empty:        %d", totalEmpty)
	logger.Printf("  Unreachable:  %d", totalErrors)
	if refresh {
		logger.Printf("  Kept (refresh failed): %d", totalRefreshFailed)
	}
}

// resolveAttachmentConfig reads .obsidian/app.json if available, otherwise
// prompts the user to choose the attachment mode interactively.
func resolveAttachmentConfig(vaultPath string, scanner *bufio.Scanner) (injector.AttachmentMode, string, string) {
	appJSON := filepath.Join(vaultPath, ".obsidian", "app.json")
	data, err := os.ReadFile(appJSON)
	if err != nil {
		fmt.Println(".obsidian/app.json not found.")
		return promptAttachmentConfig(scanner)
	}

	var cfg struct {
		AttachmentFolderPath string `json:"attachmentFolderPath"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println("Could not read .obsidian/app.json.")
		return promptAttachmentConfig(scanner)
	}

	return parseAttachmentFolderPath(cfg.AttachmentFolderPath)
}

// resolveAttachmentConfigNonInteractive reads .obsidian/app.json if available,
// otherwise falls back to a fixed _resources folder. Used with --yes flag.
func resolveAttachmentConfigNonInteractive(vaultPath string) (injector.AttachmentMode, string, string) {
	appJSON := filepath.Join(vaultPath, ".obsidian", "app.json")
	data, err := os.ReadFile(appJSON)
	if err != nil {
		fmt.Println(".obsidian/app.json not found, using default '_resources' folder.")
		return injector.ModeFixedFolder, "_resources", "_resources/  (default)"
	}

	var cfg struct {
		AttachmentFolderPath string `json:"attachmentFolderPath"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Println("Could not read .obsidian/app.json, using default '_resources' folder.")
		return injector.ModeFixedFolder, "_resources", "_resources/  (default)"
	}

	return parseAttachmentFolderPath(cfg.AttachmentFolderPath)
}

// promptAttachmentConfig shows a menu of Obsidian attachment modes and returns
// the selected mode, folder name, and a human-readable description.
func promptAttachmentConfig(scanner *bufio.Scanner) (injector.AttachmentMode, string, string) {
	fmt.Println()
	fmt.Println("Select attachment folder mode (same as Obsidian Settings → Files & Links):")
	fmt.Println()
	fmt.Println("  1) Vault root         — vault root")
	fmt.Println("  2) Fixed folder       — specific folder in vault root (e.g. _resources)")
	fmt.Println("  3) Same folder        — same folder as each note")
	fmt.Println("  4) Subfolder          — subfolder relative to each note (e.g. _resources)")
	fmt.Println()
	fmt.Printf("Selection [1-4]: ")

	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		return injector.ModeVaultRoot, "", "vault root"
	case "2":
		fmt.Printf("Folder name (e.g. _resources): ")
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			name = "_resources"
		}
		name = strings.TrimPrefix(name, "/")
		_, _, desc := parseAttachmentFolderPath("/" + name)
		return injector.ModeFixedFolder, name, desc
	case "3":
		return injector.ModeSameAsNote, "", "same folder as each note"
	case "4":
		fmt.Printf("Subfolder name (e.g. _resources): ")
		scanner.Scan()
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			name = "_resources"
		}
		name = strings.TrimPrefix(strings.TrimPrefix(name, "./"), "/")
		_, _, desc := parseAttachmentFolderPath("./" + name)
		return injector.ModeSubfolder, name, desc
	default:
		fmt.Println("Invalid selection, using fixed folder '_resources'.")
		return injector.ModeFixedFolder, "_resources", "_resources/  (attachmentFolderPath=\"/_resources\")"
	}
}

// parseAttachmentFolderPath converts an Obsidian attachmentFolderPath value
// to an AttachmentMode, folder name, and human-readable description.
//
// Formats:
//   - ""          → vault root
//   - "/folder"   → fixed folder in vault root
//   - "./"        → same folder as note
//   - "./folder"  → subfolder relative to note
func parseAttachmentFolderPath(p string) (injector.AttachmentMode, string, string) {
	switch {
	case p == "":
		return injector.ModeVaultRoot, "", "vault root  (attachmentFolderPath=\"\")"

	case strings.HasPrefix(p, "/"):
		name := strings.TrimPrefix(p, "/")
		return injector.ModeFixedFolder, name,
			fmt.Sprintf("%s/  (attachmentFolderPath=%q)", name, p)

	case p == "./" || p == ".":
		return injector.ModeSameAsNote, "",
			fmt.Sprintf("same folder as each note  (attachmentFolderPath=%q)", p)

	case strings.HasPrefix(p, "./"):
		name := p[2:]
		return injector.ModeSubfolder, name,
			fmt.Sprintf("%s/ relative to each note  (attachmentFolderPath=%q)", name, p)

	default:
		// Plain name with no prefix — treated as fixed folder at vault root.
		return injector.ModeFixedFolder, p,
			fmt.Sprintf("%s/  (attachmentFolderPath=%q)", p, p)
	}
}

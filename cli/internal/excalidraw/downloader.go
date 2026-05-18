package excalidraw

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const maxDownloadBytes = 50 * 1024 * 1024

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Download fetches the Excalidraw scene, decrypts it, and saves it to destDir.
// Returns the absolute path of the saved .excalidraw file and whether it already
// existed on disk (cached). For #room= links, connects via Socket.IO WebSocket.
//
// When force is true the cached check is skipped and the file is always
// re-downloaded. The new content is written to a temporary file first; the
// existing file is replaced only on success, so a failed refresh never
// corrupts a previously good copy.
func Download(link Link, destDir string, force bool) (path string, cached bool, err error) {
	if link.Kind == Room {
		return DownloadRoom(link, destDir, force)
	}

	destFile := filepath.Join(destDir, "excalidraw-"+link.ID+".excalidraw")

	if !force {
		if _, err := os.Stat(destFile); err == nil {
			return destFile, true, nil // already downloaded
		}
	}

	resp, err := httpClient.Get(link.APIURL())
	if err != nil {
		return "", false, fmt.Errorf("fetch %s: %w", link.APIURL(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("server returned HTTP %d for %s", resp.StatusCode, link.APIURL())
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes))
	if err != nil {
		return "", false, fmt.Errorf("read response body: %w", err)
	}

	plaintext, err := Decrypt(link.Key, data)
	if err != nil {
		return "", false, fmt.Errorf("decrypt scene: %w", err)
	}

	writeTarget := destFile
	if force {
		writeTarget = destFile + ".tmp"
	}

	if err := os.WriteFile(writeTarget, plaintext, 0644); err != nil {
		return "", false, fmt.Errorf("write file: %w", err)
	}

	if force {
		if err := os.Rename(writeTarget, destFile); err != nil {
			os.Remove(writeTarget)
			return "", false, fmt.Errorf("replace file: %w", err)
		}
	}

	return destFile, false, nil
}

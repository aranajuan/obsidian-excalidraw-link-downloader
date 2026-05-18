package excalidraw

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const roomWSURL = "wss://oss-collab.excalidraw.com/socket.io/?EIO=4&transport=websocket"

// DownloadRoom connects to the Excalidraw room server via Socket.IO WebSocket.
//
// Flow when room has active participants:
//
//	downloader joins → server sends "new-user" to existing client → client broadcasts scene → we receive it
//
// Flow when room is empty (data lives in browser localStorage):
//
//	downloader joins → gets "first-in-room" → opens the room URL in the default browser →
//	browser loads, restores from localStorage, connects → server sends "new-user" to browser →
//	browser broadcasts scene → we receive it
func DownloadRoom(link Link, destDir string, force bool) (path string, cached bool, err error) {
	destFile := filepath.Join(destDir, "excalidraw-"+link.ID+".excalidraw")
	if !force {
		if _, err := os.Stat(destFile); err == nil {
			return destFile, true, nil
		}
	}

	conn, _, err := websocket.DefaultDialer.Dial(
		roomWSURL,
		http.Header{"Origin": []string{"https://excalidraw.com"}},
	)
	if err != nil {
		return "", false, fmt.Errorf("connect to room server: %w", err)
	}
	defer conn.Close()

	// Initial deadline covers the case where another participant is already in
	// the room and responds quickly.
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Engine.IO v4: receive OPEN packet "0{...}"
	if _, msg, err := conn.ReadMessage(); err != nil {
		return "", false, fmt.Errorf("read EIO open: %w", err)
	} else if len(msg) == 0 || msg[0] != '0' {
		return "", false, fmt.Errorf("expected EIO open, got: %s", msg)
	}

	// Socket.IO v4: connect to default namespace
	if err := conn.WriteMessage(websocket.TextMessage, []byte("40")); err != nil {
		return "", false, fmt.Errorf("send SIO connect: %w", err)
	}
	if err := waitForSIOConnect(conn); err != nil {
		return "", false, err
	}

	// Join the room
	joinMsg := fmt.Sprintf(`42["join-room","%s"]`, link.ID)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(joinMsg)); err != nil {
		return "", false, fmt.Errorf("send join-room: %w", err)
	}

	// onFirstInRoom is called when we are alone in the room.
	// The drawing lives in the browser's localStorage, so we open the room
	// URL in the default browser. When it loads, it restores from localStorage
	// and re-broadcasts the scene to us via the room server.
	onFirstInRoom := func() {
		openInBrowser(link.URL)
	}

	ciphertext, iv, browserOpened, err := readUntilBroadcast(conn, onFirstInRoom)
	if err != nil {
		return "", false, err
	}

	plaintext, err := DecryptGCM(link.Key, ciphertext, iv)
	if err != nil {
		return "", false, fmt.Errorf("decrypt: %w", err)
	}

	fileData, err := sceneToExcalidrawFile(plaintext)
	// When the browser was just opened it may broadcast empty scenes while
	// React initializes and IndexedDB loads. Give it a full minute from now.
	if errors.Is(err, ErrEmptyCanvas) && browserOpened {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
	for errors.Is(err, ErrEmptyCanvas) && browserOpened {
		ciphertext, iv, _, err = readUntilBroadcast(conn, nil)
		if err != nil {
			return "", false, err
		}
		plaintext, err = DecryptGCM(link.Key, ciphertext, iv)
		if err != nil {
			return "", false, fmt.Errorf("decrypt: %w", err)
		}
		fileData, err = sceneToExcalidrawFile(plaintext)
	}
	if err != nil {
		return "", false, fmt.Errorf("build .excalidraw: %w", err)
	}

	writeTarget := destFile
	if force {
		writeTarget = destFile + ".tmp"
	}

	if err := os.WriteFile(writeTarget, fileData, 0644); err != nil {
		return "", false, fmt.Errorf("write file: %w", err)
	}

	if force {
		if err := os.Rename(writeTarget, destFile); err != nil {
			os.Remove(writeTarget)
			return "", false, fmt.Errorf("replace file: %w", err)
		}
	}

	if browserOpened {
		closeBrowserTab(link.URL)
	}

	return destFile, false, nil
}

// openInBrowser opens url in the system's default browser without blocking.
func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		return
	}
	cmd.Start() // fire-and-forget
}

// closeBrowserTab closes any open browser tab whose URL contains the excalidraw.com
// domain. Called after a successful room download to clean up the tab that was
// auto-opened to re-broadcast the scene from localStorage.
func closeBrowserTab(_ string) {
	if runtime.GOOS != "darwin" {
		log.Println("Note: Browser tab could not be closed automatically on this OS. Please close it manually.")
		return
	}

	// Chrome-compatible browsers share the same AppleScript API.
	for _, app := range []string{"Google Chrome", "Chromium", "Microsoft Edge", "Brave Browser"} {
		script := fmt.Sprintf(`tell application "%s"
	if it is running then
		repeat with w in windows
			repeat with t in tabs of w
				try
					if URL of t contains "excalidraw.com" then close t
				end try
			end repeat
		end repeat
	end if
end tell`, app)
		exec.Command("osascript", "-e", script).Run()
	}

	// Arc has its own AppleScript dictionary.
	exec.Command("osascript", "-e", `tell application "Arc"
	if it is running then
		repeat with w in windows
			repeat with t in tabs of w
				try
					if URL of t contains "excalidraw.com" then close t
				end try
			end repeat
		end repeat
	end if
end tell`).Run()

	// Safari uses a different object model.
	exec.Command("osascript", "-e", `tell application "Safari"
	if it is running then
		repeat with w in windows
			repeat with t in tabs of w
				try
					if URL of t contains "excalidraw.com" then close t
				end try
			end repeat
		end repeat
	end if
end tell`).Run()
}

// waitForSIOConnect reads messages until it finds the Socket.IO connect ack "40{...}".
func waitForSIOConnect(conn *websocket.Conn) error {
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read SIO connect ack: %w", err)
		}
		if mt != websocket.TextMessage {
			continue
		}
		text := string(msg)
		if text == "2" {
			conn.WriteMessage(websocket.TextMessage, []byte("3"))
			continue
		}
		if strings.HasPrefix(text, "40") {
			return nil
		}
	}
}

// readUntilBroadcast reads Socket.IO messages until it receives a "client-broadcast"
// binary event, returning (ciphertext, iv, firstInRoomCalled).
//
// Socket.IO binary event wire format:
//
//	Text frame:   "45N-["event",{placeholder0},...,{placeholderN-1}]"  (N = attachment count)
//	Binary frame: one WebSocket binary message per attachment, in order
//
// For "client-broadcast": N=2, attachment[0]=ciphertext, attachment[1]=iv
//
// onFirstInRoom is called once when the server signals we are the only client in the room.
// firstInRoomCalled reports whether that callback was triggered during this call.
func readUntilBroadcast(conn *websocket.Conn, onFirstInRoom func()) (ciphertext, iv []byte, firstInRoomCalled bool, err error) {
	var pendingCount int
	var binFrames [][]byte
	var accBytes int
	var isBroadcast bool
	firstInRoomHandled := false

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return nil, nil, firstInRoomHandled, fmt.Errorf("read: %w", err)
		}

		if mt == websocket.BinaryMessage {
			accBytes += len(data)
			if accBytes > maxDownloadBytes {
				return nil, nil, firstInRoomHandled, fmt.Errorf("accumulated broadcast data exceeds 50MB limit")
			}
			binFrames = append(binFrames, data)
			if len(binFrames) == pendingCount {
				if isBroadcast && len(binFrames) >= 2 {
					return binFrames[0], binFrames[1], firstInRoomHandled, nil
				}
				pendingCount = 0
				binFrames = nil
				accBytes = 0
				isBroadcast = false
			}
			continue
		}

		text := string(data)

		// EIO ping → pong
		if text == "2" {
			conn.WriteMessage(websocket.TextMessage, []byte("3"))
			continue
		}

		// SIO binary event header: "45N-[...]"
		if len(text) >= 4 && text[:2] == "45" {
			dashIdx := strings.Index(text, "-")
			if dashIdx > 2 {
				n, e := strconv.Atoi(text[2:dashIdx])
				if e == nil && n > 0 {
					pendingCount = n
					var parts []string
					if err := json.Unmarshal([]byte(text[dashIdx+1:]), &parts); err == nil && len(parts) > 0 && parts[0] == "client-broadcast" {
						isBroadcast = true
					} else {
						isBroadcast = false
					}
				binFrames = nil
				accBytes = 0
				continue
				}
			}
		}

		// SIO text event: "42[...]"
		if strings.HasPrefix(text, "42") {
			if strings.Contains(text, `"first-in-room"`) && !firstInRoomHandled {
				firstInRoomHandled = true
				if onFirstInRoom != nil {
					onFirstInRoom() // open in browser; keep waiting for the broadcast
				}
			}
			// Ignore room-user-change and other events
		}
	}
}

// sceneData is the JSON structure of a decrypted Excalidraw room broadcast.
type sceneData struct {
	Type    string `json:"type"`
	Payload struct {
		Elements []json.RawMessage `json:"elements"`
	} `json:"payload"`
}

// ErrEmptyCanvas is returned when the room broadcast contains no elements.
// This means the current browser has no saved data for this room — the drawing
// was likely created on another participant's machine. Re-running while that
// participant has the room open will succeed.
var ErrEmptyCanvas = fmt.Errorf("room broadcast has no elements — data may only exist on another participant's browser")

// sceneToExcalidrawFile converts the decrypted room broadcast JSON into the
// standard .excalidraw file format that Obsidian's Excalidraw plugin can open.
// Returns ErrEmptyCanvas if the scene has no elements.
func sceneToExcalidrawFile(data []byte) ([]byte, error) {
	var scene sceneData
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("unmarshal scene: %w", err)
	}

	if len(scene.Payload.Elements) == 0 {
		return nil, ErrEmptyCanvas
	}

	file := map[string]any{
		"type":     "excalidraw",
		"version":  2,
		"source":   "https://excalidraw.com",
		"elements": scene.Payload.Elements,
		"appState": map[string]any{
			"viewBackgroundColor": "#ffffff",
		},
		"files": map[string]any{},
	}

	return json.MarshalIndent(file, "", "  ")
}

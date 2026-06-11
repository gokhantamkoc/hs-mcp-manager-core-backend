package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gorilla/websocket"
)

const WorkspaceRoot = "/workspace"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Relaxed for local LAN testing
}

type CloneRequest struct {
	RepoURL string `json:"repo_url"`
	Name    string `json:"name"`
}

type FilePayload struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func main() {
	if err := os.MkdirAll(WorkspaceRoot, 0755); err != nil {
		log.Fatalf("Failed to initialize workspace: %v", err)
	}

	// REST Endpoints
	http.HandleFunc("/api/workspace/clone", handleClone)
	http.HandleFunc("/api/workspace/files", handleListFiles)
	http.HandleFunc("/api/workspace/file/read", handleReadFile)
	http.HandleFunc("/api/workspace/file/save", handleSaveFile)

	// Live Streaming WebSocket Playground Endpoint
	http.HandleFunc("/api/mcp/ws-test", handleLiveTestMCP)

	log.Println("MCP Orchestrator Backend running securely on :8080...")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux)))
}

// Git Cloner
func handleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req CloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	targetPath := filepath.Join(WorkspaceRoot, req.Name)
	cmd := exec.Command("git", "clone", req.RepoURL, targetPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, fmt.Sprintf("Git clone error: %s", string(out)), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "cloned"})
}

// Tree View Builder for Monaco Editor
func handleListFiles(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	searchDir := filepath.Join(WorkspaceRoot, repo)
	var files []string
	_ = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Base(path)[0] != '.' {
			rel, _ := filepath.Rel(searchDir, path)
			files = append(files, rel)
		}
		return nil
	})
	json.NewEncoder(w).Encode(files)
}

// File Reader
func handleReadFile(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	relPath := r.URL.Query().Get("path")
	content, err := os.ReadFile(filepath.Join(WorkspaceRoot, repo, relPath))
	if err != nil {
		http.Error(w, "File read error", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(FilePayload{Path: relPath, Content: string(content)})
}

// File Writer
func handleSaveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	repo := r.URL.Query().Get("repo")
	var payload FilePayload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	_ = os.WriteFile(filepath.Join(WorkspaceRoot, repo, payload.Path), []byte(payload.Content), 0644)
	w.WriteHeader(http.StatusOK)
}

// Interactive Subprocess WebSocket Broker
func handleLiveTestMCP(w http.ResponseWriter, r *http.Request) {
	repoName := r.URL.Query().Get("repo")
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer wsConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = filepath.Join(WorkspaceRoot, repoName)

	stdinPipe, _ := cmd.StdinPipe()
	stdoutPipe, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		_ = wsConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Subprocess launch failed"}`))
		return
	}

	cmdDone := make(chan error, 1)
	go func() { cmdDone <- cmd.Wait() }()

	// Process Outbound (MCP Stdout -> Browser WebSocket)
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			_ = wsConn.WriteMessage(websocket.TextMessage, scanner.Bytes())
		}
	}()

	// Process Inbound (Browser WebSocket -> MCP Stdin)
	go func() {
		for {
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				cancel()
				return
			}
			_, _ = stdinPipe.Write(append(msg, '\n'))
		}
	}()

	select {
	case <-ctx.Done():
	case <-cmdDone:
		_ = wsConn.WriteMessage(websocket.TextMessage, []byte(`{"status":"process_terminated"}`))
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(w, r)
	})
}
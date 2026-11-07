package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) checkAuth(r *http.Request) bool {
	if s.token == "" {
		return true
	}
	return r.Header.Get("X-API-Token") == s.token
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Query().Get("path")
	target, err := s.resolvePath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, err := os.Stat(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if !info.IsDir() {
		http.Error(w, "not a directory", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			log.Printf("read dir info: %v", err)
			continue
		}
		result = append(result, DirEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  fi.Size(),
			Mtime: fi.ModTime(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"path":    path,
		"entries": result,
	})
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	target, err := s.resolvePath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := os.ReadFile(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data)
}

func (s *Server) handleWrite(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	target, err := s.resolvePath(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(target, []byte(req.Content), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = s.broadcast(map[string]any{
		"type": "write",
		"path": req.Path,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = struct{}{}
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		_ = conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var payload map[string]any
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}

		kind, _ := payload["type"].(string)
		if strings.EqualFold(kind, "hello") {
			_ = conn.WriteJSON(map[string]any{
				"type":   "hello_ok",
				"server": "elmer-syncd",
				"root":   s.root,
			})
		}
	}
}

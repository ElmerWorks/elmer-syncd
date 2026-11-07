package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type DirEntry struct {
	Name  string    `json:"name"`
	IsDir bool      `json:"isDir"`
	Size  int64     `json:"size"`
	Mtime time.Time `json:"mtime"`
}

type Server struct {
	root   string
	token  string
	router *chi.Mux

	upgrader websocket.Upgrader

	clientsMu sync.Mutex
	clients   map[*websocket.Conn]struct{}

	watcher *fsnotify.Watcher

	timersMu sync.Mutex
	timers   map[string]*time.Timer
}

func New(root, token string) (*Server, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	s := &Server{
		root:    absRoot,
		token:   token,
		router:  chi.NewRouter(),
		clients: make(map[*websocket.Conn]struct{}),
		timers:  make(map[string]*time.Timer),
		watcher: w,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	s.routes()
	s.addRecursiveWatch(absRoot)

	go s.watchLoop()

	return s, nil
}

func (s *Server) routes() {
	s.router.Use(s.authMiddleware)

	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("elmer-syncd ok"))
	})

	s.router.Get("/api/list", s.handleList)
	s.router.Get("/api/read", s.handleRead)
	s.router.Post("/api/write", s.handleWrite)
	s.router.Get("/ws", s.handleWS)
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Header.Get("X-API-Token") == s.token {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func (s *Server) resolvePath(p string) (string, error) {
	if p == "" {
		return s.root, nil
	}

	clean := filepath.Clean(p)
	if clean == "." {
		return s.root, nil
	}

	abs := clean
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(s.root, clean)
	}

	rootAbs, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return "", err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("path outside root")
	}

	return filepath.Clean(abs), nil
}

func (s *Server) broadcast(v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			_ = conn.Close()
			delete(s.clients, conn)
		}
	}

	return nil
}

func (s *Server) addRecursiveWatch(root string) {
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if err := s.watcher.Add(path); err != nil {
				log.Printf("watch add %s: %v", path, err)
			}
		}
		return nil
	})
}

func (s *Server) watchLoop() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				s.debounceEvent(event.Name)
			}

		case err, ok := <-s.watcher.Errors:
			if ok && err != nil {
				log.Printf("watch error: %v", err)
			}
		}
	}
}

func (s *Server) debounceEvent(path string) {
	s.timersMu.Lock()
	if t, ok := s.timers[path]; ok {
		t.Stop()
	}
	timer := time.AfterFunc(250*time.Millisecond, func() {
		s.broadcast(map[string]any{
			"type": "fs_event",
			"path": path,
		})
		s.timersMu.Lock()
		delete(s.timers, path)
		s.timersMu.Unlock()
	})
	s.timers[path] = timer
	s.timersMu.Unlock()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.watcher != nil {
		_ = s.watcher.Close()
	}
	return nil
}

// Package remote provides remote debugging support for DevTools.
package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// SimpleServer provides a simple HTTP server for remote debugging.
// For full WebSocket support, integrate gorilla/websocket or similar.
type SimpleServer struct {
	mu      sync.RWMutex
	bridge  *ChromiumBridge
	clients map[string]bool
}

// NewSimpleServer creates a new simple server.
func NewSimpleServer(bridge *ChromiumBridge) *SimpleServer {
	return &SimpleServer{
		bridge:  bridge,
		clients: make(map[string]bool),
	}
}

// Start starts the HTTP server on the given address.
func (s *SimpleServer) Start(addr string) error {
	mux := http.NewServeMux()

	// Serve inspector HTML page
	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, s.bridge.GetInspectorHTML())
	})

	// WebSocket endpoint placeholder
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "WebSocket endpoint - requires gorilla/websocket",
			"message": "Install: go get github.com/gorilla/websocket",
		})
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"server": "mint-devtools",
		})
	})

	// API endpoint to get bridge stats
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := s.bridge.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	fmt.Printf("Server listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// Broadcast sends a message to all connected clients.
func (s *SimpleServer) Broadcast(msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// In simple mode, just log the message
	log.Printf("Broadcast: %+v", msg)
}

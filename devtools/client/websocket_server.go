// Package client provides WebSocket server for Web Dashboard.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// WebSocketServer handles WebSocket connections for the dashboard.
type WebSocketServer struct {
	mu      sync.RWMutex
	clients map[string]*websocket.Conn
	running bool
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(port int) *WebSocketServer {
	_ = port // Port is managed by the WebDashboard
	return &WebSocketServer{
		clients: make(map[string]*websocket.Conn),
		running: false,
	}
}

// Start marks the WebSocket server as running.
func (ws *WebSocketServer) Start() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.running {
		return nil
	}

	ws.running = true
	log.Printf("[WebSocket] Server handler ready")
	return nil
}

// Stop stops the WebSocket server.
func (ws *WebSocketServer) Stop() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.running {
		return nil
	}

	ws.running = false

	// Close all client connections
	for clientID, conn := range ws.clients {
		conn.Close()
		log.Printf("[WebSocket] Closed client: %s", clientID)
	}
	ws.clients = make(map[string]*websocket.Conn)

	return nil
}

// Handler returns the WebSocket handler for use with http.ServeMux.
func (ws *WebSocketServer) Handler() http.Handler {
	return websocket.Handler(ws.handleConnection)
}

// handleConnection handles an individual WebSocket connection.
func (ws *WebSocketServer) handleConnection(conn *websocket.Conn) {
	clientID := conn.Request().RemoteAddr
	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}

	ws.mu.Lock()
	ws.clients[clientID] = conn
	ws.mu.Unlock()

	log.Printf("[WebSocket] Client connected: %s", clientID)

	// Send welcome message
	welcome := map[string]interface{}{
		"type":      "welcome",
		"client_id": clientID,
		"timestamp": time.Now().Format(time.RFC3339),
		"server":    "mint-webdashboard",
	}
	_ = websocket.JSON.Send(conn, welcome)

	// Message loop
	var msg json.RawMessage
	for {
		err := websocket.JSON.Receive(conn, &msg)
		if err != nil {
			if err != io.EOF {
				log.Printf("[WebSocket] Receive error from %s: %v", clientID, err)
			}
			break
		}

		// Parse message
		var baseMsg map[string]interface{}
		if err := json.Unmarshal(msg, &baseMsg); err != nil {
			log.Printf("[WebSocket] Parse error from %s: %v", clientID, err)
			continue
		}

		// Handle message
		if msgType, ok := baseMsg["type"].(string); ok {
			ws.handleMessage(clientID, conn, msgType, baseMsg)
		}
	}

	// Cleanup
	ws.mu.Lock()
	delete(ws.clients, clientID)
	ws.mu.Unlock()

	conn.Close()
	log.Printf("[WebSocket] Client disconnected: %s", clientID)
}

// handleMessage handles a WebSocket message.
func (ws *WebSocketServer) handleMessage(clientID string, conn *websocket.Conn, msgType string, msg map[string]interface{}) {
	switch msgType {
	case "ping":
		websocket.JSON.Send(conn, map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Format(time.RFC3339),
		})

	case "subscribe":
		log.Printf("[WebSocket] Client %s subscribed", clientID)

	case "get_frames":
		websocket.JSON.Send(conn, map[string]interface{}{
			"type":    "error",
			"message": "Use HTTP API for frame data",
		})

	default:
		log.Printf("[WebSocket] Unknown message type: %s from %s", msgType, clientID)
	}
}

// Broadcast sends a message to all connected clients.
func (ws *WebSocketServer) Broadcast(msg interface{}) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for clientID, conn := range ws.clients {
		if err := websocket.JSON.Send(conn, msg); err != nil {
			log.Printf("[WebSocket] Broadcast error to %s: %v", clientID, err)
		}
	}
}

// BroadcastUpdate broadcasts a typed update to all clients.
func (ws *WebSocketServer) BroadcastUpdate(updateType string, data interface{}) {
	msg := map[string]interface{}{
		"type":      updateType,
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	ws.Broadcast(msg)
}

// GetClientCount returns the number of connected clients.
func (ws *WebSocketServer) GetClientCount() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return len(ws.clients)
}

// GetClientIDs returns all client IDs.
func (ws *WebSocketServer) GetClientIDs() []string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	ids := make([]string, 0, len(ws.clients))
	for id := range ws.clients {
		ids = append(ids, id)
	}
	return ids
}

// IsRunning returns whether the server is running.
func (ws *WebSocketServer) IsRunning() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.running
}

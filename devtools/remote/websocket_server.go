// Package remote provides remote debugging support for DevTools.
package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/snapshot"
	"golang.org/x/net/websocket"
)

// WebSocketServer handles WebSocket connections for remote debugging.
type WebSocketServer struct {
	mu      sync.RWMutex
	bridge  *ChromiumBridge
	clients map[*websocket.Conn]string
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(bridge *ChromiumBridge) *WebSocketServer {
	return &WebSocketServer{
		bridge:  bridge,
		clients: make(map[*websocket.Conn]string),
	}
}

// Handler returns the WebSocket handler for use with http.ServeMux.
func (s *WebSocketServer) Handler() http.Handler {
	return websocket.Handler(s.handleConnection)
}

// handleConnection handles an individual WebSocket connection.
func (s *WebSocketServer) handleConnection(ws *websocket.Conn) {
	// Generate client ID
	clientID := fmt.Sprintf("client-%d", len(s.clients))

	s.mu.Lock()
	s.clients[ws] = clientID
	s.mu.Unlock()

	log.Printf("WebSocket client connected: %s", clientID)
	defer func() {
		s.mu.Lock()
		delete(s.clients, ws)
		s.mu.Unlock()
		ws.Close()
		log.Printf("WebSocket client disconnected: %s", clientID)
	}()

	// Send handshake acknowledgment
	handshakeAck := map[string]interface{}{
		"version":    ProtocolVersion,
		"type":       "handshake_ack",
		"server_id":  "mint-devtools",
		"session_id": clientID,
	}
	s.send(ws, handshakeAck)

	// Message loop
	var msg json.RawMessage
	for {
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			if err != io.EOF {
				log.Printf("WebSocket receive error: %v", err)
			}
			break
		}

		// Parse message
		var baseMsg Message
		if err := json.Unmarshal(msg, &baseMsg); err != nil {
			log.Printf("Parse error: %v", err)
			// Send error response
			s.sendError(ws, "", fmt.Sprintf("Parse error: %v", err))
			continue
		}

		// Create session wrapper
		session := s.bridge.server.CreateSession(clientID)

		// Handle message
		response := s.bridge.handleMessage(session, &baseMsg)
		if response != nil {
			// Log response details for debugging
			log.Printf("WebSocket sending: type=%s, id=%s, payload=%T", response.Type, response.ID, response.Payload)
			s.send(ws, response)
		} else {
			log.Printf("WebSocket: no response for message type=%s", baseMsg.Type)
		}
	}
}

// send sends a message to the WebSocket client.
func (s *WebSocketServer) send(ws *websocket.Conn, msg interface{}) error {
	return websocket.JSON.Send(ws, msg)
}

// sendError sends an error message.
func (s *WebSocketServer) sendError(ws *websocket.Conn, msgID string, errMsg string) {
	errorMsg := map[string]interface{}{
		"version": ProtocolVersion,
		"type":    "error",
		"id":      msgID,
		"error":   errMsg,
	}
	s.send(ws, errorMsg)
}

// Broadcast sends a message to all connected clients.
func (s *WebSocketServer) Broadcast(msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ws := range s.clients {
		if err := s.send(ws, msg); err != nil {
			log.Printf("Broadcast error: %v", err)
		}
	}
}

// BroadcastEvent broadcasts an event to all subscribed clients.
func (s *WebSocketServer) BroadcastEvent(event *EventPayload) {
	s.Broadcast(map[string]interface{}{
		"version": ProtocolVersion,
		"type":    "event",
		"payload": event,
	})
}

// BroadcastSnapshot broadcasts a snapshot to all clients.
func (s *WebSocketServer) BroadcastSnapshot(snap *snapshot.Snapshot) {
	// Convert snapshot to remote format
	components := make([]ComponentData, 0, len(snap.States))
	for _, state := range snap.States {
		comp := ComponentData{
			NodeID:  state.NodeID,
			Type:    state.Type,
			Props:   state.Props,
			State:   state.State,
			Bounds: RectData{
				X:      state.Bounds.X,
				Y:      state.Bounds.Y,
				Width:  state.Bounds.Width,
				Height: state.Bounds.Height,
			},
			Children: state.Children,
			Visible:  state.Visible,
			Focused:  state.Focused,
		}
		components = append(components, comp)
	}

	payload := SnapshotPayload{
		FrameID:   snap.FrameID,
		Timestamp: snap.Timestamp,
		WindowState: WindowState{
			Width:  snap.Global.WindowSize.Width,
			Height: snap.Global.WindowSize.Height,
		},
		Components: components,
	}

	s.Broadcast(map[string]interface{}{
		"version": ProtocolVersion,
		"type":    "snapshot",
		"payload": payload,
	})
}

// GetClientCount returns the number of connected clients.
func (s *WebSocketServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// =============================================================================
// Integrated Server (HTTP + WebSocket)
// =============================================================================

// DevToolsServer combines HTTP and WebSocket serving.
type DevToolsServer struct {
	mu              sync.RWMutex
	wsServer       *WebSocketServer
	httpServeMux    *http.ServeMux
	bridge          *ChromiumBridge
	snapshotManager *snapshot.Manager
	port            int
}

// NewDevToolsServer creates a new DevTools server.
func NewDevToolsServer(port int, dt *devtools.DevTools, sm *snapshot.Manager) *DevToolsServer {
	bridge := NewChromiumBridge(dt, sm)
	bridge.Enable()

	wsServer := NewWebSocketServer(bridge)

	server := &DevToolsServer{
		wsServer:        wsServer,
		httpServeMux:     http.NewServeMux(),
		bridge:          bridge,
		snapshotManager: sm,
		port:             port,
	}

	// Setup HTTP routes
	server.setupRoutes()

	return server
}

// setupRoutes configures HTTP routes.
func (s *DevToolsServer) setupRoutes() {
	// Serve inspector HTML page at /debug
	s.httpServeMux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, s.bridge.GetInspectorHTML())
	})

	// WebSocket endpoint at /ws
	s.httpServeMux.Handle("/ws", s.wsServer.Handler())

	// Health check
	s.httpServeMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "ok",
			"server":       "mint-devtools",
			"version":      ProtocolVersion,
			"ws_clients":   s.wsServer.GetClientCount(),
			"snapshots":     s.snapshotManager.GetStats().TotalSnapshots,
		})
	})

	// API: Get all snapshots
	s.httpServeMux.HandleFunc("/api/snapshots", func(w http.ResponseWriter, r *http.Request) {
		snapshots := s.snapshotManager.GetAll()
		w.Header().Set("Content-Type", "application/json")

		result := make([]map[string]interface{}, 0, len(snapshots))
		for _, snap := range snapshots {
			result = append(result, map[string]interface{}{
				"id":         string(snap.ID),
				"frame_id":   int(snap.FrameID),
				"timestamp":  snap.Timestamp.Format(time.RFC3339),
				"components": len(snap.States),
			})
		}
		json.NewEncoder(w).Encode(result)
	})

	// API: Get specific snapshot
	s.httpServeMux.HandleFunc("/api/snapshot/", func(w http.ResponseWriter, r *http.Request) {
		// Extract frame ID from path
		// /api/snapshot/123
		// For simplicity, use query parameter instead
		frameStr := r.URL.Query().Get("frame")
		if frameStr == "" {
			http.Error(w, "Missing frame parameter", 400)
			return
		}

		var frameID int
		if _, err := fmt.Sscanf(frameStr, "%d", &frameID); err != nil {
			http.Error(w, "Invalid frame ID", 400)
			return
		}

		snap, ok := s.snapshotManager.Get(devtools.FrameID(frameID))
		if !ok {
			http.Error(w, "Snapshot not found", 404)
			return
		}

		// Convert to JSON-serializable format
		components := make([]ComponentData, 0, len(snap.States))
		for _, state := range snap.States {
			comp := ComponentData{
				NodeID:  state.NodeID,
				Type:    state.Type,
				Props:   state.Props,
				State:   state.State,
				Bounds: RectData{
					X:      state.Bounds.X,
					Y:      state.Bounds.Y,
					Width:  state.Bounds.Width,
					Height: state.Bounds.Height,
				},
				Children: state.Children,
				Visible:  state.Visible,
				Focused:  state.Focused,
			}
			components = append(components, comp)
		}

		result := SnapshotPayload{
			FrameID:   snap.FrameID,
			Timestamp: snap.Timestamp,
			WindowState: WindowState{
				Width:  snap.Global.WindowSize.Width,
				Height: snap.Global.WindowSize.Height,
			},
			Components: components,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// API: Diff two snapshots
	s.httpServeMux.HandleFunc("/api/diff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")

		if fromStr == "" || toStr == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing from/to parameters"})
			return
		}

		// Support both frame IDs and snapshot IDs
		var fromSnap, toSnap *snapshot.Snapshot

		// Try as frame IDs first
		var fromInt, toInt int
		_, fromErr := fmt.Sscanf(fromStr, "%d", &fromInt)
		_, toErr := fmt.Sscanf(toStr, "%d", &toInt)

		if fromErr == nil && toErr == nil {
			// Use frame IDs
			fromSnap, _ = s.snapshotManager.Get(devtools.FrameID(fromInt))
			toSnap, _ = s.snapshotManager.Get(devtools.FrameID(toInt))
		} else {
			// Try as snapshot IDs - not implemented, return error
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "Snapshot ID lookup not implemented, use frame IDs"})
			return
		}

		if fromSnap == nil || toSnap == nil {
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Snapshot not found",
				"available_frames": getAvailableFrameIDs(s.snapshotManager),
			})
			return
		}

		differ := snapshot.NewDiffer()
		diff := differ.Compare(fromSnap, toSnap)

		// Convert changes to JSON-serializable format
		changes := make([]ChangeData, 0, len(diff.Changes))
		for _, change := range diff.Changes {
			changes = append(changes, ChangeData{
				NodeID:   change.NodeID,
				Type:     change.ChangeType.String(),
				Path:     change.Path,
				OldValue: change.OldValue,
				NewValue: change.NewValue,
			})
		}

		result := DiffPayload{
			From:    devtools.FrameID(fromInt),
			To:      devtools.FrameID(toInt),
			Changes: changes,
		}

		json.NewEncoder(w).Encode(result)
	})
}

// Start starts the server on the configured port.
func (s *DevToolsServer) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("DevTools server starting on %s", addr)
	log.Printf("  Inspector: http://localhost%s/debug", addr)
	log.Printf("  WebSocket: ws://localhost%s/ws", addr)

	return http.ListenAndServe(addr, s.httpServeMux)
}

// StartInBackground starts the server in a goroutine.
func (s *DevToolsServer) StartInBackground() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("DevTools server starting on %s (background)", addr)

	go func() {
		if err := http.ListenAndServe(addr, s.httpServeMux); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	return nil
}

// GetStats returns server statistics.
func (s *DevToolsServer) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"port":         s.port,
		"ws_clients":   s.wsServer.GetClientCount(),
		"bridge_stats": s.bridge.GetStats(),
		"snapshots":    s.snapshotManager.GetStats(),
	}
}

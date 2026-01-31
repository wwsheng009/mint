// Package protocol provides unified WebSocket server for DevTools communication.
//
// This server supports both:
//   - TUI Debugging: Terminal-based debugging
//   - Remote Debugging: Web-based debugging with CDP compatibility
package protocol

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

// =============================================================================
// Message Handler
// =============================================================================

// MessageHandler handles incoming WebSocket messages.
type MessageHandler func(session *Session, msg *Message) *Message

// SnapshotHandler provides snapshot data for queries.
type SnapshotHandler func(frameID devtools.FrameID) (*snapshot.Snapshot, bool)

// RangeHandler provides a range of snapshots.
type RangeHandler func(from, to devtools.FrameID) []*snapshot.Snapshot

// =============================================================================
// WebSocket Server Configuration
// =============================================================================

// WebSocketServerConfig configures the WebSocket server.
type WebSocketServerConfig struct {
	Port              int
	Path              string // WebSocket path, e.g., "/ws"
	EnableCdp         bool   // Enable CDP compatibility
	EnableTuiCommands bool   // Enable TUI-specific commands
	HeartbeatInterval int    // Heartbeat interval in seconds
}

// DefaultWebSocketServerConfig returns default configuration.
func DefaultWebSocketServerConfig() WebSocketServerConfig {
	return WebSocketServerConfig{
		Port:              9222,
		Path:              "/ws",
		EnableCdp:         true,
		EnableTuiCommands: true,
		HeartbeatInterval: 30,
	}
}

// =============================================================================
// WebSocket Server
// =============================================================================

// WebSocketServer handles WebSocket connections for DevTools.
type WebSocketServer struct {
	mu              sync.RWMutex
	config          WebSocketServerConfig
	clients         map[*websocket.Conn]*ClientInfo
	sessionManager  *SessionManager
	messageHandler  MessageHandler
	snapshotHandler SnapshotHandler
	rangeHandler    RangeHandler
	running         bool
}

// ClientInfo contains information about a connected client.
type ClientInfo struct {
	ID          string
	Session     *Session
	ConnectedAt int64
	Protocol    string // "tui" or "remote"
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(config WebSocketServerConfig) *WebSocketServer {
	if config.Path == "" {
		config.Path = "/ws"
	}

	return &WebSocketServer{
		config:         config,
		clients:        make(map[*websocket.Conn]*ClientInfo),
		sessionManager: NewSessionManager(),
		running:        false,
	}
}

// SetMessageHandler sets the message handler.
func (s *WebSocketServer) SetMessageHandler(handler MessageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageHandler = handler
}

// SetSnapshotHandler sets the snapshot handler.
func (s *WebSocketServer) SetSnapshotHandler(handler SnapshotHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshotHandler = handler
}

// SetRangeHandler sets the range handler.
func (s *WebSocketServer) SetRangeHandler(handler RangeHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rangeHandler = handler
}

// Handler returns the WebSocket handler for use with http.ServeMux.
func (s *WebSocketServer) Handler() http.Handler {
	return websocket.Handler(s.handleConnection)
}

// handleConnection handles an individual WebSocket connection.
func (s *WebSocketServer) handleConnection(ws *websocket.Conn) {
	clientID := ws.Request().RemoteAddr
	if clientID == "" {
		clientID = fmt.Sprintf("client-%s", ws.Request().URL.RequestURI())
	}

	s.mu.Lock()
	s.running = true

	// Generate client info
	info := &ClientInfo{
		ID:          clientID,
		ConnectedAt: time.Now().Unix(), // placeholder for connection time
	}
	s.clients[ws] = info
	s.mu.Unlock()

	log.Printf("[WebSocket] Client connected: %s", clientID)

	defer func() {
		s.mu.Lock()
		delete(s.clients, ws)
		s.mu.Unlock()
		ws.Close()
		log.Printf("[WebSocket] Client disconnected: %s", clientID)
	}()

	// Send initial acknowledgment
	handshakeAck := map[string]interface{}{
		"version":    Version,
		"type":       string(TypeHandshakeAck),
		"server_id":  "mint-devtools",
		"session_id": clientID,
		"timestamp":  info.ConnectedAt,
	}
	s.send(ws, handshakeAck)

	// Message loop
	for {
		var msgJSON json.RawMessage
		err := websocket.JSON.Receive(ws, &msgJSON)
		if err != nil {
			if err != io.EOF {
				log.Printf("[WebSocket] Receive error from %s: %v", clientID, err)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(msgJSON, &msg); err != nil {
			log.Printf("[WebSocket] Parse error from %s: %v", clientID, err)
			s.sendError(ws, "", fmt.Sprintf("Parse error: %v", err))
			continue
		}

		// Ensure session exists
		if info.Session == nil {
			if msg.Type == TypeHandshake {
				// Handle handshake to determine protocol
				info.Session = s.handleHandshake(clientID, &msg)
				if info.Session != nil {
					info.Protocol = info.Session.Protocol()
				}
			}
		}

		// Update activity
		if info.Session != nil {
			info.Session.UpdateActivity()
		}

		// Handle message
		response := s.handleMessage(info.Session, &msg)
		if response != nil {
			s.send(ws, response)
		}
	}
}

// handleHandshake handles the handshake message.
func (s *WebSocketServer) handleHandshake(clientID string, msg *Message) *Session {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil
	}

	// Determine protocol type
	protocol := "remote" // default
	if proto, ok := payload["protocol"].(string); ok {
		protocol = proto
	}

	// Create session
	session := s.sessionManager.CreateSession(clientID, protocol)

	// Store capabilities if provided
	if caps, ok := payload["capabilities"].([]interface{}); ok {
		capStrs := make([]string, 0, len(caps))
		for _, c := range caps {
			if cs, ok := c.(string); ok {
				capStrs = append(capStrs, cs)
			}
		}
		log.Printf("[WebSocket] Client %s capabilities: %v", clientID, capStrs)
	}

	log.Printf("[WebSocket] Session created: %s (protocol: %s)", session.ID(), protocol)

	return session
}

// handleMessage handles an incoming message.
func (s *WebSocketServer) handleMessage(session *Session, msg *Message) *Message {
	// Check if TUI command is enabled
	if isTUICommand(msg.Type) && !s.config.EnableTuiCommands {
		return NewError(msg.ID, "TUI commands are disabled")
	}

	// Use custom handler if set
	if s.messageHandler != nil {
		return s.messageHandler(session, msg)
	}

	// Default message handling
	switch msg.Type {
	case TypeHandshake:
		return s.handleHandshakeAck(session, msg)
	case TypeGetSnapshot:
		return s.handleGetSnapshot(session, msg)
	case TypeGetRange:
		return s.handleGetRange(session, msg)
	case TypeGetDiff:
		return s.handleGetDiff(session, msg)
	case TypeSubscribe:
		return s.handleSubscribe(session, msg)
	case TypeUnsubscribe:
		return s.handleUnsubscribe(session, msg)
	case TypeHeartbeat:
		return s.handleHeartbeat(session, msg)
	case TypeInspect:
		return s.handleInspect(session, msg)
	case TypeHighlight:
		return s.handleHighlight(session, msg)
	case TypeGetFrame:
		return s.handleGetFrame(session, msg)
	case TypeGetTimeline:
		return s.handleGetTimeline(session, msg)
	default:
		return NewError(msg.ID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

// handleHandshakeAck handles handshake acknowledgment.
func (s *WebSocketServer) handleHandshakeAck(session *Session, msg *Message) *Message {
	if session == nil {
		return NewError(msg.ID, "session not created")
	}

	ack := HandshakeAckPayload{
		ServerID: "mint-devtools",
		Version:  Version,
		Capabilities: []string{
			"snapshots",
			"events",
			"diffs",
		},
		SessionID: session.ID(),
	}

	if s.config.EnableTuiCommands {
		ack.Capabilities = append(ack.Capabilities, "inspect", "highlight", "replay")
	}

	if s.config.EnableCdp {
		ack.Capabilities = append(ack.Capabilities, "cdp")
	}

	return NewMessageWithID(msg.ID, TypeHandshakeAck, ack)
}

// handleGetSnapshot handles snapshot requests.
func (s *WebSocketServer) handleGetSnapshot(session *Session, msg *Message) *Message {
	if s.snapshotHandler == nil {
		return NewError(msg.ID, "snapshot handler not configured")
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	frameIDFloat, _ := payload["frame_id"].(float64)
	frameID := devtools.FrameID(frameIDFloat)

	snap, exists := s.snapshotHandler(frameID)
	if !exists {
		return NewError(msg.ID, fmt.Sprintf("snapshot not found for frame %d", frameID))
	}

	// Convert to protocol format
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

	return NewMessageWithID(msg.ID, TypeSnapshot, result)
}

// handleGetRange handles range requests.
func (s *WebSocketServer) handleGetRange(session *Session, msg *Message) *Message {
	if s.rangeHandler == nil {
		return NewError(msg.ID, "range handler not configured")
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	fromFloat, _ := payload["from"].(float64)
	toFloat, _ := payload["to"].(float64)

	from := devtools.FrameID(fromFloat)
	to := devtools.FrameID(toFloat)

	snapshots := s.rangeHandler(from, to)

	frames := make([]FrameSummary, 0, len(snapshots))
	for _, snap := range snapshots {
		frames = append(frames, FrameSummary{
			FrameID:   snap.FrameID,
			Timestamp: snap.Timestamp,
			Events:    snap.Metadata.FramesSinceLast,
			Mutations: snap.Metadata.MutationsCount,
			Layouts:   snap.Metadata.LayoutsCount,
		})
	}

	result := RangePayload{Frames: frames}
	return NewMessageWithID(msg.ID, TypeGetRange, result)
}

// handleGetDiff handles diff requests.
func (s *WebSocketServer) handleGetDiff(session *Session, msg *Message) *Message {
	if s.snapshotHandler == nil {
		return NewError(msg.ID, "snapshot handler not configured")
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	fromFloat, _ := payload["from"].(float64)
	toFloat, _ := payload["to"].(float64)

	from := devtools.FrameID(fromFloat)
	to := devtools.FrameID(toFloat)

	fromSnap, fromOk := s.snapshotHandler(from)
	toSnap, toOk := s.snapshotHandler(to)

	if !fromOk || !toOk {
		return NewError(msg.ID, "snapshot not found for diff")
	}

	differ := snapshot.NewDiffer()
	diff := differ.Compare(fromSnap, toSnap)

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
		From:    from,
		To:      to,
		Changes: changes,
	}

	return NewMessageWithID(msg.ID, TypeDiff, result)
}

// handleSubscribe handles subscription requests.
func (s *WebSocketServer) handleSubscribe(session *Session, msg *Message) *Message {
	if session == nil {
		return NewError(msg.ID, "no session")
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	eventType, _ := payload["event_type"].(string)
	session.Subscribe(string(eventType))

	return NewMessageWithID(msg.ID, TypeResponse, map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"event_type": eventType,
			"status":     "subscribed",
		},
	})
}

// handleUnsubscribe handles unsubscribe requests.
func (s *WebSocketServer) handleUnsubscribe(session *Session, msg *Message) *Message {
	if session == nil {
		return NewError(msg.ID, "no session")
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	eventType, _ := payload["event_type"].(string)
	session.Unsubscribe(string(eventType))

	return NewMessageWithID(msg.ID, TypeResponse, map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"event_type": eventType,
			"status":     "unsubscribed",
		},
	})
}

// handleHeartbeat handles heartbeat messages.
func (s *WebSocketServer) handleHeartbeat(session *Session, msg *Message) *Message {
	if session != nil {
		session.UpdateActivity()
	}
	return NewMessage(TypeHeartbeat, nil)
}

// handleInspect handles TUI inspect commands.
func (s *WebSocketServer) handleInspect(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	nodeID, _ := payload["node_id"].(string)

	return NewMessageWithID(msg.ID, TypeResponse, map[string]interface{}{
		"success": true,
		"data": InspectResultPayload{
			NodeID: nodeID,
			Type:   "Container",
			State:  make(map[string]interface{}),
		},
	})
}

// handleHighlight handles TUI highlight commands.
func (s *WebSocketServer) handleHighlight(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	nodeID, _ := payload["node_id"].(string)
	color, _ := payload["color"].(string)

	return NewMessageWithID(msg.ID, TypeResponse, map[string]interface{}{
		"success": true,
		"data": HighlightResultPayload{
			NodeID: nodeID,
			Color:  color,
			Status: "highlighted",
		},
	})
}

// handleGetFrame handles frame data requests.
func (s *WebSocketServer) handleGetFrame(session *Session, msg *Message) *Message {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return NewError(msg.ID, "invalid payload")
	}

	frameIDFloat, _ := payload["frame_id"].(float64)
	frameID := devtools.FrameID(frameIDFloat)

	return NewMessageWithID(msg.ID, TypeResponse, map[string]interface{}{
		"success": true,
		"data": FramePayload{
			FrameID: frameID,
			Events:  []EventData{},
			State:   make(map[string]interface{}),
		},
	})
}

// handleGetTimeline handles timeline requests.
func (s *WebSocketServer) handleGetTimeline(session *Session, msg *Message) *Message {
	return NewMessageWithID(msg.ID, TypeResponse, map[string]interface{}{
		"success": true,
		"data": TimelinePayload{
			Frames: []FrameSummary{},
			Count:  0,
		},
	})
}

// send sends a message to the WebSocket client.
func (s *WebSocketServer) send(ws *websocket.Conn, msg interface{}) error {
	return websocket.JSON.Send(ws, msg)
}

// sendError sends an error message.
func (s *WebSocketServer) sendError(ws *websocket.Conn, msgID string, errMsg string) {
	errorMsg := map[string]interface{}{
		"version": Version,
		"type":    TypeError,
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
			log.Printf("[WebSocket] Broadcast error: %v", err)
		}
	}
}

// BroadcastEvent broadcasts an event to all subscribed sessions.
func (s *WebSocketServer) BroadcastEvent(event *EventPayload) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := NewMessage(TypeEvent, event)
	data, _ := msg.Serialize()

	for ws, info := range s.clients {
		if info.Session != nil && info.Session.IsSubscribed(string(TypeEvent)) {
			var rawMsg json.RawMessage
			copy(rawMsg[:], data[:])
			_ = websocket.JSON.Send(ws, rawMsg)
		}
	}
}

// BroadcastSnapshot broadcasts a snapshot to all subscribed sessions.
func (s *WebSocketServer) BroadcastSnapshot(snap *snapshot.Snapshot) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert to protocol format
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

	msg := NewMessage(TypeSnapshot, payload)

	for ws, info := range s.clients {
		if info.Session != nil && info.Session.IsSubscribed(string(TypeSnapshot)) {
			s.send(ws, msg)
		}
	}
}

// GetClientCount returns the number of connected clients.
func (s *WebSocketServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// GetSessionManager returns the session manager.
func (s *WebSocketServer) GetSessionManager() *SessionManager {
	return s.sessionManager
}

// IsRunning returns whether the server is running.
func (s *WebSocketServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// =============================================================================
// Utilities
// =============================================================================

// isTUICommand checks if a message type is a TUI-specific command.
func isTUICommand(msgType MessageType) bool {
	switch msgType {
	case TypeInspect, TypeHighlight, TypeReplayStart, TypeReplayStop, TypeGetFrame, TypeGetTimeline:
		return true
	default:
		return false
	}
}

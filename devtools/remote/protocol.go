// Package remote provides remote debugging support for DevTools.
package remote

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// =============================================================================
// Remote Debugging Protocol
// =============================================================================

// Protocol version
const ProtocolVersion = "1.0.0"

// Message types
const (
	// Client -> Server messages
	TypeHandshake      = "handshake"
	TypeSubscribe      = "subscribe"
	TypeUnsubscribe    = "unsubscribe"
	TypeGetSnapshot    = "get_snapshot"
	TypeGetRange       = "get_range"
	TypeGetDiff        = "get_diff"
	TypeSetBreakpoint  = "set_breakpoint"
	TypeClearBreakpoint = "clear_breakpoint"
	TypeEvaluate       = "evaluate"
	TypeResume         = "resume"
	TypeStep           = "step"

	// Server -> Client messages
	TypeHandshakeAck   = "handshake_ack"
	TypeEvent          = "event"
	TypeSnapshot       = "snapshot"
	TypeDiff           = "diff"
	TypeError          = "error"
	TypeBreakpointHit  = "breakpoint_hit"
	TypeEvaluationResult = "evaluation_result"
)

// Message is the base protocol message.
type Message struct {
	Version  string      `json:"version"`
	Type     string      `json:"type"`
	ID       string      `json:"id,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// =============================================================================
// Handshake
// =============================================================================

// HandshakePayload is the initial handshake payload.
type HandshakePayload struct {
	ClientID    string   `json:"client_id"`
	Capabilities []string `json:"capabilities"`
	Version     string   `json:"version"`
}

// HandshakeAckPayload is the handshake acknowledgment.
type HandshakeAckPayload struct {
	ServerID      string   `json:"server_id"`
	Version       string   `json:"version"`
	Capabilities  []string `json:"capabilities"`
	SessionID     string   `json:"session_id"`
}

// =============================================================================
// Event Messages
// =============================================================================

// EventPayload represents an event message.
type EventPayload struct {
	FrameID    devtools.FrameID `json:"frame_id"`
	EventType  string           `json:"event_type"`
	NodeID     devtools.NodeID   `json:"node_id"`
	Phase      string           `json:"phase"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time        `json:"timestamp"`
}

// =============================================================================
// Snapshot Messages
// =============================================================================

// GetSnapshotPayload is the payload for getting a snapshot.
type GetSnapshotPayload struct {
	FrameID devtools.FrameID `json:"frame_id"`
	IncludeState bool        `json:"include_state"`
	IncludeChildren bool     `json:"include_children"`
}

// SnapshotPayload contains a snapshot.
type SnapshotPayload struct {
	FrameID   devtools.FrameID `json:"frame_id"`
	Timestamp time.Time        `json:"timestamp"`
	WindowState WindowState    `json:"window_state"`
	Components []ComponentData `json:"components"`
}

// WindowState represents the window state.
type WindowState struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ComponentData represents component data for remote transfer.
type ComponentData struct {
	NodeID   devtools.NodeID            `json:"node_id"`
	Type     string                     `json:"type"`
	Props    map[string]interface{}     `json:"props,omitempty"`
	State    map[string]interface{}     `json:"state,omitempty"`
	Bounds   RectData                   `json:"bounds"`
	Children []devtools.NodeID          `json:"children,omitempty"`
	Visible  bool                       `json:"visible"`
	Focused  bool                       `json:"focused"`
}

// RectData represents rectangle data.
type RectData struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// =============================================================================
// Range Messages
// =============================================================================

// GetRangePayload is the payload for getting a range of frames.
type GetRangePayload struct {
	From devtools.FrameID `json:"from"`
	To   devtools.FrameID `json:"to"`
	Limit int             `json:"limit,omitempty"`
}

// RangePayload contains frame data.
type RangePayload struct {
	Frames []FrameSummary `json:"frames"`
}

// FrameSummary summarizes a frame.
type FrameSummary struct {
	FrameID   devtools.FrameID `json:"frame_id"`
	Timestamp time.Time        `json:"timestamp"`
	Events    int              `json:"events"`
	Mutations int             `json:"mutations"`
	Layouts   int              `json:"layouts"`
}

// =============================================================================
// Diff Messages
// =============================================================================

// GetDiffPayload is the payload for getting a diff.
type GetDiffPayload struct {
	From devtools.FrameID `json:"from"`
	To   devtools.FrameID `json:"to"`
}

// DiffPayload contains diff data.
type DiffPayload struct {
	From    devtools.FrameID `json:"from"`
	To      devtools.FrameID `json:"to"`
	Changes []ChangeData     `json:"changes"`
}

// ChangeData represents a single change.
type ChangeData struct {
	NodeID   devtools.NodeID `json:"node_id"`
	Type     string          `json:"type"` // added, removed, modified
	Path     string          `json:"path,omitempty"`
	OldValue interface{}     `json:"old_value,omitempty"`
	NewValue interface{}     `json:"new_value,omitempty"`
}

// =============================================================================
// Breakpoint Messages
// =============================================================================

// SetBreakpointPayload sets a breakpoint.
type SetBreakpointPayload struct {
	Breakpoint BreakpointData `json:"breakpoint"`
}

// BreakpointData represents a breakpoint.
type BreakpointData struct {
	ID       string           `json:"id"`
	NodeID   devtools.NodeID   `json:"node_id,omitempty"`
	FrameID  devtools.FrameID  `json:"frame_id,omitempty"`
	Condition string          `json:"condition,omitempty"`
	Enabled  bool             `json:"enabled"`
}

// ClearBreakpointPayload clears a breakpoint.
type ClearBreakpointPayload struct {
	BreakpointID string `json:"breakpoint_id"`
}

// BreakpointHitPayload is sent when a breakpoint is hit.
type BreakpointHitPayload struct {
	BreakpointID string           `json:"breakpoint_id"`
	FrameID      devtools.FrameID `json:"frame_id"`
	Context      map[string]interface{} `json:"context"`
}

// =============================================================================
// Evaluation Messages
// =============================================================================

// EvaluatePayload evaluates an expression.
type EvaluatePayload struct {
	Expression string           `json:"expression"`
	Context    map[string]interface{} `json:"context,omitempty"`
	FrameID    devtools.FrameID `json:"frame_id,omitempty"`
}

// EvaluationResultPayload contains the evaluation result.
type EvaluationResultPayload struct {
	Result    interface{} `json:"result"`
	Error     string      `json:"error,omitempty"`
	Type      string      `json:"type,omitempty"`
	FrameID   devtools.FrameID `json:"frame_id,omitempty"`
}

// =============================================================================
// Remote Session
// =============================================================================

// Session represents a remote debugging session.
type Session struct {
	mu            sync.RWMutex
	id            string
	clientID      string
	connected     bool
	subscriptions map[string]bool
	breakpoints   map[string]*BreakpointData
 createdAt     time.Time
	lastActivity  time.Time
}

// NewSession creates a new session.
func NewSession(sessionID, clientID string) *Session {
	return &Session{
		id:            sessionID,
		clientID:      clientID,
		connected:     true,
		subscriptions: make(map[string]bool),
		breakpoints:   make(map[string]*BreakpointData),
		createdAt:     time.Now(),
		lastActivity:  time.Now(),
	}
}

// ID returns the session ID.
func (s *Session) ID() string {
	return s.id
}

// ClientID returns the client ID.
func (s *Session) ClientID() string {
	return s.clientID
}

// IsConnected returns true if the session is connected.
func (s *Session) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// Close closes the session.
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connected = false
}

// UpdateActivity updates the last activity time.
func (s *Session) UpdateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now()
}

// IsStale returns true if the session is stale (no activity for a while).
func (s *Session) IsStale(timeout time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.lastActivity) > timeout
}

// Subscribe subscribes to a message type.
func (s *Session) Subscribe(messageType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[messageType] = true
}

// Unsubscribe unsubscribes from a message type.
func (s *Session) Unsubscribe(messageType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscriptions, messageType)
}

// IsSubscribed returns true if subscribed to a message type.
func (s *Session) IsSubscribed(messageType string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subscriptions[messageType]
}

// AddBreakpoint adds a breakpoint.
func (s *Session) AddBreakpoint(id string, bp *BreakpointData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.breakpoints[id] = bp
}

// RemoveBreakpoint removes a breakpoint.
func (s *Session) RemoveBreakpoint(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.breakpoints, id)
}

// GetBreakpoints returns all breakpoints.
func (s *Session) GetBreakpoints() map[string]*BreakpointData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*BreakpointData, len(s.breakpoints))
	for k, v := range s.breakpoints {
		result[k] = v
	}
	return result
}

// =============================================================================
// Remote Server
// =============================================================================

// Server handles remote debugging connections.
type Server struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	enabled    bool
	port       int
	path       string
	msgHandler MessageHandler
}

// MessageHandler handles incoming messages.
type MessageHandler func(session *Session, msg *Message) *Message

// NewServer creates a new remote debugging server.
func NewServer(port int, path string) *Server {
	return &Server{
		sessions: make(map[string]*Session),
		enabled:  false,
		port:     port,
		path:     path,
	}
}

// Enable enables the remote server.
func (s *Server) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable disables the remote server.
func (s *Server) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// IsEnabled returns true if the server is enabled.
func (s *Server) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetMessageHandler sets the message handler.
func (s *Server) SetMessageHandler(handler MessageHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgHandler = handler
}

// CreateSession creates a new session.
func (s *Server) CreateSession(clientID string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := fmt.Sprintf("session-%d-%s", time.Now().UnixNano(), clientID)
	session := NewSession(sessionID, clientID)
	s.sessions[sessionID] = session

	return session
}

// GetSession returns a session by ID.
func (s *Server) GetSession(sessionID string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	return session, exists
}

// RemoveSession removes a session.
func (s *Server) RemoveSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

// GetSessions returns all active sessions.
func (s *Server) GetSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		result = append(result, session)
	}
	return result
}

// CleanupStaleSessions removes stale sessions.
func (s *Server) CleanupStaleSessions(timeout time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, session := range s.sessions {
		if session.IsStale(timeout) {
			session.Close()
			delete(s.sessions, id)
			count++
		}
	}
	return count
}

// =============================================================================
// Message Utilities
// =============================================================================

// NewMessage creates a new protocol message.
func NewMessage(msgType string, payload interface{}) *Message {
	return &Message{
		Version: ProtocolVersion,
		Type:    msgType,
		Payload: payload,
	}
}

// NewMessageWithID creates a new message with a specific ID.
func NewMessageWithID(id, msgType string, payload interface{}) *Message {
	return &Message{
		Version: ProtocolVersion,
		Type:    msgType,
		ID:      id,
		Payload: payload,
	}
}

// NewError creates an error message.
func NewError(originalID string, errMsg string) *Message {
	return &Message{
		Version: ProtocolVersion,
		Type:    TypeError,
		ID:      originalID,
		Error:   errMsg,
	}
}

// Serialize converts a message to JSON.
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage creates a message from JSON.
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// =============================================================================
// Chrome DevTools Protocol (CDP) Compatibility
// =============================================================================

// CDPMessage represents a Chrome DevTools Protocol message.
type CDPMessage struct {
	ID     int64           `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *CDPError       `json:"error,omitempty"`
}

// CDPError represents a CDP error.
type CDPError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// CDPDomains maps DevTools features to CDP domains.
var CDPDomains = map[string]string{
	"snapshot": "DOM",
	"event":    "Runtime",
	"pattern":  "Debugger",
	"causal":   "Profiler",
}

// ToCDP converts a DevTools message to CDP format.
func ToCDP(msg *Message) (*CDPMessage, error) {
	cdp := &CDPMessage{}

	switch msg.Type {
	case TypeGetSnapshot:
		cdp.Method = "DOM.getSnapshot"
	case TypeEvent:
		cdp.Method = "Runtime.evaluate"
	case TypeGetDiff:
		cdp.Method = "DOM.compare"
	default:
		return nil, fmt.Errorf("unsupported message type for CDP: %s", msg.Type)
	}

	// Convert payload to params
	if msg.Payload != nil {
		params, err := json.Marshal(msg.Payload)
		if err != nil {
			return nil, err
		}
		cdp.Params = params
	}

	return cdp, nil
}

// FromCDP converts a CDP message to DevTools format.
func FromCDP(cdp *CDPMessage) (*Message, error) {
	msg := &Message{
		Version: ProtocolVersion,
	}

	if cdp.Error != nil {
		msg.Type = TypeError
		msg.Error = cdp.Error.Message
		return msg, nil
	}

	// Parse method to determine message type
	// For now, return a generic event message
	msg.Type = TypeEvent

	return msg, nil
}

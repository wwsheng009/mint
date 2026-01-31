// Package protocol provides unified protocol definitions for DevTools communication.
//
// This protocol supports both:
//   - TUI Debugging: Terminal-based debugging with inspect, highlight, replay
//   - Remote Debugging: WebSocket-based remote debugging with CDP compatibility
//
// Message Flow:
//
//	Client -> Server:
//	  - handshake: Initial connection
//	  - subscribe/unsubscribe: Event subscriptions
//	  - get_snapshot, get_range, get_diff: Snapshot queries
//	  - set_breakpoint, clear_breakpoint: Breakpoint management
//	  - inspect, highlight: TUI-specific commands
//	  - evaluate: Expression evaluation
//
//	Server -> Client:
//	  - handshake_ack: Connection acknowledgment
//	  - event: Real-time events
//	  - snapshot: Snapshot data
//	  - diff: Diff results
//	  - error: Error responses
package protocol

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// Protocol version
const Version = "1.0.0"

// =============================================================================
// Message Types
// =============================================================================

// MessageType represents the type of protocol message.
type MessageType string

const (
	// Handshake messages
	TypeHandshake    MessageType = "handshake"
	TypeHandshakeAck MessageType = "handshake_ack"

	// Subscription messages
	TypeSubscribe   MessageType = "subscribe"
	TypeUnsubscribe MessageType = "unsubscribe"

	// Query messages
	TypeGetSnapshot MessageType = "get_snapshot"
	TypeGetRange    MessageType = "get_range"
	TypeGetDiff     MessageType = "get_diff"

	// Breakpoint messages
	TypeSetBreakpoint    MessageType = "set_breakpoint"
	TypeClearBreakpoint  MessageType = "clear_breakpoint"
	TypeBreakpointHit    MessageType = "breakpoint_hit"

	// Evaluation messages
	TypeEvaluate          MessageType = "evaluate"
	TypeEvaluationResult  MessageType = "evaluation_result"

	// TUI-specific messages
	TypeInspect   MessageType = "inspect"
	TypeHighlight MessageType = "highlight"
	TypeReplayStart MessageType = "replay_start"
	TypeReplayStop  MessageType = "replay_stop"
	TypeGetFrame    MessageType = "get_frame"
	TypeGetTimeline MessageType = "get_timeline"

	// Server messages
	TypeEvent    MessageType = "event"
	TypeSnapshot MessageType = "snapshot"
	TypeDiff     MessageType = "diff"
	TypeState    MessageType = "state"
	TypeResponse MessageType = "response"
	TypeError    MessageType = "error"

	// Control messages
	TypeHeartbeat MessageType = "heartbeat"
)

// =============================================================================
// Base Message
// =============================================================================

// Message is the base protocol message.
type Message struct {
	Version  string      `json:"version"`
	Type     MessageType `json:"type"`
	ID       string      `json:"id,omitempty"`
	Timestamp time.Time  `json:"timestamp,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// NewMessage creates a new protocol message.
func NewMessage(msgType MessageType, payload interface{}) *Message {
	return &Message{
		Version:  Version,
		Type:     msgType,
		Payload:  payload,
		Timestamp: time.Now(),
	}
}

// NewMessageWithID creates a new message with a specific ID.
func NewMessageWithID(id string, msgType MessageType, payload interface{}) *Message {
	return &Message{
		Version:   Version,
		Type:      msgType,
		ID:        id,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// NewError creates an error message.
func NewError(originalID string, errMsg string) *Message {
	return &Message{
		Version:   Version,
		Type:      TypeError,
		ID:        originalID,
		Error:     errMsg,
		Timestamp: time.Now(),
	}
}

// NewResponse creates a response message.
func NewResponse(originalID string, success bool, data interface{}, errMsg string) *Message {
	return &Message{
		Version:   Version,
		Type:      TypeResponse,
		ID:        originalID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"success": success,
			"data":    data,
			"error":   errMsg,
		},
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
// Handshake
// =============================================================================

// HandshakePayload is the initial handshake payload.
type HandshakePayload struct {
	ClientID    string   `json:"client_id"`
	Capabilities []string `json:"capabilities"`
	Version     string   `json:"version"`
	Protocol    string   `json:"protocol,omitempty"` // "tui" or "remote"
}

// HandshakeAckPayload is the handshake acknowledgment.
type HandshakeAckPayload struct {
	ServerID     string   `json:"server_id"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	SessionID    string   `json:"session_id"`
}

// =============================================================================
// Event Messages
// =============================================================================

// EventPayload represents an event message.
type EventPayload struct {
	FrameID   devtools.FrameID `json:"frame_id"`
	EventType string           `json:"event_type"`
	NodeID    devtools.NodeID   `json:"node_id,omitempty"`
	Phase     string           `json:"phase,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time        `json:"timestamp"`
}

// =============================================================================
// Snapshot Messages
// =============================================================================

// GetSnapshotPayload is the payload for getting a snapshot.
type GetSnapshotPayload struct {
	FrameID         devtools.FrameID `json:"frame_id"`
	IncludeState    bool             `json:"include_state,omitempty"`
	IncludeChildren bool             `json:"include_children,omitempty"`
}

// SnapshotPayload contains a snapshot.
type SnapshotPayload struct {
	FrameID      devtools.FrameID `json:"frame_id"`
	Timestamp    time.Time        `json:"timestamp"`
	WindowState  WindowState      `json:"window_state"`
	Components   []ComponentData  `json:"components"`
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
	From  devtools.FrameID `json:"from"`
	To    devtools.FrameID `json:"to"`
	Limit int              `json:"limit,omitempty"`
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
	Mutations int              `json:"mutations"`
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
	ID        string           `json:"id"`
	NodeID    devtools.NodeID  `json:"node_id,omitempty"`
	FrameID   devtools.FrameID `json:"frame_id,omitempty"`
	Condition string           `json:"condition,omitempty"`
	Enabled   bool             `json:"enabled"`
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
	Expression string                 `json:"expression"`
	Context    map[string]interface{} `json:"context,omitempty"`
	FrameID    devtools.FrameID       `json:"frame_id,omitempty"`
}

// EvaluationResultPayload contains the evaluation result.
type EvaluationResultPayload struct {
	Result  interface{}      `json:"result"`
	Error   string           `json:"error,omitempty"`
	Type    string           `json:"type,omitempty"`
	FrameID devtools.FrameID `json:"frame_id,omitempty"`
}

// =============================================================================
// TUI-specific Messages
// =============================================================================

// InspectPayload represents an inspect request.
type InspectPayload struct {
	NodeID string `json:"node_id"`
}

// InspectResultPayload contains the inspect result.
type InspectResultPayload struct {
	NodeID string                 `json:"node_id"`
	Type   string                 `json:"type"`
	State  map[string]interface{} `json:"state"`
}

// HighlightPayload represents a highlight request.
type HighlightPayload struct {
	NodeID string `json:"node_id"`
	Color  string `json:"color,omitempty"`
}

// HighlightResultPayload contains the highlight result.
type HighlightResultPayload struct {
	NodeID  string `json:"node_id"`
	Color   string `json:"color"`
	Status  string `json:"status"`
}

// GetFramePayload requests frame data.
type GetFramePayload struct {
	FrameID devtools.FrameID `json:"frame_id"`
}

// FramePayload contains frame data.
type FramePayload struct {
	FrameID devtools.FrameID     `json:"frame_id"`
	Events  []EventData          `json:"events"`
	State   map[string]interface{} `json:"state"`
}

// EventData represents event data in frames.
type EventData struct {
	Type   string                 `json:"type"`
	NodeID string                 `json:"node_id,omitempty"`
	Data   map[string]interface{} `json:"data"`
}

// GetTimelinePayload requests timeline data.
type GetTimelinePayload struct {
	From devtools.FrameID `json:"from,omitempty"`
	To   devtools.FrameID `json:"to,omitempty"`
	Limit int             `json:"limit,omitempty"`
}

// TimelinePayload contains timeline data.
type TimelinePayload struct {
	Frames []FrameSummary `json:"frames"`
	Count  int             `json:"count"`
}

// ReplayPayload controls replay.
type ReplayPayload struct {
	From     devtools.FrameID `json:"from,omitempty"`
	To       devtools.FrameID `json:"to,omitempty"`
	Speed    float64          `json:"speed,omitempty"`
}

// =============================================================================
// Subscription
// =============================================================================

// SubscribePayload subscribes to events.
type SubscribePayload struct {
	EventType string `json:"event_type"`
}

// UnsubscribePayload unsubscribes from events.
type UnsubscribePayload struct {
	EventType string `json:"event_type"`
}

// =============================================================================
// Session
// =============================================================================

// Session represents a debugging session.
type Session struct {
	mu            sync.RWMutex
	id            string
	clientID      string
	connected     bool
	protocol      string // "tui" or "remote"
	subscriptions map[string]bool
	breakpoints   map[string]*BreakpointData
	createdAt     time.Time
	lastActivity  time.Time
}

// NewSession creates a new session.
func NewSession(sessionID, clientID, protocol string) *Session {
	return &Session{
		id:            sessionID,
		clientID:      clientID,
		protocol:      protocol,
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientID
}

// Protocol returns the protocol type.
func (s *Session) Protocol() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.protocol
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

// IsStale returns true if the session is stale.
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
// Session Manager
// =============================================================================

// SessionManager manages debugging sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session.
func (sm *SessionManager) CreateSession(clientID, protocol string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := fmt.Sprintf("session-%d-%s", time.Now().UnixNano(), clientID)
	session := NewSession(sessionID, clientID, protocol)
	sm.sessions[sessionID] = session

	return session
}

// GetSession returns a session by ID.
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	return session, exists
}

// RemoveSession removes a session.
func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// GetSessions returns all active sessions.
func (sm *SessionManager) GetSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		result = append(result, session)
	}
	return result
}

// GetSessionsByProtocol returns sessions by protocol type.
func (sm *SessionManager) GetSessionsByProtocol(protocol string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*Session, 0)
	for _, session := range sm.sessions {
		if session.Protocol() == protocol {
			result = append(result, session)
		}
	}
	return result
}

// CleanupStaleSessions removes stale sessions.
func (sm *SessionManager) CleanupStaleSessions(timeout time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := 0
	for id, session := range sm.sessions {
		if session.IsStale(timeout) {
			session.Close()
			delete(sm.sessions, id)
			count++
		}
	}
	return count
}

// =============================================================================
// CDP Compatibility
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
		Version: Version,
	}

	if cdp.Error != nil {
		msg.Type = TypeError
		msg.Error = cdp.Error.Message
		return msg, nil
	}

	msg.Type = TypeEvent
	return msg, nil
}

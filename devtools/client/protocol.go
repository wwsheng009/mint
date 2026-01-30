// Package client provides WebSocket protocol for DevTools remote debugging.
//
// This file implements the WebSocket protocol for communication
// between the TUI application and remote debug clients.
package client

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// ProtocolVersion is the current protocol version.
const ProtocolVersion = "1.0.0"

// MessageType represents the type of WebSocket message.
type MessageType string

const (
	// MessageHello is the handshake message.
	MessageHello MessageType = "hello"
	// MessageHelloAck acknowledges the handshake.
	MessageHelloAck MessageType = "hello_ack"
	// MessageEvent is an event message.
	MessageEvent MessageType = "event"
	// MessageState is a state update message.
	MessageState MessageType = "state"
	// MessageCommand is a command message.
	MessageCommand MessageType = "command"
	// MessageResponse is a response to a command.
	MessageResponse MessageType = "response"
	// MessageError is an error message.
	MessageError MessageType = "error"
	// MessageHeartbeat is a heartbeat message.
	MessageHeartbeat MessageType = "heartbeat"
)

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Version   MessageType   `json:"version"`
	Type      MessageType   `json:"type"`
	ID        string        `json:"id,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Data      interface{}   `json:"data"`
}

// HelloData contains handshake data.
type HelloData struct {
	ProtocolVersion string                 `json:"protocol_version"`
	ClientID       string                 `json:"client_id"`
	Capabilities   []string               `json:"capabilities"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// HelloAckData contains handshake acknowledgment data.
type HelloAckData struct {
	ServerVersion string                 `json:"server_version"`
	SessionID     string                 `json:"session_id"`
	Capabilities   []string               `json:"capabilities"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// EventData contains event data.
type EventData struct {
	FrameID   devtools.FrameID    `json:"frame_id"`
	Type      string               `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

// StateData contains state update data.
type StateData struct {
	FrameID      devtools.FrameID `json:"frame_id"`
	ComponentID  uint32           `json:"component_id,omitempty"`
	PropertyName string           `json:"property_name,omitempty"`
	OldValue     interface{}      `json:"old_value,omitempty"`
	NewValue     interface{}      `json:"new_value,omitempty"`
}

// CommandData contains command data.
type CommandData struct {
	Command string        `json:"command"`
	Args    []string      `json:"args,omitempty"`
	Options []interface{} `json:"options,omitempty"`
}

// ResponseData contains response data.
type ResponseData struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorData contains error data.
type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// WebSocketHandler handles WebSocket connections and messages.
type WebSocketHandler struct {
	mu            sync.RWMutex
	connected     bool
	sessionID      string
	clientID       string
	messageCounter int64
	callbacks      map[string]chan WSMessage
	heartbeat      time.Duration
	lastHeartbeat  time.Time
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		callbacks:     make(map[string]chan WSMessage),
		heartbeat:     30 * time.Second,
		lastHeartbeat: time.Now(),
	}
}

// Connect handles a new connection.
func (wh *WebSocketHandler) Connect(clientID string) (*WSMessage, error) {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	wh.clientID = clientID
	wh.connected = true
	wh.sessionID = fmt.Sprintf("session_%d", time.Now().Unix())

	// Create hello message
	helloMsg := &WSMessage{
		Type:      MessageHello,
		ID:        wh.nextID(),
		Timestamp: time.Now(),
		Data: HelloAckData{
			ServerVersion: ProtocolVersion,
			SessionID:     wh.sessionID,
			Capabilities: []string{
				"timeline",
				"causal_graph",
				"snapshots",
				"replay",
				"inspect",
				"profiler",
			},
		},
	}

	return helloMsg, nil
}

// Disconnect handles disconnection.
func (wh *WebSocketHandler) Disconnect() {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	wh.connected = false

	// Close all callback channels
	for _, ch := range wh.callbacks {
		close(ch)
	}
	wh.callbacks = make(map[string]chan WSMessage)
}

// HandleMessage handles an incoming WebSocket message.
func (wh *WebSocketHandler) HandleMessage(msg *WSMessage) (*WSMessage, error) {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	if !wh.connected {
		return nil, fmt.Errorf("not connected")
	}

	switch msg.Type {
	case MessageHello:
		return wh.handleHello(msg)
	case MessageCommand:
		return wh.handleCommand(msg)
	case MessageHeartbeat:
		return wh.handleHeartbeat()
	default:
		return &WSMessage{
			Type:      MessageError,
			ID:        msg.ID,
			Timestamp: time.Now(),
			Data: ErrorData{
				Code:    400,
				Message: "unknown message type",
			},
		}, nil
	}
}

// handleHello handles a hello message.
func (wh *WebSocketHandler) handleHello(msg *WSMessage) (*WSMessage, error) {
	// Validate protocol version
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid hello data")
	}

	protocolVersion, _ := data["protocol_version"].(string)
	if protocolVersion != ProtocolVersion {
		return &WSMessage{
			Type:      MessageError,
			ID:        msg.ID,
			Timestamp: time.Now(),
			Data: ErrorData{
				Code:    1001,
				Message: "unsupported protocol version",
				Details: fmt.Sprintf("Server version: %s, Client version: %s", ProtocolVersion, protocolVersion),
			},
		}, nil
	}

	wh.lastHeartbeat = time.Now()

	return &WSMessage{
		Type:      MessageHelloAck,
		ID:        msg.ID,
		Timestamp: time.Now(),
		Data: HelloAckData{
			ServerVersion: ProtocolVersion,
			SessionID:     wh.sessionID,
			Capabilities: []string{
				"timeline",
				"causal_graph",
				"snapshots",
				"replay",
			},
		},
	}, nil
}

// handleCommand handles a command message.
func (wh *WebSocketHandler) handleCommand(msg *WSMessage) (*WSMessage, error) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid command data")
	}

	command, _ := data["command"].(string)

	// Handle different commands
	var result interface{}
	var err error

	switch command {
	case "inspect":
		result, err = wh.cmdInspect(data)
	case "highlight":
		result, err = wh.cmdHighlight(data)
	case "get_frame":
		result, err = wh.cmdGetFrame(data)
	case "get_timeline":
		result, err = wh.cmdGetTimeline(data)
	case "replay_start":
		result, err = wh.cmdReplayStart(data)
	case "replay_stop":
		result, err = wh.cmdReplayStop(data)
	default:
		err = fmt.Errorf("unknown command: %s", command)
	}

	responseData := ResponseData{
		Success: err == nil,
		Data:    result,
	}

	if err != nil {
		responseData.Error = err.Error()
	}

	return &WSMessage{
		Type:      MessageResponse,
		ID:        msg.ID,
		Timestamp: time.Now(),
		Data:      responseData,
	}, nil
}

// handleHeartbeat handles a heartbeat message.
func (wh *WebSocketHandler) handleHeartbeat() (*WSMessage, error) {
	wh.lastHeartbeat = time.Now()

	return &WSMessage{
		Type:      MessageHeartbeat,
		ID:        "",
		Timestamp: time.Now(),
	}, nil
}

// cmdInspect handles the inspect command.
func (wh *WebSocketHandler) cmdInspect(data map[string]interface{}) (interface{}, error) {
	nodeID, _ := data["node_id"].(string)

	return map[string]interface{}{
		"node_id": nodeID,
		"type":    "Container",
		"state":   make(map[string]interface{}),
	}, nil
}

// cmdHighlight handles the highlight command.
func (wh *WebSocketHandler) cmdHighlight(data map[string]interface{}) (interface{}, error) {
	nodeID, _ := data["node_id"].(string)
	color, _ := data["color"].(string)

	return map[string]interface{}{
		"node_id": nodeID,
		"color":   color,
		"status":  "highlighted",
	}, nil
}

// cmdGetFrame handles the get_frame command.
func (wh *WebSocketHandler) cmdGetFrame(data map[string]interface{}) (interface{}, error) {
	frameIDVal, _ := data["frame_id"].(float64)
	frameID := devtools.FrameID(frameIDVal)

	return map[string]interface{}{
		"frame_id": frameID,
		"events":   []interface{}{},
		"state":    make(map[string]interface{}),
	}, nil
}

// cmdGetTimeline handles the get_timeline command.
func (wh *WebSocketHandler) cmdGetTimeline(data map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"frames": []interface{}{},
		"count":  0,
	}, nil
}

// cmdReplayStart handles the replay_start command.
func (wh *WebSocketHandler) cmdReplayStart(data map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"status": "replaying",
	}, nil
}

// cmdReplayStop handles the replay_stop command.
func (wh *WebSocketHandler) cmdReplayStop(data map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"status": "stopped",
	}, nil
}

// SendEvent sends an event message to the client.
func (wh *WebSocketHandler) SendEvent(frameID devtools.FrameID, eventType string, data map[string]interface{}) error {
	wh.mu.RLock()
	defer wh.mu.RUnlock()

	if !wh.connected {
		return fmt.Errorf("not connected")
	}

	msg := &WSMessage{
		Type:      MessageEvent,
		ID:        wh.nextID(),
		Timestamp: time.Now(),
		Data: EventData{
			FrameID: frameID,
			Type:    eventType,
			Data:    data,
		},
	}

	return wh.send(msg)
}

// SendState sends a state update message to the client.
func (wh *WebSocketHandler) SendState(frameID devtools.FrameID, componentID uint32, property string, oldValue, newValue interface{}) error {
	wh.mu.RLock()
	defer wh.mu.RUnlock()

	if !wh.connected {
		return fmt.Errorf("not connected")
	}

	msg := &WSMessage{
		Type:      MessageState,
		ID:        wh.nextID(),
		Timestamp: time.Now(),
		Data: StateData{
			FrameID:      frameID,
			ComponentID:  componentID,
			PropertyName: property,
			OldValue:     oldValue,
			NewValue:     newValue,
		},
	}

	return wh.send(msg)
}

// send sends a message (implementation-specific).
func (wh *WebSocketHandler) send(msg *WSMessage) error {
	// This would interface with the actual WebSocket connection
	// For now, just serialize to JSON to verify it works
	_, err := json.Marshal(msg)
	return err
}

// nextID returns the next message ID.
func (wh *WebSocketHandler) nextID() string {
	wh.messageCounter++
	return fmt.Sprintf("msg_%d", wh.messageCounter)
}

// GetSessionID returns the session ID.
func (wh *WebSocketHandler) GetSessionID() string {
	wh.mu.RLock()
	defer wh.mu.RUnlock()
	return wh.sessionID
}

// IsConnected returns whether connected.
func (wh *WebSocketHandler) IsConnected() bool {
	wh.mu.RLock()
	defer wh.mu.RUnlock()
	return wh.connected
}

// ShouldSendHeartbeat returns whether a heartbeat should be sent.
func (wh *WebSocketHandler) ShouldSendHeartbeat() bool {
	wh.mu.RLock()
	defer wh.mu.RUnlock()

	return time.Since(wh.lastHeartbeat) > wh.heartbeat
}

// SendHeartbeat sends a heartbeat message.
func (wh *WebSocketHandler) SendHeartbeat() error {
	msg := &WSMessage{
		Type:      MessageHeartbeat,
		Timestamp: time.Now(),
	}

	return wh.send(msg)
}

// BroadcastData broadcasts data to all connected clients.
func (wh *WebSocketHandler) BroadcastData(dataType string, data interface{}) error {
	wh.mu.RLock()
	defer wh.mu.RUnlock()

	if !wh.connected {
		return fmt.Errorf("not connected")
	}

	msg := &WSMessage{
		Type:      MessageType(dataType),
		ID:        wh.nextID(),
		Timestamp: time.Now(),
		Data:      data,
	}

	return wh.send(msg)
}

// MessageQueue represents a queue of WebSocket messages.
type MessageQueue struct {
	mu     sync.RWMutex
	messages []WSMessage
	maxSize int
}

// NewMessageQueue creates a new message queue.
func NewMessageQueue(maxSize int) *MessageQueue {
	return &MessageQueue{
		messages: make([]WSMessage, 0, maxSize),
		maxSize: maxSize,
	}
}

// Push adds a message to the queue.
func (mq *MessageQueue) Push(msg WSMessage) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if len(mq.messages) >= mq.maxSize {
		return fmt.Errorf("queue full")
	}

	mq.messages = append(mq.messages, msg)
	return nil
}

// Pop removes and returns the oldest message.
func (mq *MessageQueue) Pop() (WSMessage, bool) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if len(mq.messages) == 0 {
		return WSMessage{}, false
	}

	msg := mq.messages[0]
	mq.messages = mq.messages[1:]
	return msg, true
}

// Peek returns the oldest message without removing it.
func (mq *MessageQueue) Peek() (WSMessage, bool) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	if len(mq.messages) == 0 {
		return WSMessage{}, false
	}

	return mq.messages[0], true
}

// Clear clears all messages.
func (mq *MessageQueue) Clear() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.messages = make([]WSMessage, 0, mq.maxSize)
}

// Size returns the number of messages.
func (mq *MessageQueue) Size() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return len(mq.messages)
}

// ConnectionInfo represents information about a connection.
type ConnectionInfo struct {
	SessionID    string    `json:"session_id"`
	ClientID     string    `json:"client_id"`
	ConnectedAt  time.Time `json:"connected_at"`
	MessageCount int64    `json:"message_count"`
	Capabilities []string `json:"capabilities"`
}

// GetConnectionInfo returns connection information.
func (wh *WebSocketHandler) GetConnectionInfo() *ConnectionInfo {
	wh.mu.RLock()
	defer wh.mu.RUnlock()

	return &ConnectionInfo{
		SessionID:     wh.sessionID,
		ClientID:      wh.clientID,
		ConnectedAt:   time.Now(), // Simplified
		MessageCount:  wh.messageCounter,
		Capabilities: []string{
			"send",
			"receive",
			"heartbeat",
		},
	}
}

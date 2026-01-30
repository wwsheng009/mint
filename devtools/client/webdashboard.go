// Package client provides Web Dashboard for DevTools.
//
// This file implements a web-based dashboard for remote debugging.
package client

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// WebDashboard provides a web-based debug dashboard.
type WebDashboard struct {
	mu            sync.RWMutex
	enabled       bool
	port          int
	websocket     *WebSocketHandler

	// Dashboard data
	frames        []*DashboardFrame
	components    map[string]*DashboardComponent
	metrics       *DashboardMetrics

	// Real-time updates
	updateChan    chan *DashboardUpdate
}

// DashboardFrame represents a frame in the dashboard.
type DashboardFrame struct {
	FrameID      devtools.FrameID
	Timestamp    time.Time
	Duration     time.Duration
	EventCount   int
	MutationCount int
	LayoutCount  int
	RepaintCount int
}

// DashboardComponent represents a component in the dashboard.
type DashboardComponent struct {
	ID         string
	Type       string
	Properties map[string]interface{}
	Styles     map[string]interface{}
	Children   []string
}

// DashboardMetrics represents dashboard metrics.
type DashboardMetrics struct {
	FPS          float64
	FrameTime    time.Duration
	LayoutTime   time.Duration
	PaintTime    time.Duration
	MemoryUsage  uint64
	ComponentCount int
}

// DashboardUpdate represents a dashboard update.
type DashboardUpdate struct {
	Type      string
	Timestamp time.Time
	Data      interface{}
}

// NewWebDashboard creates a new web dashboard.
func NewWebDashboard(port int) *WebDashboard {
	ws := NewWebSocketHandler()

	return &WebDashboard{
		port:       port,
		websocket:  ws,
		frames:     make([]*DashboardFrame, 0, 100),
		components: make(map[string]*DashboardComponent),
		metrics:    &DashboardMetrics{},
		updateChan: make(chan *DashboardUpdate, 256),
	}
}

// Start starts the web dashboard.
func (wd *WebDashboard) Start() error {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	if wd.enabled {
		return fmt.Errorf("dashboard already running")
	}

	wd.enabled = true

	// Start update handler
	go wd.updateLoop()

	// TODO: Start HTTP server
	// TODO: Start WebSocket server

	return nil
}

// Stop stops the web dashboard.
func (wd *WebDashboard) Stop() error {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	if !wd.enabled {
		return fmt.Errorf("dashboard not running")
	}

	wd.enabled = false
	close(wd.updateChan)
	return nil
}

// IsRunning returns whether the dashboard is running.
func (wd *WebDashboard) IsRunning() bool {
	wd.mu.RLock()
	defer wd.mu.RUnlock()
	return wd.enabled
}

// AddFrame adds a frame to the dashboard.
func (wd *WebDashboard) AddFrame(frame *DashboardFrame) {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	wd.frames = append(wd.frames, frame)

	// Trim to 100 frames
	if len(wd.frames) > 100 {
		wd.frames = wd.frames[1:]
	}

	// Send update
	wd.sendUpdate(&DashboardUpdate{
		Type:      "frame",
		Timestamp: time.Now(),
		Data:      frame,
	})
}

// UpdateComponent updates a component in the dashboard.
func (wd *WebDashboard) UpdateComponent(id string, component *DashboardComponent) {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	wd.components[id] = component

	wd.sendUpdate(&DashboardUpdate{
		Type:      "component",
		Timestamp: time.Now(),
		Data:      component,
	})
}

// UpdateMetrics updates the dashboard metrics.
func (wd *WebDashboard) UpdateMetrics(metrics *DashboardMetrics) {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	wd.metrics = metrics

	wd.sendUpdate(&DashboardUpdate{
		Type:      "metrics",
		Timestamp: time.Now(),
		Data:      metrics,
	})
}

// GetFrames returns all frames.
func (wd *WebDashboard) GetFrames() []*DashboardFrame {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	frames := make([]*DashboardFrame, len(wd.frames))
	copy(frames, wd.frames)
	return frames
}

// GetFrame returns a specific frame.
func (wd *WebDashboard) GetFrame(frameID devtools.FrameID) *DashboardFrame {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	for _, frame := range wd.frames {
		if frame.FrameID == frameID {
			return frame
		}
	}
	return nil
}

// GetComponent returns a component.
func (wd *WebDashboard) GetComponent(id string) *DashboardComponent {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	return wd.components[id]
}

// GetMetrics returns the current metrics.
func (wd *WebDashboard) GetMetrics() *DashboardMetrics {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	return wd.metrics
}

// GetWebSocketHandler returns the WebSocket handler.
func (wd *WebDashboard) GetWebSocketHandler() *WebSocketHandler {
	return wd.websocket
}

// sendUpdate sends an update to all connected clients.
func (wd *WebDashboard) sendUpdate(update *DashboardUpdate) {
	select {
	case wd.updateChan <- update:
	default:
		// Channel full, drop update
	}
}

// updateLoop handles dashboard updates.
func (wd *WebDashboard) updateLoop() {
	for update := range wd.updateChan {
		if !wd.enabled {
			return
		}

		// Broadcast update via WebSocket
		data, _ := json.Marshal(update)
		_ = data
		// TODO: Send to WebSocket clients
	}
}

// ExportData exports dashboard data as JSON.
func (wd *WebDashboard) ExportData() ([]byte, error) {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	data := map[string]interface{}{
		"frames":     wd.frames,
		"components":  wd.components,
		"metrics":     wd.metrics,
		"exported_at": time.Now(),
	}

	return json.MarshalIndent(data, "", "  ")
}

// ImportData imports dashboard data from JSON.
func (wd *WebDashboard) ImportData(data []byte) error {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	var imported map[string]interface{}
	if err := json.Unmarshal(data, &imported); err != nil {
		return err
	}

	// Import frames
	if frames, ok := imported["frames"].([]interface{}); ok {
		for _, f := range frames {
			frameJSON, _ := json.Marshal(f)
			var frame DashboardFrame
			if err := json.Unmarshal(frameJSON, &frame); err == nil {
				wd.frames = append(wd.frames, &frame)
			}
		}
	}

	// Import metrics
	if metricsJSON, ok := imported["metrics"].([]byte); ok {
		var metrics DashboardMetrics
		if err := json.Unmarshal(metricsJSON, &metrics); err == nil {
			wd.metrics = &metrics
		}
	}

	return nil
}

// GenerateReport generates a debug report.
func (wd *WebDashboard) GenerateReport() *DebugReport {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	report := &DebugReport{
		GeneratedAt: time.Now(),
		FrameCount:  len(wd.frames),
		ComponentCount: len(wd.components),
		Metrics:     *wd.metrics,
	}

	// Analyze frames
	if len(wd.frames) > 0 {
		report.TotalDuration = wd.frames[len(wd.frames)-1].Timestamp.Sub(wd.frames[0].Timestamp)
		report.AvgFrameTime = report.TotalDuration / time.Duration(len(wd.frames))
	}

	// Find slow frames
	for _, frame := range wd.frames {
		if frame.Duration > report.AvgFrameTime*2 {
			report.SlowFrames = append(report.SlowFrames, frame)
		}
	}

	return report
}

// DebugReport represents a comprehensive debug report.
type DebugReport struct {
	GeneratedAt    time.Time
	FrameCount     int
	ComponentCount int
	TotalDuration  time.Duration
	AvgFrameTime   time.Duration
	SlowFrames     []*DashboardFrame
	Metrics        DashboardMetrics
}

// GetPort returns the dashboard port.
func (wd *WebDashboard) GetPort() int {
	return wd.port
}

// SetPort sets the dashboard port.
func (wd *WebDashboard) SetPort(port int) {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	wd.port = port
}

// ClientInfo represents information about a connected client.
type ClientInfo struct {
	ID           string    `json:"id"`
	ConnectedAt   time.Time `json:"connected_at"`
	UserAgent     string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	MessageCount int64     `json:"message_count"`
}

// ClientManager manages connected clients.
type ClientManager struct {
	mu      sync.RWMutex
	clients map[string]*ClientInfo
}

// NewClientManager creates a new client manager.
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*ClientInfo),
	}
}

// AddClient adds a client.
func (cm *ClientManager) AddClient(client *ClientInfo) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[client.ID] = client
}

// RemoveClient removes a client.
func (cm *ClientManager) RemoveClient(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, id)
}

// GetClient returns a client.
func (cm *ClientManager) GetClient(id string) (*ClientInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.clients[id]
	return client, exists
}

// GetAllClients returns all clients.
func (cm *ClientManager) GetAllClients() []*ClientInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clients := make([]*ClientInfo, 0, len(cm.clients))
	for _, c := range cm.clients {
		clients = append(clients, c)
	}
	return clients
}

// GetClientCount returns the number of connected clients.
func (cm *ClientManager) GetClientCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.clients)
}

// Broadcast broadcasts a message to all clients.
func (cm *ClientManager) Broadcast(message interface{}) {
	cm.mu.RLock()
	clients := make([]*ClientInfo, 0, len(cm.clients))
	for _, c := range cm.clients {
		clients = append(clients, c)
	}
	cm.mu.RUnlock()

	for _, client := range clients {
		// Send message to client
		_ = client
	}
}

// HTTPServer represents an HTTP server for the dashboard.
type HTTPServer struct {
	mu    sync.RWMutex
	port  int
	dashboard *WebDashboard
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(port int, dashboard *WebDashboard) *HTTPServer {
	return &HTTPServer{
		port:      port,
		dashboard: dashboard,
	}
}

// Start starts the HTTP server.
func (hs *HTTPServer) Start() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// TODO: Start actual HTTP server
	return nil
}

// Stop stops the HTTP server.
func (hs *HTTPServer) Stop() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// TODO: Stop HTTP server
	return nil
}

// GetPort returns the server port.
func (hs *HTTPServer) GetPort() int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.port
}

// StaticFileHandler handles static file requests.
type StaticFileHandler struct {
	rootDir string
}

// NewStaticFileHandler creates a new static file handler.
func NewStaticFileHandler(rootDir string) *StaticFileHandler {
	return &StaticFileHandler{
		rootDir: rootDir,
	}
}

// Serve serves a static file.
func (sfh *StaticFileHandler) Serve(path string) ([]byte, error) {
	// TODO: Implement static file serving
	return nil, fmt.Errorf("not implemented")
}

// APIHandler handles API requests.
type APIHandler struct {
	dashboard *WebDashboard
}

// NewAPIHandler creates a new API handler.
func NewAPIHandler(dashboard *WebDashboard) *APIHandler {
	return &APIHandler{
		dashboard: dashboard,
	}
}

// HandleAPI handles an API request.
func (ah *APIHandler) HandleAPI(endpoint string, method string, data []byte) ([]byte, error) {
	switch endpoint {
	case "/api/frames":
		return ah.handleGetFrames()
	case "/api/metrics":
		return ah.handleGetMetrics()
	case "/api/components":
		return ah.handleGetComponents()
	case "/api/report":
		return ah.handleGetReport()
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

// handleGetFrames handles GET /api/frames.
func (ah *APIHandler) handleGetFrames() ([]byte, error) {
	frames := ah.dashboard.GetFrames()
	return json.Marshal(frames)
}

// handleGetMetrics handles GET /api/metrics.
func (ah *APIHandler) handleGetMetrics() ([]byte, error) {
	metrics := ah.dashboard.GetMetrics()
	return json.Marshal(metrics)
}

// handleGetComponents handles GET /api/components.
func (ah *APIHandler) handleGetComponents() ([]byte, error) {
	ah.dashboard.mu.RLock()
	defer ah.dashboard.mu.RUnlock()

	components := make(map[string]*DashboardComponent)
	for k, v := range ah.dashboard.components {
		components[k] = v
	}

	return json.Marshal(components)
}

// handleGetReport handles GET /api/report.
func (ah *APIHandler) handleGetReport() ([]byte, error) {
	report := ah.dashboard.GenerateReport()
	return json.Marshal(report)
}

// RemoteDebugSession represents a remote debugging session.
type RemoteDebugSession struct {
	ID           string
	StartTime    time.Time
	EndTime      time.Time
	ClientInfo   *ClientInfo
	Dashboard    *WebDashboard
	Active       bool
	Capabilities []string
}

// NewRemoteDebugSession creates a new remote debug session.
func NewRemoteDebugSession(clientID string, port int) *RemoteDebugSession {
	dashboard := NewWebDashboard(port)

	return &RemoteDebugSession{
		ID:        fmt.Sprintf("session_%d", time.Now().Unix()),
		StartTime: time.Now(),
		Dashboard: dashboard,
		Active:    true,
		Capabilities: []string{
			"inspect",
			"highlight",
			"timeline",
			"causal_graph",
			"snapshots",
			"replay",
			"metrics",
		},
	}
}

// Start starts the remote debug session.
func (rds *RemoteDebugSession) Start() error {
	return rds.Dashboard.Start()
}

// Stop stops the remote debug session.
func (rds *RemoteDebugSession) Stop() error {
	rds.Active = false
	rds.EndTime = time.Now()
	return rds.Dashboard.Stop()
}

// IsActive returns whether the session is active.
func (rds *RemoteDebugSession) IsActive() bool {
	return rds.Active
}

// GetDashboard returns the dashboard.
func (rds *RemoteDebugSession) GetDashboard() *WebDashboard {
	return rds.Dashboard
}

// GetSessionInfo returns session information.
func (rds *RemoteDebugSession) GetSessionInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":           rds.ID,
		"start_time":   rds.StartTime,
		"end_time":     rds.EndTime,
		"active":       rds.Active,
		"capabilities": rds.Capabilities,
		"client_id":     rds.ClientInfo.ID,
	}
}

// Package client provides Web Dashboard for DevTools.
//
// This file implements a web-based dashboard for remote debugging.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// WebDashboard provides a web-based debug dashboard.
type WebDashboard struct {
	mu            sync.RWMutex
	enabled       bool
	port          int
	wsServer      *WebSocketServer
	httpServer    *http.Server
	apiHandler    *APIHandler

	// Dashboard data
	frames        []*DashboardFrame
	components    map[string]*DashboardComponent
	metrics       *DashboardMetrics

	// Real-time updates
	updateChan    chan *DashboardUpdate
}

// DashboardFrame represents a frame in the dashboard.
type DashboardFrame struct {
	FrameID      devtools.FrameID `json:"frameId"`
	Timestamp    time.Time        `json:"timestamp"`
	Duration     time.Duration    `json:"duration"`
	EventCount   int              `json:"eventCount"`
	MutationCount int             `json:"mutationCount"`
	LayoutCount  int              `json:"layoutCount"`
	RepaintCount int              `json:"repaintCount"`
}

// DashboardComponent represents a component in the dashboard.
type DashboardComponent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Styles     map[string]interface{} `json:"styles"`
	Children   []string               `json:"children"`
}

// DashboardMetrics represents dashboard metrics.
type DashboardMetrics struct {
	FPS            float64       `json:"fps"`
	FrameTime      time.Duration `json:"frameTime"`
	LayoutTime     time.Duration `json:"layoutTime"`
	PaintTime      time.Duration `json:"paintTime"`
	MemoryUsage    uint64        `json:"memoryUsage"`
	ComponentCount int           `json:"componentCount"`
}

// DashboardUpdate represents a dashboard update.
type DashboardUpdate struct {
	Type      string
	Timestamp time.Time
	Data      interface{}
}

// NewWebDashboard creates a new web dashboard.
func NewWebDashboard(port int) *WebDashboard {
	wsServer := NewWebSocketServer(port)
	apiHandler := NewAPIHandler(nil) // Will be set after dashboard creation

	dashboard := &WebDashboard{
		port:       port,
		wsServer:   wsServer,
		apiHandler: apiHandler,
		frames:     make([]*DashboardFrame, 0, 100),
		components: make(map[string]*DashboardComponent),
		metrics:    &DashboardMetrics{},
		updateChan: make(chan *DashboardUpdate, 256),
	}

	// Set dashboard reference in API handler
	apiHandler.dashboard = dashboard

	return dashboard
}

// Start starts the web dashboard.
func (wd *WebDashboard) Start() error {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	if wd.enabled {
		return fmt.Errorf("dashboard already running")
	}

	wd.enabled = true

	// Start WebSocket server (marks as ready, doesn't start HTTP)
	if err := wd.wsServer.Start(); err != nil {
		wd.enabled = false
		return fmt.Errorf("failed to start WebSocket server: %w", err)
	}

	// Start update handler
	go wd.updateLoop()

	// Create HTTP server with all routes
	mux := http.NewServeMux()

	// Serve dashboard HTML
	mux.HandleFunc("/", wd.handleDashboard)
	mux.HandleFunc("/debug", wd.handleDashboard)

	// WebSocket endpoint
	mux.Handle("/ws", wd.wsServer.Handler())

	// API endpoints
	mux.HandleFunc("/api/frames", wd.handleAPIFrames)
	mux.HandleFunc("/api/metrics", wd.handleAPIMetrics)
	mux.HandleFunc("/api/components", wd.handleAPIComponents)
	mux.HandleFunc("/api/report", wd.handleAPIReport)
	mux.HandleFunc("/api/export", wd.handleAPIExport)
	mux.HandleFunc("/api/import", wd.handleAPIImport)
	mux.HandleFunc("/health", wd.handleHealth)

	wd.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", wd.port),
		Handler: mux,
	}

	go func() {
		log.Printf("[WebDashboard] Server started on http://localhost:%d", wd.port)
		log.Printf("[WebDashboard]  Dashboard: http://localhost:%d/", wd.port)
		log.Printf("[WebDashboard]  WebSocket: ws://localhost:%d/ws", wd.port)
		if err := wd.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[WebDashboard] Server error: %v", err)
		}
	}()

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

	// Stop WebSocket server
	if err := wd.wsServer.Stop(); err != nil {
		log.Printf("[WebDashboard] Error stopping WebSocket server: %v", err)
	}

	// Stop HTTP server
	if wd.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := wd.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[WebDashboard] Error stopping HTTP server: %v", err)
		}
	}

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

// GetWebSocketServer returns the WebSocket server.
func (wd *WebDashboard) GetWebSocketServer() *WebSocketServer {
	return wd.wsServer
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
		wd.wsServer.BroadcastUpdate(update.Type, update.Data)
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
		ClientInfo: &ClientInfo{
			ID:         clientID,
			ConnectedAt: time.Now(),
		},
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
		"client_id":    rds.ClientInfo.ID,
	}
}

// =============================================================================
// HTTP Handlers for WebDashboard
// =============================================================================

// handleDashboard serves the dashboard HTML page.
func (wd *WebDashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Try to read from embedded file, fallback to inline template
	html := getDashboardHTML()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// getDashboardHTML returns the dashboard HTML content.
func getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mint DevTools Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #1e1e1e; color: #d4d4d4; }
        .header { background: #252526; padding: 10px 20px; border-bottom: 1px solid #3e3e42; display: flex; justify-content: space-between; align-items: center; }
        .header h1 { font-size: 18px; font-weight: 500; }
        .status { display: flex; gap: 20px; font-size: 12px; }
        .status-item { display: flex; align-items: center; gap: 5px; }
        .status-dot { width: 8px; height: 8px; border-radius: 50%; background: #4ec9b0; }
        .status-dot.disconnected { background: #f48771; }
        .container { display: flex; height: calc(100vh - 45px); }
        .sidebar { width: 200px; background: #252526; border-right: 1px solid #3e3e42; padding: 10px; }
        .sidebar-item { padding: 8px 12px; cursor: pointer; border-radius: 4px; margin-bottom: 2px; }
        .sidebar-item:hover, .sidebar-item.active { background: #37373d; }
        .main { flex: 1; padding: 20px; overflow-y: auto; }
        .card { background: #252526; border: 1px solid #3e3e42; border-radius: 6px; padding: 15px; margin-bottom: 15px; }
        .card h3 { font-size: 14px; margin-bottom: 10px; color: #9cdcfe; }
        .metric-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
        .metric { background: #1e1e1e; padding: 10px; border-radius: 4px; text-align: center; }
        .metric-value { font-size: 24px; font-weight: bold; color: #4ec9b0; }
        .metric-label { font-size: 11px; color: #858585; margin-top: 5px; }
        .table { width: 100%; border-collapse: collapse; font-size: 12px; }
        .table th { text-align: left; padding: 8px; background: #1e1e1e; border-bottom: 1px solid #3e3e42; }
        .table td { padding: 8px; border-bottom: 1px solid #2d2d2d; }
        .table tr:hover { background: #2a2a2a; }
        .empty-state { text-align: center; padding: 40px; color: #858585; }
        .hidden { display: none; }
        .json-preview { background: #1e1e1e; padding: 10px; border-radius: 4px; font-family: monospace; font-size: 11px; overflow-x: auto; }
        .component-item { padding: 8px; background: #1e1e1e; margin-bottom: 5px; border-radius: 4px; }
        .component-type { color: #4ec9b0; font-weight: bold; }
        .component-id { color: #9cdcfe; font-size: 11px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Mint DevTools Dashboard</h1>
        <div class="status">
            <div class="status-item">
                <span class="status-dot disconnected" id="ws-status"></span>
                <span id="ws-text">Connecting...</span>
            </div>
            <div class="status-item">
                <span>Frames:</span>
                <span id="frame-count">0</span>
            </div>
        </div>
    </div>
    <div class="container">
        <div class="sidebar">
            <div class="sidebar-item active" data-view="dashboard">Dashboard</div>
            <div class="sidebar-item" data-view="frames">Frames</div>
            <div class="sidebar-item" data-view="components">Components</div>
            <div class="sidebar-item" data-view="report">Report</div>
        </div>
        <div class="main" id="main-content">
            <div id="view-dashboard" class="view">
                <div class="card">
                    <h3>Performance Metrics</h3>
                    <div class="metric-grid">
                        <div class="metric"><div class="metric-value" id="fps">--</div><div class="metric-label">FPS</div></div>
                        <div class="metric"><div class="metric-value" id="frame-time">--</div><div class="metric-label">Frame Time (ms)</div></div>
                        <div class="metric"><div class="metric-value" id="memory">--</div><div class="metric-label">Memory (MB)</div></div>
                        <div class="metric"><div class="metric-value" id="components-count">--</div><div class="metric-label">Components</div></div>
                    </div>
                </div>
                <div class="card">
                    <h3>Recent Frames</h3>
                    <div id="frames-empty" class="empty-state">Waiting for data...</div>
                    <table class="table" id="frames-table-container" style="display:none;">
                        <thead><tr><th>Frame ID</th><th>Time</th><th>Duration</th><th>Events</th><th>Mutations</th></tr></thead>
                        <tbody id="frames-table"></tbody>
                    </table>
                </div>
            </div>
            <div id="view-frames" class="view hidden">
                <div class="card">
                    <h3>All Frames</h3>
                    <div id="all-frames-empty" class="empty-state">No frames captured yet</div>
                    <table class="table" id="all-frames-table-container" style="display:none;">
                        <thead><tr><th>Frame ID</th><th>Time</th><th>Duration</th><th>Events</th><th>Mutations</th><th>Layouts</th><th>Repaints</th></tr></thead>
                        <tbody id="all-frames-table"></tbody>
                    </table>
                </div>
            </div>
            <div id="view-components" class="view hidden">
                <div class="card">
                    <h3>Components</h3>
                    <div id="components-empty" class="empty-state">No components yet</div>
                    <div id="components-list"></div>
                </div>
            </div>
            <div id="view-report" class="view hidden">
                <div class="card">
                    <h3>Debug Report</h3>
                    <div id="report-content" class="json-preview">Loading...</div>
                </div>
            </div>
        </div>
    </div>
    <script>
        let currentView='dashboard',allFrames=[],allComponents={},currentMetrics=null;
        const ws=new WebSocket('ws://'+window.location.host+'/ws');
        ws.onopen=()=>{document.getElementById('ws-status').classList.remove('disconnected');document.getElementById('ws-text').textContent='Connected';};
        ws.onclose=()=>{document.getElementById('ws-status').classList.add('disconnected');document.getElementById('ws-text').textContent='Disconnected';};
        ws.onmessage=(e)=>{try{const msg=JSON.parse(e.data);if(msg.type==='frame')addFrame(msg.data);if(msg.type==='metrics')updateMetrics(msg.data);if(msg.type==='component')updateComponent(msg.data);}catch(err){}};
        document.querySelectorAll('.sidebar-item').forEach(item=>{item.addEventListener('click',function(e){const v=item.getAttribute('data-view');switchView(v);e.preventDefault();});});
        function switchView(n){document.querySelectorAll('.sidebar-item').forEach(i=>i.classList.remove('active'));document.querySelector('[data-view="'+n+'"]').classList.add('active');document.querySelectorAll('.view').forEach(v=>v.classList.add('hidden'));const t=document.getElementById('view-'+n);if(t)t.classList.remove('hidden');currentView=n;if(n==='frames')renderAllFrames();if(n==='components')renderComponents();if(n==='report')renderReport();}
        function updateMetrics(d){currentMetrics=d;document.getElementById('fps').textContent=(d.fps||0).toFixed(1);document.getElementById('frame-time').textContent=d.frameTime?(d.frameTime/1000000).toFixed(2):'--';document.getElementById('memory').textContent=d.memoryUsage?(d.memoryUsage/1024/1024).toFixed(1):'--';document.getElementById('components-count').textContent=d.componentCount||Object.keys(allComponents).length;}
        function addFrame(f){allFrames.push(f);if(allFrames.length>100)allFrames.shift();document.getElementById('frame-count').textContent=allFrames.length;const tc=document.getElementById('frames-table-container'),es=document.getElementById('frames-empty');if(tc)tc.style.display='';if(es)es.style.display='none';const tb=document.getElementById('frames-table');if(tb){const r=tb.insertRow(0);r.innerHTML='<td>'+(f.frameId||'--')+'</td><td>'+(f.timestamp?new Date(f.timestamp).toLocaleTimeString():'--')+'</td><td>'+(f.duration?(f.duration/1000000).toFixed(2)+'ms':'--')+'</td><td>'+(f.eventCount||0)+'</td><td>'+(f.mutationCount||0)+'</td>';if(tb.rows.length>10)tb.deleteRow(10);}if(currentView==='frames')renderAllFrames();}
        function renderAllFrames(){const tc=document.getElementById('all-frames-table-container'),es=document.getElementById('all-frames-empty'),tb=document.getElementById('all-frames-table');if(!tb)return;tb.innerHTML='';if(allFrames.length===0){if(tc)tc.style.display='none';if(es)es.style.display='';return;}if(tc)tc.style.display='';if(es)es.style.display='none';[...allFrames].reverse().forEach(f=>{const r=tb.insertRow();r.innerHTML='<td>'+(f.frameId||'--')+'</td><td>'+(f.timestamp?new Date(f.timestamp).toLocaleTimeString():'--')+'</td><td>'+(f.duration?(f.duration/1000000).toFixed(2)+'ms':'--')+'</td><td>'+(f.eventCount||0)+'</td><td>'+(f.mutationCount||0)+'</td><td>'+(f.layoutCount||0)+'</td><td>'+(f.repaintCount||0)+'</td>';});}
        function updateComponent(d){if(d&&d.id){allComponents[d.id]=d;if(currentView==='components')renderComponents();if(currentMetrics){currentMetrics.componentCount=Object.keys(allComponents).length;document.getElementById('components-count').textContent=currentMetrics.componentCount;}}}
        function renderComponents(){const l=document.getElementById('components-list'),es=document.getElementById('components-empty');if(!l)return;l.innerHTML='';const ids=Object.keys(allComponents);if(ids.length===0){if(es)es.style.display='';return;}if(es)es.style.display='none';ids.forEach(id=>{const c=allComponents[id],div=document.createElement('div');div.className='component-item';div.innerHTML='<div class="component-type">'+(c.type||'Unknown')+'</div><div class="component-id">'+id+'</div><div style="font-size:11px;color:#858585;margin-top:4px;">Props: '+JSON.stringify(c.properties||{}).substring(0,50)+'...</div>';l.appendChild(div);});}
        function renderReport(){const c=document.getElementById('report-content');if(!c)return;c.textContent=JSON.stringify({timestamp:new Date().toISOString(),frames:{total:allFrames.length,recent:allFrames.slice(-10)},components:{total:Object.keys(allComponents).length,ids:Object.keys(allComponents)},metrics:currentMetrics},null,2);}
        function loadInitialData(){fetch('/api/metrics').then(r=>r.json()).then(updateMetrics).catch(e=>console.error(e));fetch('/api/frames').then(r=>r.json()).then(d=>{if(Array.isArray(d)){allFrames=d;d.forEach(f=>addFrame(f));}}).catch(e=>console.error(e));fetch('/api/components').then(r=>r.json()).then(d=>{if(typeof d==='object')for(let id in d)allComponents[id]=d[id];}).catch(e=>console.error(e));}
        loadInitialData();console.log('[Dashboard] Ready');
    </script>
</body>
</html>`
}

// handleAPIFrames handles GET /api/frames.
func (wd *WebDashboard) handleAPIFrames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	frames := wd.GetFrames()
	json.NewEncoder(w).Encode(frames)
}

// handleAPIMetrics handles GET /api/metrics.
func (wd *WebDashboard) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metrics := wd.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// handleAPIComponents handles GET /api/components.
func (wd *WebDashboard) handleAPIComponents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	wd.mu.RLock()
	components := make(map[string]*DashboardComponent)
	for k, v := range wd.components {
		components[k] = v
	}
	wd.mu.RUnlock()
	json.NewEncoder(w).Encode(components)
}

// handleAPIReport handles GET /api/report.
func (wd *WebDashboard) handleAPIReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	report := wd.GenerateReport()
	json.NewEncoder(w).Encode(report)
}

// handleAPIExport handles GET /api/export.
func (wd *WebDashboard) handleAPIExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := wd.ExportData()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}

// handleAPIImport handles POST /api/import.
func (wd *WebDashboard) handleAPIImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	dataBytes, _ := json.Marshal(data)
	if err := wd.ImportData(dataBytes); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "imported"})
}

// handleHealth handles GET /health.
func (wd *WebDashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"server":       "mint-webdashboard",
		"version":      "1.0.0",
		"enabled":      wd.IsRunning(),
		"ws_clients":   wd.wsServer.GetClientCount(),
		"frame_count":  len(wd.GetFrames()),
		"component_count": len(wd.components),
	})
}

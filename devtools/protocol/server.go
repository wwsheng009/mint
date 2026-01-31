// Package protocol provides unified DevTools server for both TUI and remote debugging.
//
// This is the main server that combines:
//   - HTTP API for snapshots, diffs, metrics
//   - WebSocket server for real-time updates
//   - Web dashboard UI
//   - Support for both TUI and remote debugging protocols
//
// Usage:
//
//	import "github.com/wwsheng009/mint/devtools/protocol"
//
//	server := protocol.NewServer(protocol.ServerConfig{
//	    Port: 8080,
//	    EnableDashboard: true,
//	    EnableTuiCommands: true,
//	})
//	server.SetSnapshotManager(snapshotManager)
//	server.Start()
package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/snapshot"
)

// =============================================================================
// Server Configuration
// =============================================================================

// ServerConfig configures the DevTools server.
type ServerConfig struct {
	Port              int
	EnableDashboard   bool // Enable web dashboard UI
	EnableTuiCommands bool // Enable TUI-specific commands (inspect, highlight, replay)
	EnableCdp         bool // Enable Chrome DevTools Protocol compatibility
}

// DefaultServerConfig returns default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:              8080,
		EnableDashboard:   true,
		EnableTuiCommands: true,
		EnableCdp:         true,
	}
}

// =============================================================================
// Server - Unified DevTools Server
// =============================================================================

// Server is the unified DevTools server.
type Server struct {
	mu              sync.RWMutex
	config          ServerConfig
	running         bool
	httpServer      *http.Server
	wsServer        *WebSocketServer
	snapshotManager *snapshot.Manager
	devtools        *devtools.DevTools

	// Metrics and dashboard data
	metrics         *Metrics
	frames          []*FrameData
	components      map[string]*DashboardComponentData

	// Shutdown handling
	shutdownCtx     context.Context
	shutdownCancel  context.CancelFunc
}

// Metrics represents performance metrics.
type Metrics struct {
	Timestamp      time.Time     `json:"timestamp"`
	FPS            float64       `json:"fps"`
	FrameTime      time.Duration `json:"-"` // Custom marshaled to ms
	LayoutTime     time.Duration `json:"-"` // Custom marshaled to ms
	PaintTime      time.Duration `json:"-"` // Custom marshaled to ms
	MemoryUsage    uint64        `json:"memoryUsage"`
	ComponentCount int           `json:"componentCount"`
	FrameCount     int           `json:"frameCount"`
	// Exported as milliseconds for JSON
	FrameTimeMs    int64 `json:"frameTime"`
	LayoutTimeMs   int64 `json:"layoutTime"`
	PaintTimeMs    int64 `json:"paintTime"`
}

// MarshalJSON implements json.Marshaler for Metrics.
func (m *Metrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Timestamp      time.Time `json:"timestamp"`
		FPS            float64   `json:"fps"`
	FrameTimeMs    int64     `json:"frameTime"`
		LayoutTimeMs   int64     `json:"layoutTime"`
		PaintTimeMs    int64     `json:"paintTime"`
		MemoryUsage    uint64     `json:"memoryUsage"`
		ComponentCount int        `json:"componentCount"`
		FrameCount     int        `json:"frameCount"`
	}{
		Timestamp:      m.Timestamp,
		FPS:            m.FPS,
		FrameTimeMs:    m.FrameTime.Milliseconds(),
		LayoutTimeMs:   m.LayoutTime.Milliseconds(),
		PaintTimeMs:    m.PaintTime.Milliseconds(),
		MemoryUsage:    m.MemoryUsage,
		ComponentCount: m.ComponentCount,
		FrameCount:     m.FrameCount,
	})
}

// FrameData represents frame information.
type FrameData struct {
	FrameID       devtools.FrameID `json:"frameId"`
	Timestamp     time.Time        `json:"timestamp"`
	Duration      time.Duration    `json:"-"` // Custom marshaled
	EventCount    int              `json:"eventCount"`
	MutationCount int              `json:"mutationCount"`
	LayoutCount   int              `json:"layoutCount"`
	RepaintCount  int              `json:"repaintCount"`
	DurationMs    int64            `json:"durationMs"` // For JSON serialization
}

// MarshalJSON implements json.Marshaler for FrameData.
func (f *FrameData) MarshalJSON() ([]byte, error) {
	type Alias FrameData
	return json.Marshal(&struct {
		FrameID       devtools.FrameID `json:"frameId"`
		Timestamp     time.Time        `json:"timestamp"`
		EventCount    int              `json:"eventCount"`
		MutationCount int              `json:"mutationCount"`
		LayoutCount   int              `json:"layoutCount"`
		RepaintCount  int              `json:"repaintCount"`
		DurationMs    int64            `json:"durationMs"`
	}{
		FrameID:       f.FrameID,
		Timestamp:     f.Timestamp,
		EventCount:    f.EventCount,
		MutationCount: f.MutationCount,
		LayoutCount:   f.LayoutCount,
		RepaintCount:  f.RepaintCount,
		DurationMs:    f.Duration.Milliseconds(),
	})
}

// UnmarshalJSON implements json.Unmarshaler for FrameData.
func (f *FrameData) UnmarshalJSON(data []byte) error {
	type Alias FrameData
	aux := &struct {
		FrameID       devtools.FrameID `json:"frameId"`
		Timestamp     time.Time        `json:"timestamp"`
		EventCount    int              `json:"eventCount"`
		MutationCount int              `json:"mutationCount"`
		LayoutCount   int              `json:"layoutCount"`
		RepaintCount  int              `json:"repaintCount"`
		DurationMs    int64            `json:"durationMs"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	f.FrameID = aux.FrameID
	f.Timestamp = aux.Timestamp
	f.EventCount = aux.EventCount
	f.MutationCount = aux.MutationCount
	f.LayoutCount = aux.LayoutCount
	f.RepaintCount = aux.RepaintCount
	f.Duration = time.Duration(aux.DurationMs) * time.Millisecond
	return nil
}

// DashboardComponentData represents component state for the dashboard.
// This is the client-facing version, distinct from protocol.ComponentData.
type DashboardComponentData struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Styles     map[string]interface{} `json:"styles"`
	Children   []string               `json:"children"`
	Visible    bool                   `json:"visible"`
	Focused    bool                   `json:"focused"`
	Bounds     RectData               `json:"bounds"`
}

// Report represents a debug report.
type Report struct {
	GeneratedAt time.Time                 `json:"generatedAt"`
	Metrics     *Metrics                  `json:"metrics"`
	Frames      []*FrameSummary           `json:"frames"`
	Components  []*DashboardComponentData `json:"components"`
}

// NewServer creates a new DevTools server.
func NewServer(config ServerConfig) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	wsConfig := WebSocketServerConfig{
		Port:              config.Port,
		Path:              "/ws",
		EnableCdp:         config.EnableCdp,
		EnableTuiCommands: config.EnableTuiCommands,
		HeartbeatInterval: 30,
	}

	return &Server{
		config:         config,
		wsServer:       NewWebSocketServer(wsConfig),
		metrics:        &Metrics{Timestamp: time.Now()},
		frames:         make([]*FrameData, 0, 100),
		components:     make(map[string]*DashboardComponentData),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}
}

// SetSnapshotManager sets the snapshot manager.
func (s *Server) SetSnapshotManager(sm *snapshot.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshotManager = sm

	// Configure WebSocket server with snapshot handlers
	s.wsServer.SetSnapshotHandler(func(frameID devtools.FrameID) (*snapshot.Snapshot, bool) {
		return sm.Get(frameID)
	})

	s.wsServer.SetRangeHandler(func(from, to devtools.FrameID) []*snapshot.Snapshot {
		return sm.GetRange(from, to)
	})
}

// SetDevTools sets the DevTools instance.
func (s *Server) SetDevTools(dt *devtools.DevTools) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devtools = dt
}

// Start starts the server.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running on port %d", s.config.Port)
	}

	s.running = true

	// Create HTTP multiplexer
	mux := http.NewServeMux()

	// Dashboard UI
	if s.config.EnableDashboard {
		mux.HandleFunc("/", s.handleDashboard)
		mux.HandleFunc("/debug", s.handleDebugInspector)
	}

	// WebSocket endpoint
	mux.Handle("/ws", s.wsServer.Handler())

	// API endpoints
	s.setupAPIRoutes(mux)

	// Start HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("[DevTools] Server started on http://localhost:%d", s.config.Port)
		if s.config.EnableDashboard {
			log.Printf("[DevTools]   Dashboard: http://localhost:%d/", s.config.Port)
		}
		log.Printf("[DevTools]   Inspector: http://localhost:%d/debug", s.config.Port)
		log.Printf("[DevTools]   WebSocket: ws://localhost:%d/ws", s.config.Port)
		log.Printf("[DevTools]   API:       http://localhost:%d/api/*", s.config.Port)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[DevTools] Server error: %v", err)
		}
	}()

	return nil
}

// StartInBackground starts the server in a goroutine.
func (s *Server) StartInBackground() error {
	return s.Start()
}

// Stop stops the server.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false

	// Cancel shutdown context
	s.shutdownCancel()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[DevTools] Shutdown error: %v", err)
		return err
	}

	log.Printf("[DevTools] Server stopped")
	return nil
}

// IsRunning returns whether the server is running.
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetPort returns the server port.
func (s *Server) GetPort() int {
	return s.config.Port
}

// GetWebSocketServer returns the WebSocket server.
func (s *Server) GetWebSocketServer() *WebSocketServer {
	return s.wsServer
}

// =============================================================================
// API Routes
// =============================================================================

func (s *Server) setupAPIRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// Snapshot API (if snapshot manager is set)
	mux.HandleFunc("/api/snapshots", s.handleGetSnapshots)
	mux.HandleFunc("/api/snapshot/", s.handleGetSnapshot)
	mux.HandleFunc("/api/diff", s.handleGetDiff)

	// Dashboard data API
	mux.HandleFunc("/api/frames", s.handleGetFrames)
	mux.HandleFunc("/api/metrics", s.handleGetMetrics)
	mux.HandleFunc("/api/components", s.handleGetComponents)
	mux.HandleFunc("/api/report", s.handleGetReport)

	// Import/Export
	mux.HandleFunc("/api/export", s.handleExport)
	mux.HandleFunc("/api/import", s.handleImport)
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	s.mu.RLock()
	response := map[string]interface{}{
		"status":       "ok",
		"server":       "mint-devtools",
		"version":      Version,
		"port":         s.config.Port,
		"ws_clients":   s.wsServer.GetClientCount(),
		"running":      s.running,
		"capabilities": []string{"snapshots", "diffs", "metrics", "frames", "components"},
	}

	if s.config.EnableTuiCommands {
		response["capabilities"] = append(response["capabilities"].([]string), "tui_commands")
	}

	if s.snapshotManager != nil {
		response["snapshots"] = s.snapshotManager.GetStats()
	}

	s.mu.RUnlock()

	json.NewEncoder(w).Encode(response)
}

// handleDashboard serves the main dashboard UI.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDashboardHTML())
}

// handleDebugInspector serves the debug inspector UI.
func (s *Server) handleDebugInspector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getInspectorHTML())
}

// handleGetSnapshots returns all snapshots.
func (s *Server) handleGetSnapshots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.snapshotManager == nil {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	snapshots := s.snapshotManager.GetAll()
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
}

// handleGetSnapshot returns a specific snapshot.
func (s *Server) handleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.snapshotManager == nil {
		http.Error(w, "Snapshot manager not configured", 500)
		return
	}

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

	// Convert to protocol format
	components := make([]DashboardComponentData, 0, len(snap.States))
	for _, state := range snap.States {
		comp := DashboardComponentData{
			ID:       string(state.NodeID),
			Type:     state.Type,
			Properties: state.Props,
			Styles: map[string]interface{}{
				"x":      state.Bounds.X,
				"y":      state.Bounds.Y,
				"width":  state.Bounds.Width,
				"height": state.Bounds.Height,
			},
			Children: nodeIDsToStrings(state.Children),
			Visible:  state.Visible,
			Focused:  state.Focused,
		}
		components = append(components, comp)
	}

	result := map[string]interface{}{
		"frame_id":   int(snap.FrameID),
		"timestamp":  snap.Timestamp.Format(time.RFC3339),
		"components": components,
	}

	json.NewEncoder(w).Encode(result)
}

// handleGetDiff returns the diff between two snapshots.
func (s *Server) handleGetDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.snapshotManager == nil {
		http.Error(w, "Snapshot manager not configured", 500)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing from/to parameters"})
		return
	}

	var fromInt, toInt int
	_, fromErr := fmt.Sscanf(fromStr, "%d", &fromInt)
	_, toErr := fmt.Sscanf(toStr, "%d", &toInt)

	if fromErr != nil || toErr != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid frame IDs"})
		return
	}

	fromSnap, fromOk := s.snapshotManager.Get(devtools.FrameID(fromInt))
	toSnap, toOk := s.snapshotManager.Get(devtools.FrameID(toInt))

	if !fromOk || !toOk {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Snapshot not found",
		})
		return
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
		From:    devtools.FrameID(fromInt),
		To:      devtools.FrameID(toInt),
		Changes: changes,
	}

	json.NewEncoder(w).Encode(result)
}

// handleGetFrames returns all frames.
func (s *Server) handleGetFrames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.RLock()
	defer s.mu.RUnlock()
	json.NewEncoder(w).Encode(s.frames)
}

// handleGetMetrics returns current metrics.
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Manually construct JSON to ensure proper serialization
	metrics := s.metrics
	data := map[string]interface{}{
		"timestamp":      metrics.Timestamp,
		"fps":            metrics.FPS,
		"frameTime":      metrics.FrameTime.Milliseconds(),
		"layoutTime":     metrics.LayoutTime.Milliseconds(),
		"paintTime":      metrics.PaintTime.Milliseconds(),
		"memoryUsage":    metrics.MemoryUsage,
		"componentCount": metrics.ComponentCount,
		"frameCount":     metrics.FrameCount,
	}
	json.NewEncoder(w).Encode(data)
}

// handleGetComponents returns all components.
func (s *Server) handleGetComponents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DashboardComponentData, 0, len(s.components))
	for _, comp := range s.components {
		result = append(result, comp)
	}
	json.NewEncoder(w).Encode(result)
}

// handleGetReport generates a debug report.
func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Generate frame summaries
	frames := make([]*FrameSummary, 0, len(s.frames))
	for _, f := range s.frames {
		frames = append(frames, &FrameSummary{
			FrameID:   f.FrameID,
			Timestamp: f.Timestamp,
			Events:    f.EventCount,
			Mutations: f.MutationCount,
			Layouts:   f.LayoutCount,
		})
	}

	// Get components
	components := make([]*DashboardComponentData, 0, len(s.components))
	for _, comp := range s.components {
		components = append(components, comp)
	}

	report := Report{
		GeneratedAt: time.Now(),
		Metrics:     s.metrics,
		Frames:      frames,
		Components:  components,
	}

	json.NewEncoder(w).Encode(report)
}

// handleExport exports all dashboard data.
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := map[string]interface{}{
		"version":    Version,
		"exported_at": time.Now().Format(time.RFC3339),
		"metrics":     s.metrics,
		"frames":      s.frames,
		"components":  s.components,
	}

	json.NewEncoder(w).Encode(data)
}

// handleImport imports dashboard data.
func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Import metrics
	if metricsData, ok := data["metrics"].(map[string]interface{}); ok {
		s.metrics = importMetrics(metricsData)
	}

	// Import frames
	if framesData, ok := data["frames"].([]interface{}); ok {
		s.frames = importFrames(framesData)
	}

	// Import components
	if componentsData, ok := data["components"].([]interface{}); ok {
		s.components = importComponents(componentsData)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"imported": map[string]interface{}{
			"frames_count":     len(s.frames),
			"components_count": len(s.components),
		},
	})
}

// =============================================================================
// Public API for updating data
// =============================================================================

// AddFrame adds a frame to the dashboard.
func (s *Server) AddFrame(frame *FrameData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.frames = append(s.frames, frame)

	// Keep only last 100 frames
	if len(s.frames) > 100 {
		s.frames = s.frames[len(s.frames)-100:]
	}

	// Broadcast via WebSocket
	s.wsServer.Broadcast(map[string]interface{}{
		"type": "frame_added",
		"data": frame,
	})
}

// GetFrames returns all frames.
func (s *Server) GetFrames() []*FrameData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.frames
}

// GetFrame returns a specific frame.
func (s *Server) GetFrame(frameID devtools.FrameID) *FrameData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.frames {
		if f.FrameID == frameID {
			return f
		}
	}
	return nil
}

// UpdateComponent updates a component's state.
func (s *Server) UpdateComponent(id string, comp *DashboardComponentData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.components[id] = comp

	// Broadcast via WebSocket
	s.wsServer.Broadcast(map[string]interface{}{
		"type": "component_updated",
		"data": comp,
	})
}

// GetComponents returns all components.
func (s *Server) GetComponents() map[string]*DashboardComponentData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*DashboardComponentData, len(s.components))
	for k, v := range s.components {
		result[k] = v
	}
	return result
}

// GetComponent returns a specific component.
func (s *Server) GetComponent(id string) (*DashboardComponentData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	comp, ok := s.components[id]
	return comp, ok
}

// UpdateMetrics updates the performance metrics.
func (s *Server) UpdateMetrics(metrics *Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = metrics
	s.metrics.Timestamp = time.Now()

	// Debug logging
	log.Printf("[DevTools] UpdateMetrics: FPS=%.1f FrameTime=%dms ComponentCount=%d",
		metrics.FPS, metrics.FrameTime.Milliseconds(), metrics.ComponentCount)

	// Broadcast via WebSocket - convert to map to ensure proper serialization
	data := map[string]interface{}{
		"timestamp":      metrics.Timestamp,
		"fps":            metrics.FPS,
		"frameTime":      metrics.FrameTime.Milliseconds(),
		"layoutTime":     metrics.LayoutTime.Milliseconds(),
		"paintTime":      metrics.PaintTime.Milliseconds(),
		"memoryUsage":    metrics.MemoryUsage,
		"componentCount": metrics.ComponentCount,
		"frameCount":     metrics.FrameCount,
	}
	s.wsServer.Broadcast(map[string]interface{}{
		"type": "metrics_updated",
		"data": data,
	})
}

// GetMetrics returns current metrics.
func (s *Server) GetMetrics() *Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}

// GenerateReport generates a debug report.
func (s *Server) GenerateReport() *Report {
	s.mu.RLock()
	defer s.mu.RUnlock()

	frames := make([]*FrameSummary, 0, len(s.frames))
	for _, f := range s.frames {
		frames = append(frames, &FrameSummary{
			FrameID:   f.FrameID,
			Timestamp: f.Timestamp,
			Events:    f.EventCount,
			Mutations: f.MutationCount,
			Layouts:   f.LayoutCount,
		})
	}

	components := make([]*DashboardComponentData, 0, len(s.components))
	for _, comp := range s.components {
		components = append(components, comp)
	}

	return &Report{
		GeneratedAt: time.Now(),
		Metrics:     s.metrics,
		Frames:      frames,
		Components:  components,
	}
}

// =============================================================================
// Utilities
// =============================================================================

func nodeIDsToStrings(ids []devtools.NodeID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = string(id)
	}
	return result
}

func importMetrics(data map[string]interface{}) *Metrics {
	m := &Metrics{Timestamp: time.Now()}
	if fps, ok := data["fps"].(float64); ok {
		m.FPS = fps
	}
	if mem, ok := data["memoryUsage"].(float64); ok {
		m.MemoryUsage = uint64(mem)
	}
	if cc, ok := data["componentCount"].(float64); ok {
		m.ComponentCount = int(cc)
	}
	if fc, ok := data["frameCount"].(float64); ok {
		m.FrameCount = int(fc)
	}
	return m
}

func importFrames(data []interface{}) []*FrameData {
	frames := make([]*FrameData, 0, len(data))
	for _, item := range data {
		if frameMap, ok := item.(map[string]interface{}); ok {
			f := &FrameData{}
			if fid, ok := frameMap["frameId"].(float64); ok {
				f.FrameID = devtools.FrameID(fid)
			}
			if ts, ok := frameMap["timestamp"].(string); ok {
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					f.Timestamp = t
				}
			}
			if ec, ok := frameMap["eventCount"].(float64); ok {
				f.EventCount = int(ec)
			}
			if mc, ok := frameMap["mutationCount"].(float64); ok {
				f.MutationCount = int(mc)
			}
			if lc, ok := frameMap["layoutCount"].(float64); ok {
				f.LayoutCount = int(lc)
			}
			if rc, ok := frameMap["repaintCount"].(float64); ok {
				f.RepaintCount = int(rc)
			}
			frames = append(frames, f)
		}
	}
	return frames
}

func importComponents(data []interface{}) map[string]*DashboardComponentData {
	components := make(map[string]*DashboardComponentData)
	for _, item := range data {
		if compMap, ok := item.(map[string]interface{}); ok {
			c := &DashboardComponentData{}
			if id, ok := compMap["id"].(string); ok {
				c.ID = id
			}
			if typ, ok := compMap["type"].(string); ok {
				c.Type = typ
			}
			if props, ok := compMap["properties"].(map[string]interface{}); ok {
				c.Properties = props
			}
			if styles, ok := compMap["styles"].(map[string]interface{}); ok {
				c.Styles = styles
			}
			if visible, ok := compMap["visible"].(bool); ok {
				c.Visible = visible
			}
			if focused, ok := compMap["focused"].(bool); ok {
				c.Focused = focused
			}
			if c.ID != "" {
				components[c.ID] = c
			}
		}
	}
	return components
}

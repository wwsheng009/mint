// Package remote provides remote debugging support for DevTools.
package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/protocol"
	"github.com/wwsheng009/mint/devtools/snapshot"
)

// =============================================================================
// HTTP Server for Remote Debugging (no WebSocket dependency)
// =============================================================================

// HTTPServer provides HTTP API for remote debugging.
type HTTPServer struct {
	mu              sync.RWMutex
	bridge          *ChromiumBridge
	snapshotManager *snapshot.Manager
	port            int
	mux             *http.ServeMux
}

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(port int, dt *devtools.DevTools, sm *snapshot.Manager) *HTTPServer {
	bridge := NewChromiumBridge(dt, sm)

	return &HTTPServer{
		bridge:          bridge,
		snapshotManager: sm,
		port:            port,
		mux:             http.NewServeMux(),
	}
}

// Start starts the HTTP server.
func (s *HTTPServer) Start() error {
	s.setupRoutes()

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("DevTools HTTP server starting on %s", addr)
	log.Printf("  Inspector: http://localhost%s/debug", addr)
	log.Printf("  API:       http://localhost%s/api/snapshots", addr)
	log.Printf("  Diff:      http://localhost%s/api/diff?from=X&to=Y", addr)
	log.Printf("  Health:    http://localhost%s/health", addr)

	return http.ListenAndServe(addr, s.mux)
}

// StartInBackground starts the server in a goroutine.
func (s *HTTPServer) StartInBackground() {
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
}

// setupRoutes configures HTTP routes.
func (s *HTTPServer) setupRoutes() {
	// Serve inspector HTML page
	s.mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, s.bridge.GetInspectorHTML())
	})

	// Health check
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"server":    "mint-devtools",
			"version":   protocol.Version,
			"snapshots": s.snapshotManager.GetStats().TotalSnapshots,
		})
	})

	// API: Get all snapshots
	s.mux.HandleFunc("/api/snapshots", func(w http.ResponseWriter, r *http.Request) {
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

	// API: Get specific snapshot by frame ID
	s.mux.HandleFunc("/api/snapshot/", func(w http.ResponseWriter, r *http.Request) {
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
		components := make([]protocol.ComponentData, 0, len(snap.States))
		for _, state := range snap.States {
			comp := protocol.ComponentData{
				NodeID:  state.NodeID,
				Type:    state.Type,
				Props:   state.Props,
				State:   state.State,
				Bounds: protocol.RectData{
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

		result := protocol.SnapshotPayload{
			FrameID:   snap.FrameID,
			Timestamp: snap.Timestamp,
			WindowState: protocol.WindowState{
				Width:  snap.Global.WindowSize.Width,
				Height: snap.Global.WindowSize.Height,
			},
			Components: components,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// API: Diff two snapshots
	s.mux.HandleFunc("/api/diff", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")

		if fromStr == "" || toStr == "" {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing from/to parameters"})
			return
		}

		// Parse as frame IDs
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
				"available_frames": getAvailableFrameIDs(s.snapshotManager),
			})
			return
		}

		differ := snapshot.NewDiffer()
		diff := differ.Compare(fromSnap, toSnap)

		// Convert changes to JSON-serializable format
		changes := make([]protocol.ChangeData, 0, len(diff.Changes))
		for _, change := range diff.Changes {
			changes = append(changes, protocol.ChangeData{
				NodeID:   change.NodeID,
				Type:     change.ChangeType.String(),
				Path:     change.Path,
				OldValue: change.OldValue,
				NewValue: change.NewValue,
			})
		}

		result := protocol.DiffPayload{
			From:    devtools.FrameID(fromInt),
			To:      devtools.FrameID(toInt),
			Changes: changes,
		}

		json.NewEncoder(w).Encode(result)
	})

	// API: Export for Chromium
	s.mux.HandleFunc("/api/export", func(w http.ResponseWriter, r *http.Request) {
		data, err := s.bridge.ExportForChromium()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
}

// getAvailableFrameIDs returns a list of available frame IDs.
func getAvailableFrameIDs(sm *snapshot.Manager) []int {
	snapshots := sm.GetAll()
	ids := make([]int, 0, len(snapshots))
	for _, snap := range snapshots {
		ids = append(ids, int(snap.FrameID))
	}
	return ids
}

// GetStats returns server statistics.
func (s *HTTPServer) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"port":      s.port,
		"snapshots": s.snapshotManager.GetStats(),
		"enabled":   s.bridge.IsEnabled(),
	}
}

// DevToolsServer Demo - Using Unified Protocol for Real-time Debugging
//
// This example demonstrates how to use the DevToolsServer for real-time
// debugging and monitoring of TUI applications.
//
// Usage: go run main.go
//
// Then open your browser to:
//
//	http://localhost:8080/
//
// API Endpoints:
//
//	GET /health          - Health check
//	GET /api/frames      - Get all frames
//	GET /api/metrics     - Get performance metrics
//	GET /api/components  - Get all components
//	GET /api/report      - Generate debug report
//	GET /api/snapshots   - Get all snapshots
//	GET /api/diff        - Compare two snapshots
//	GET /api/export      - Export dashboard data
//	POST /api/import     - Import dashboard data
//
// WebSocket:
//
//	ws://localhost:8080/ws - Real-time updates
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/protocol"
)

func main() {
	fmt.Println("=== DevToolsServer Demo (New Protocol) ===")
	fmt.Println()

	// 1. Create DevTools
	dt := devtools.New()
	dt.Enable()

	// 2. Create and start DevToolsServer (using unified protocol package)
	port := 8080
	server := protocol.NewServer(protocol.ServerConfig{
		Port:              port,
		EnableDashboard:   true,
		EnableTuiCommands: true,
	})

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start DevToolsServer: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	fmt.Println("DevToolsServer started!")
	fmt.Printf("  Dashboard:  http://localhost:%d/\n", port)
	fmt.Printf("  Inspector:  http://localhost:%d/debug\n", port)
	fmt.Printf("  WebSocket:  ws://localhost:%d/ws\n", port)
	fmt.Printf("  Health:     http://localhost:%d/health\n", port)
	fmt.Printf("  API:        http://localhost:%d/api/*\n", port)
	fmt.Println()

	// 3. Run simulation
	fmt.Println("Running simulation...")
	go runSimulation(dt, server)

	// 4. Show stats periodically
	go showStats(server)

	// 5. Wait for interrupt signal
	fmt.Println("Press Ctrl+C to stop...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n=== Shutdown ===")
	stats := server.GetMetrics()
	fmt.Printf("Final FPS: %.1f\n", stats.FPS)
}

// runSimulation simulates TUI activity and updates the dashboard.
func runSimulation(dt *devtools.DevTools, server *protocol.Server) {
	frameCount := 0
	componentCount := 0

	for {
		frameCount++

		// Begin frame
		dt.BeginFrame()
		frameStartTime := time.Now()

		// Simulate events
		for i := 0; i < 3; i++ {
			nodeID := devtools.NodeID(fmt.Sprintf("node-%d", i%5))
			dt.RecordEvent("keypress", string(nodeID), "bubble", map[string]interface{}{
				"key":  fmt.Sprintf("key-%d", i),
				"ctrl": i%2 == 0,
			})
		}

		// Simulate component updates
		for i := 0; i < 2; i++ {
			compID := fmt.Sprintf("component-%d", i%10)
			server.UpdateComponent(compID, &protocol.DashboardComponentData{
				ID:   compID,
				Type: getRandomType(),
				Properties: map[string]interface{}{
					"label":    fmt.Sprintf("Item %d", componentCount),
					"active":   componentCount%3 == 0,
					"count":    componentCount,
					"updated":  time.Now().Format(time.RFC3339),
				},
				Styles: map[string]interface{}{
					"color":  getRandomColor(),
					"width":  100 + (componentCount % 50),
					"height": 30 + (componentCount % 20),
				},
			})
			componentCount++
		}

		// End frame
		dt.EndFrame()
		frameDuration := time.Since(frameStartTime)

		// Add frame to server
		server.AddFrame(&protocol.FrameData{
			FrameID:      devtools.FrameID(frameCount),
			Timestamp:    time.Now(),
			Duration:     frameDuration,
			EventCount:   3,
			MutationCount: 2,
			LayoutCount:  1,
			RepaintCount: 1,
		})

		// Update metrics every 10 frames
		if frameCount%10 == 0 {
			// Use simulated frame time for 60 FPS (~16.67ms per frame)
			simulatedFrameTime := 16 * time.Millisecond
			server.UpdateMetrics(&protocol.Metrics{
				FPS:            60.0,
				FrameTime:      simulatedFrameTime,
				LayoutTime:     simulatedFrameTime / 3,
				PaintTime:      simulatedFrameTime / 4,
				MemoryUsage:    uint64(50_000_000 + frameCount*100_000),
				ComponentCount: componentCount,
				FrameCount:     frameCount,
			})
		}

		// Print progress
		if frameCount%20 == 0 {
			fmt.Printf("  Captured %d frames...\n", frameCount)
		}

		// Wait before next frame
		time.Sleep(500 * time.Millisecond)
	}
}

// showStats periodically shows dashboard statistics.
func showStats(server *protocol.Server) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := server.GetMetrics()
		frames := server.GetFrames()

		fmt.Printf("\n--- Stats ---\n")
		fmt.Printf("Frames: %d\n", len(frames))
		fmt.Printf("Components: %d\n", metrics.ComponentCount)
		fmt.Printf("FPS: %.1f\n", metrics.FPS)

		// Show health endpoint response
		resp, err := http.Get("http://localhost:8080/health")
		if err == nil {
			var health map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&health)
			resp.Body.Close()
			fmt.Printf("Health: %v\n", health["status"])
		}
		fmt.Println()
	}
}

// getRandomType returns a random component type.
func getRandomType() string {
	types := []string{"Button", "Label", "Container", "List", "Input"}
	return types[(int(time.Now().Unix()) / 10) % len(types)]
}

// getRandomColor returns a random color.
func getRandomColor() string {
	colors := []string{"#4ec9b0", "#dcdcaa", "#ce9178", "#569cd6", "#c586c0"}
	return colors[(int(time.Now().Unix()) / 5) % len(colors)]
}

// Remote Debugging Demo - Using ChromiumBridge
// This is a standalone demo - run separately from main.go
// Usage: go run remote_demo.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/observation"
	v1 "github.com/wwsheng009/mint/devtools/observation/v1"
	"github.com/wwsheng009/mint/devtools/remote"
	"github.com/wwsheng009/mint/devtools/snapshot"
)

func main() {
	fmt.Println("=== DevTools Remote Debugging Demo ===")
	fmt.Println()

	// 1. Create DevTools and Snapshot Manager
	dt := devtools.New()
	dt.Enable()

	snapshotMgr := snapshot.NewManager(100)

	// 2. Create and start DevTools server (HTTP + WebSocket)
	server := remote.NewDevToolsServer(9222, dt, snapshotMgr)
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	fmt.Println("Remote debugging server started!")
	fmt.Println("  Inspector: http://localhost:9222/debug")
	fmt.Println("  WebSocket: ws://localhost:9222/ws")
	fmt.Println("  API:       http://localhost:9222/api/snapshots")
	fmt.Println("  Diff:      http://localhost:9222/api/diff?from=0&to=3")
	fmt.Println("  Health:    http://localhost:9222/health")
	fmt.Println()

	// 3. Simulate some activity and capture snapshots
	runSimulation(dt, snapshotMgr)

	// 4. Show results
	fmt.Println("\n=== Simulation Complete ===")
	fmt.Println("Check http://localhost:9222/api/snapshots for snapshot data")
	fmt.Println("Open http://localhost:9222/debug for inspector UI")
	fmt.Println()

	// Show stats
	stats := snapshotMgr.GetStats()
	fmt.Printf("Snapshots: %d/%d\n", stats.TotalSnapshots, stats.MaxSnapshots)

	// Keep running
	fmt.Println("Press Ctrl+C to stop...")
	select {}
}

func runSimulation(dt *devtools.DevTools, sm *snapshot.Manager) {
	// Create observation layer for better data
	cfg := observation.DefaultConfig()
	cfg.InitialLevel = v1.LevelAdvanced
	obs := observation.NewLayer(cfg)
	obs.Enable(v1.LevelAdvanced)

	fmt.Println("Simulating activity and capturing snapshots...")

	for i := 0; i < 10; i++ {
		// Record frame
		dt.BeginFrame()

		// Simulate event
		nodeID := devtools.NodeID(fmt.Sprintf("node-%d", i%3))
		dt.RecordEvent("keypress", string(nodeID), "bubble", nil)

		// Record mutation
		obs.RecordMutation(nodeID, "value", i)

		dt.EndFrame()

		// Capture snapshot every few frames
		if i%3 == 0 {
			snapID := snapshot.SnapshotID(fmt.Sprintf("snap-%d", i))
			builder := snapshot.NewBuilder(snapID, devtools.FrameID(i))
			builder.SetWindowSize(80, 24)
			builder.SetLabel("test", "demo")

			// Add component state
			builder.AddComponent(&snapshot.ComponentState{
				NodeID:  nodeID,
				Type:    "Button",
				Visible: true,
				Focused: i == 0,
				Bounds:  snapshot.Rect{X: 0, Y: i, Width: 20, Height: 1},
				State:   map[string]interface{}{"clicked": i > 5},
			})

			snap, err := sm.Capture(devtools.FrameID(i), builder)
			if err != nil {
				log.Printf("Capture error: %v", err)
			} else {
				fmt.Printf("  Captured snapshot %s at frame %d\n", snap.ID, snap.FrameID)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Show diff between first and last
	all := sm.GetAll()
	if len(all) >= 2 {
		diff := snapshot.CompareSnapshots(all[0], all[len(all)-1])
		fmt.Printf("\nDiff: %s\n", diff.FormatSummary())
	}
}

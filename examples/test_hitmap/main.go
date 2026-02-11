package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layer"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Enable debug output
	os.Setenv("TUI_DEBUG_HITMAP", "true")
	os.Setenv("TUI_LAYER_DEBUG", "true")

	// Create a simple modal with buttons
	modalContent := ui.VStack(
		ui.Text("Modal Title"),
		ui.HStack(
			ui.Button("Cancel").Key("button-cancel"),
			ui.Button("OK").Key("button-ok"),
		),
	)

	modal := ui.Modal(modalContent)

	// Create layer manager
	manager := layer.NewManager()

	// Create layout engine
	engine := compute.NewEngine()

	// Define constraints (simulating a 100x30 screen)
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 30,
	}

	// Collect and layout
	err := manager.CollectAndLayout(modal, constraints, engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CollectAndLayout failed: %v\n", err)
		os.Exit(1)
	}

	// Get modal layout
	modalLayout, hasModal := manager.GetLayout(runtime.LayerModal)
	if !hasModal {
		fmt.Fprintf(os.Stderr, "❌ No modal layout found!\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\n=== MODAL LAYOUT ===\n")
	fmt.Fprintf(os.Stderr, "Modal Root: pos=(%d,%d) size=%dx%d\n",
		modalLayout.Root.Box.X, modalLayout.Root.Box.Y,
		modalLayout.Root.Box.Width, modalLayout.Root.Box.Height)

	// Get HitMap
	hitMap := modalLayout.HitMap
	if hitMap == nil {
		fmt.Fprintf(os.Stderr, "❌ Modal HitMap is nil!\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\n=== HITMAP ENTRIES ===\n")
	fmt.Fprintf(os.Stderr, "Total entries: %d\n", hitMap.Size())

	// Output all entries
	entries := hitMap.AllEntries()
	for i, entry := range entries {
		fmt.Fprintf(os.Stderr, "Entry %d: ID=%s, Bounds=(%d,%d,%dx%d)\n",
			i, entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height)
	}

	// Get merged HitMap
	mergedHitMap := manager.GetMergedHitMap()
	if mergedHitMap == nil {
		fmt.Fprintf(os.Stderr, "❌ Merged HitMap is nil!\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\n=== MERGED HITMAP ===\n")
	fmt.Fprintf(os.Stderr, "Total merged entries: %d\n", mergedHitMap.Size())

	// Output all merged entries
	mergedEntries := mergedHitMap.AllEntries()
	for i, entry := range mergedEntries {
		fmt.Fprintf(os.Stderr, "Merged Entry %d: ID=%s, Bounds=(%d,%d,%dx%d)\n",
			i, entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height)
	}

	// Verify button positions
	fmt.Fprintf(os.Stderr, "\n=== VERIFICATION ===\n")

	// Find button entries
	var cancelBtn, okBtn *runtimeevent.HitMapEntry
	for i := range entries {
		if entries[i].NodeID == "button-cancel" {
			cancelBtn = &entries[i]
		}
		if entries[i].NodeID == "button-ok" {
			okBtn = &entries[i]
		}
	}

	if cancelBtn == nil {
		fmt.Fprintf(os.Stderr, "❌ Cancel button not found in HitMap!\n")
	} else {
		fmt.Fprintf(os.Stderr, "Cancel Button: Bounds=(%d,%d,%dx%d)\n",
			cancelBtn.Bounds.X, cancelBtn.Bounds.Y,
			cancelBtn.Bounds.Width, cancelBtn.Bounds.Height)

		// Check if button is centered (not at 0,0)
		if cancelBtn.Bounds.X == 0 && cancelBtn.Bounds.Y == 0 {
			fmt.Fprintf(os.Stderr, "❌ Cancel button is at (0,0) - NOT centered! Expected center position.\n")
		} else {
			fmt.Fprintf(os.Stderr, "✅ Cancel button is NOT at (0,0) - appears to be centered\n")
		}
	}

	if okBtn == nil {
		fmt.Fprintf(os.Stderr, "❌ OK button not found in HitMap!\n")
	} else {
		fmt.Fprintf(os.Stderr, "OK Button: Bounds=(%d,%d,%dx%d)\n",
			okBtn.Bounds.X, okBtn.Bounds.Y,
			okBtn.Bounds.Width, okBtn.Bounds.Height)

		// Check if button is centered (not at 0,0)
		if okBtn.Bounds.X == 0 && okBtn.Bounds.Y == 0 {
			fmt.Fprintf(os.Stderr, "❌ OK button is at (0,0) - NOT centered! Expected center position.\n")
		} else {
			fmt.Fprintf(os.Stderr, "✅ OK button is NOT at (0,0) - appears to be centered\n")
		}
	}

	fmt.Fprintf(os.Stderr, "\n=== TEST COMPLETE ===\n")
}

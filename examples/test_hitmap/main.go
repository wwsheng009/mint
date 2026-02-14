package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/layer"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Enable debug output
	os.Setenv("TUI_DEBUG_HITMAP", "true")
	os.Setenv("TUI_DEBUG_LAYER", "true")

	// Create a simple modal with elements
	modalContent := ui.VStack(
		ui.Text("Modal Title"),
		ui.HStack(
			ui.Element("button").Prop("label", "Cancel").Prop("key", "button-cancel").Build(),
			ui.Element("button").Prop("label", "OK").Prop("key", "button-ok").Build(),
		),
	)

	modal := ui.Modal(modalContent).Build()

	// Create layer manager
	manager := layer.NewManager()

	// Create layout engine
	engine := compute.NewEngine()

	// Create fiber from vnode
	fiber := reconciler.CreateFiber(modal)

	// Define constraints (simulating a 100x30 screen)
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 30,
	}

	// Phase 4.5: Use CollectAndLayout with modal centering
	// Note: This is still using the old API temporarily for modal centering.
	// In Phase 5-7, modal centering will move to Layout Engine and we can use:
	//  1. engine.Layout(vnode, fiber, constraints)
	//  2. renderPlanes := layer.BuildFromFiber(fiber)
	//  3. event.BuildHitMapFromFiber(fiber)
	err := manager.CollectAndLayout(modal, fiber, constraints, engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CollectAndLayout failed: %v\n", err)
		os.Exit(1)
	}

	// Get modal layout
	modalLayout, hasModal := manager.GetLayout(rtui.LayerModal)
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
		fmt.Fprintf(os.Stderr, "Entry %d: ID=%s, NodeID=%d, Bounds=(%d,%d,%dx%d)\n",
			i, entry.Node.ID(), entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
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
		fmt.Fprintf(os.Stderr, "Merged Entry %d: ID=%s, NodeID=%d, Bounds=(%d,%d,%dx%d)\n",
			i, entry.Node.ID(), entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height)
	}

	// Verify button positions
	fmt.Fprintf(os.Stderr, "\n=== VERIFICATION ===\n")

	// Find button entries
	var cancelBtn, okBtn *runtimeevent.HitMapEntry
	for i := range entries {
		if entries[i].Node.ID() == "button-cancel" {
			cancelBtn = &entries[i]
		}
		if entries[i].Node.ID() == "button-ok" {
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

	// Phase 4.5: Also demonstrate new API usage
	// Build RenderPlanes directly from Fiber (without using CollectAndLayout)
	fmt.Fprintf(os.Stderr, "\n=== NEW API DEMO (Phase 3+) ===\n")
	renderPlanes := manager.BuildRenderPlanes(fiber)
	fmt.Fprintf(os.Stderr, "Built RenderPlanes with %d boxes\n", renderPlanes.CountBoxes())

	// Build HitMap directly from Fiber (Phase 6 style)
	// Using non-reflection implementation from ui package
	newHitMap := rtui.BuildHitMapFromFiber(fiber)
	fmt.Fprintf(os.Stderr, "Built HitMap from Fiber with %d entries\n", newHitMap.Size())

	// Note: The new API doesn't include modal centering yet (that's Phase 5-7)
	// So the positions will be different (at 0,0 instead of centered)
	newEntries := newHitMap.AllEntries()
	for i, entry := range newEntries {
		fmt.Fprintf(os.Stderr, "New API Entry %d: ID=%s, NodeID=%d, Bounds=(%d,%d,%dx%d)\n",
			i, entry.Node.ID(), entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height)
	}

	fmt.Fprintf(os.Stderr, "\n=== TEST COMPLETE ===\n")
	fmt.Fprintf(os.Stderr, "\nNOTE: Phase 4.5 - Using CollectAndLayout for modal centering transition.\n")
	fmt.Fprintf(os.Stderr, "In Phase 5-7, modal centering will move to Layout Engine.\n")
	fmt.Fprintf(os.Stderr, "Then we can fully migrate to the new API:\n")
	fmt.Fprintf(os.Stderr, "  1. engine.Layout(vnode, fiber, constraints)\n")
	fmt.Fprintf(os.Stderr, "  2. renderPlanes := layer.BuildFromFiber(fiber)\n")
	fmt.Fprintf(os.Stderr, "  3. hitMap := ui.BuildHitMapFromFiber(fiber)  // Non-reflection version\n")
}

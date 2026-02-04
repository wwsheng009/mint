// DevTools Panel Demo - Demonstrates how to use the DevTools control panel
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/client"
)

func main() {
	fmt.Println("=== DevTools Control Panel Demo ===")
	fmt.Println()

	// 1. Create DevTools instance
	dt := devtools.New()
	dt.Enable()

	// 2. Create the debug panel (use TuiDebugPanel for command handler compatibility)
	panel := client.NewTuiDebugPanel(dt)
	panel.Enable()

	// 3. Configure panel size
	panel.SetSize(80, 24)

	// 4. Simulate some activity
	fmt.Println("Simulating TUI activity...")
	for i := 0; i < 10; i++ {
		dt.BeginFrame()

		// Simulate some events
		nodeID := fmt.Sprintf("node-%d", i%3)
		dt.RecordEvent("keypress", nodeID, "bubble", map[string]interface{}{
			"key": fmt.Sprintf("key-%d", i),
		})

		// Update panel to current frame
		panel.SetSelectedFrame(devtools.FrameID(i))

		time.Sleep(100 * time.Millisecond)
		dt.EndFrame()
	}

	// 5. Render the panel
	fmt.Println("\n--- Panel Render Output ---")
	renderOutput := panel.Render()
	// Show first 10 lines of output
	lines := splitLines(renderOutput)
	maxLines := 10
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	for i := 0; i < maxLines; i++ {
		fmt.Println(lines[i])
	}
	if len(lines) > maxLines {
		fmt.Printf("... (%d more lines)\n", len(lines)-maxLines)
	}

	// 6. Demo keyboard input handling
	fmt.Println("\n--- Keyboard Input Demo ---")
	demoKeyboardInput(panel)

	// 7. Demo inspection
	fmt.Println("\n--- Component Inspection Demo ---")
	demoInspection(panel)

	// 8. Demo command handler
	fmt.Println("\n--- Command Handler Demo ---")
	demoCommands(panel)

	// 9. Demo debug overlay
	fmt.Println("\n--- Debug Overlay Demo ---")
	demoOverlay(dt)

	// Cleanup
	panel.Disable()
	dt.Disable()
	_ = dt.Shutdown()
	fmt.Println("\nDemo completed!")
}

// demoKeyboardInput demonstrates keyboard input handling
func demoKeyboardInput(panel *client.TuiDebugPanel) {
	keys := []rune{'t', 'c', 's', 'r', 'q'}

	fmt.Println("  Keyboard shortcuts:")
	fmt.Println("    [t] - Toggle Timeline view")
	fmt.Println("    [c] - Toggle Causal Graph view")
	fmt.Println("    [s] - Toggle Snapshots view")
	fmt.Println("    [r] - Toggle Replay view")
	fmt.Println("    [q] - Quit panel")

	for _, key := range keys {
		fmt.Printf("\n  Pressing '%c' -> ", key)
		keepRunning := panel.HandleInput(key)

		state := panel.GetState()
		switch key {
		case 't':
			fmt.Printf("Timeline: %v\n", state.ShowTimeline)
		case 'c':
			fmt.Printf("Causal: %v\n", state.ShowCausal)
		case 's':
			fmt.Printf("Snapshots: %v\n", state.ShowSnapshots)
		case 'r':
			fmt.Printf("Replay: %v\n", state.ShowReplay)
		case 'q':
			fmt.Println("Quit panel")
		}

		if !keepRunning {
			break
		}
	}
}

// demoInspection demonstrates component inspection
func demoInspection(panel *client.TuiDebugPanel) {
	nodeID := "button_submit"

	result := panel.Inspect(nodeID)
	fmt.Printf("  Inspecting: %s\n", nodeID)
	fmt.Printf("    Node ID: %s\n", result.NodeID)
	fmt.Printf("    Type: %s\n", result.Type)
	fmt.Printf("    Position: %s\n", result.Position)
	fmt.Printf("    Size: %s\n", result.Size)
	fmt.Printf("    Children: %d\n", len(result.Children))
}

// demoCommands demonstrates command usage
func demoCommands(panel *client.TuiDebugPanel) {
	cmdHandler := client.NewCommandHandler(panel)

	commands := []string{
		"help",
		"inspect button_submit",
		"frame 42",
		"stats",
	}

	for _, cmd := range commands {
		fmt.Printf("\n  > %s\n", cmd)
		result := cmdHandler.Execute(cmd)
		if result != "" {
			// Indent the result output
			for _, line := range splitLines(result) {
				fmt.Printf("    %s\n", line)
			}
		}
	}
}

// demoOverlay demonstrates the debug overlay
func demoOverlay(dt *devtools.DevTools) {
	overlay := dt.GetOverlay()
	if overlay == nil {
		fmt.Println("  Overlay not enabled (create DevTools with EnableOverlay option)")
		return
	}

	fmt.Println("  Debug Overlay:")
	fmt.Println("    Highlighting components...")

	// Highlight some regions
	dt.Highlight("button_1", 10, 5, 20, 3)
	dt.Highlight("input_field", 10, 10, 30, 1)
	dt.Highlight("status_bar", 0, 20, 80, 1)

	fmt.Println("    Highlighted: button_1, input_field, status_bar")

	// Check highlights
	components := []string{"button_1", "input_field", "status_bar"}
	for _, comp := range components {
		shown := dt.GetOverlay().IsShown(comp)
		fmt.Printf("      %s: %v\n", comp, shown)
	}

	// Clear overlay
	dt.ClearOverlay()
	fmt.Println("    Cleared all highlights")
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	lines := []string{}
	current := ""

	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

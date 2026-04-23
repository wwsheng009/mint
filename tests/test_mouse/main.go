package main

import (
	"fmt"
	"os"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
)

func main() {
	fmt.Println("=== Input Test Program ===")
	fmt.Println("Press keys, move mouse, click to see events")
	fmt.Println("Press ESC or Ctrl+C to exit")
	fmt.Println()

	reader, err := platform.NewInputReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create input reader: %v\n", err)
		return
	}

	events := make(chan platform.RawInput, 100)
	if err := reader.Start(events); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start input reader: %v\n", err)
		return
	}
	defer reader.Stop()

	fmt.Println("Input reader started. Waiting for events...")
	fmt.Println("Note: This test requires a Windows Console (cmd, PowerShell, Windows Terminal)")
	fmt.Println("      Mouse events may not work in Git Bash or MinTTY.")
	fmt.Println()

	count := 0
	mouseCount := 0
	keyCount := 0
	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Println("\n=== Timeout after 30 seconds ===")
			fmt.Printf("Total events: %d (Keys: %d, Mouse: %d)\n", count, keyCount, mouseCount)
			if count == 0 {
				fmt.Println("\nNo events received! Possible reasons:")
				fmt.Println("1. Not running in a Windows Console")
				fmt.Println("2. Running through a pipe/redirect")
				fmt.Println("3. Terminal doesn't support console API")
			}
			return

		case ev, ok := <-events:
			if !ok {
				fmt.Println("Event channel closed")
				return
			}
			count++

			switch ev.Type {
			case platform.InputMouse:
				mouseCount++
				actionStr := "Unknown"
				switch ev.MouseAction {
				case platform.MousePress:
					actionStr = "Press"
				case platform.MouseRelease:
					actionStr = "Release"
				case platform.MouseMotion:
					actionStr = "Motion"
				case platform.MouseWheelUp:
					actionStr = "WheelUp"
				case platform.MouseWheelDown:
					actionStr = "WheelDown"
				}
				buttonStr := "None"
				switch ev.MouseButton {
				case platform.MouseLeft:
					buttonStr = "Left"
				case platform.MouseRight:
					buttonStr = "Right"
				case platform.MouseMiddle:
					buttonStr = "Middle"
				}
				fmt.Printf("[MOUSE #%d] Action: %-8s Button: %-6s Position: (%3d, %3d)\n",
					mouseCount, actionStr, buttonStr, ev.MouseX, ev.MouseY)

			case platform.InputKeyPress:
				keyCount++
				keyName := fmt.Sprintf("'%c'", ev.Key)
				if ev.Special != platform.KeyUnknown {
					keyName = ev.Special.String()
				}
				fmt.Printf("[KEY #%d] Key: %-15s Modifiers: %03b\n",
					keyCount, keyName, ev.Modifiers)

				if ev.Special == platform.KeyEscape || (ev.Key == 'c' && ev.Modifiers&platform.ModCtrl != 0) {
					fmt.Println("\n=== Exiting ===")
					fmt.Printf("Total events: %d (Keys: %d, Mouse: %d)\n", count, keyCount, mouseCount)
					return
				}

			case platform.InputResize:
				fmt.Printf("[RESIZE] New size: %dx%d\n", ev.Width, ev.Height)
			}
		}
	}
}

// +build ignore

// This is a test program to check what keys Windows console is sending
// Run with: go run key_debug_main.go
// Then press Ctrl+D, Ctrl+K, Alt+K, Shift+K etc to see what's received

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/platform"
)

func main() {
	fmt.Println("Key Debug Test")
	fmt.Println("==============")
	fmt.Println("Press keys to see what Windows console sends...")
	fmt.Println("Try: Ctrl+D, Ctrl+K, Alt+K, Shift+K, F12")
	fmt.Println("Press ESC to exit")
	fmt.Println()

	// Enable debug mode
	os.Setenv("TUI_DEBUG_INPUT", "true")

	reader := platform.NewInputReader()
	if err := reader.Start(make(chan platform.RawInput, 100)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start input reader: %v\n", err)
		return
	}
	defer reader.Stop()

	inputChan := make(chan platform.RawInput, 100)
	go func() {
		for {
			input, err := reader.Read()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
				return
			}
			inputChan <- input
		}
	}()

	for input := range inputChan {
		if input.Type == platform.InputKeyPress {
			keyName := ""
			if input.Special != platform.KeyUnknown {
				keyName = fmt.Sprintf("Special=%d", input.Special)
			} else if input.Key > 0 {
				keyName = fmt.Sprintf("Key='%c'", input.Key)
			}

			modStr := ""
			if input.Modifiers&platform.ModAlt != 0 {
				modStr += "Alt+"
			}
			if input.Modifiers&platform.ModCtrl != 0 {
				modStr += "Ctrl+"
			}
			if input.Modifiers&platform.ModShift != 0 {
				modStr += "Shift+"
			}
			if modStr == "" {
				modStr = "none"
			}

			fmt.Printf("Received: %s Modifiers=%s\n", keyName, modStr)

			// Exit on ESC
			if input.Special == platform.KeyEscape {
				fmt.Println("\nExiting...")
				return
			}
		}
	}
}

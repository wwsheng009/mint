// Package main demonstrates focus switching between multiple components
// This demo shows Tab-based keyboard navigation across focusable components
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	buttonComp "github.com/wwsheng009/mint/ui/components/button"
	checkboxComp "github.com/wwsheng009/mint/ui/components/checkbox"
	inputComp "github.com/wwsheng009/mint/ui/components/input"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// SimpleApp demonstrates focusable components with static rendering
func SimpleApp() ui.VNode {
	// Button 1
	btn1 := buttonComp.New("Button 1 - First")
	btn1.SetIntent(intent.Click("btn1"))
	btn1.SetKey("btn1")

	// Input 1
	input1 := inputComp.New()
	input1.SetPlaceholder("Enter name...")
	input1.SetWidth(25)
	input1.SetChangeIntent(intent.SetState("input1-value", "entered"))
	input1.SetKey("input1")

	// Checkbox 1
	chk1 := checkboxComp.New("Option A")
	chk1.SetIntent(intent.Toggle("chk1-checked"))
	chk1.SetKey("chk1")

	// Button 2
	btn2 := buttonComp.New("Button 2 - Middle")
	btn2.SetIntent(intent.Click("btn2"))
	btn2.SetKey("btn2")

	// Input 2
	input2 := inputComp.New()
	input2.SetPlaceholder("Enter email...")
	input2.SetWidth(25)
	input2.SetChangeIntent(intent.SetState("input2-value", "entered"))
	input2.SetKey("input2")

	// Checkbox 2
	chk2 := checkboxComp.New("Option B")
	chk2.SetIntent(intent.Toggle("chk2-checked"))
	chk2.SetKey("chk2")

	// Button 3
	btn3 := buttonComp.New("Button 3 - Last")
	btn3.SetIntent(intent.Click("btn3"))
	btn3.SetKey("btn3")

	// Disabled Button
	disabledBtn := buttonComp.New("Disabled Button")
	disabledBtn.SetDisabled(true)
	disabledBtn.SetIntent(intent.Click("btn4"))
	disabledBtn.SetKey("btn4")

	// Disabled Input
	disabledInput := inputComp.New()
	disabledInput.SetValue("Disabled Input")
	disabledInput.SetDisabled(true)
	disabledInput.SetChangeIntent(intent.SetState("input3-value", ""))
	disabledInput.SetKey("input3")

	// Disabled Checkbox
	disabledChk := checkboxComp.New("Disabled Checkbox")
	disabledChk.SetDisabled(true)
	disabledChk.SetIntent(intent.Toggle("chk3-checked"))
	disabledChk.SetKey("chk3")

	return ui.NewVStack().
		SetWidth(50).
		SetGap(1).
		SetChildrenList([]ui.VNode{
			// Title
			newtext.New("=== Focus Switching Demo ==="),
			newtext.New("Press TAB to navigate focus"),
			newtext.New(""),

			// Section 1: Normal focusable components
			newtext.New("1. Focusable Components:"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]ui.VNode{
					// Button 1 - with Click intent
					btn1,

					// Input 1 - with Change intent
					input1,

					// Checkbox 1 - with Toggle intent
					chk1,

					// Button 2 - with Click intent
					btn2,

					// Input 2 - with Change intent
					input2,

					// Checkbox 2 - with Toggle intent
					chk2,

					// Button 3 - with Click intent
					btn3,
				}),

			newtext.New(""),

			// Section 2: Disabled components (not focusable)
			newtext.New("2. Disabled Components (skipped by focus):"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]ui.VNode{
					disabledBtn,
					disabledInput,
					disabledChk,
				}),

			newtext.New(""),

			// Focus instructions
			newtext.New("Navigation:"),
			newtext.New("  TAB       - Move to next focusable element"),
			newtext.New("  SHIFT+TAB - Move to previous focusable element"),
			newtext.New("  ENTER     - Activate focused button/checkbox"),
			newtext.New("  SPACE     - Toggle focused checkbox"),
		})
}

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Focus Switching Demo - Interactive Keyboard Navigation  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("This demo showcases focus management with multiple component types:")
	fmt.Println("  - Buttons (3 focusable)")
	fmt.Println("  - Input fields (2 focusable)")
	fmt.Println("  - Checkboxes (2 focusable)")
	fmt.Println("  - Disabled components (3, skipped during navigation)")
	fmt.Println("")
	fmt.Println("Intent Usage Examples:")
	fmt.Println("  Button:   SetIntent(intent.Click(\"btn1\"))")
	fmt.Println("  Input:    SetChangeIntent(intent.SetState(\"key\", \"value\"))")
	fmt.Println("  Checkbox: SetIntent(intent.Toggle(\"key\"))")
	fmt.Println("")
	fmt.Println("Using ui.Run() to start the application...")
	fmt.Println("")
	fmt.Println("Press TAB to move focus between components.")
	fmt.Println("Press ESC or CTRL+C to exit.")
	fmt.Println("")

	err := ui.Run(SimpleApp,
		ui.WithWidth(60),
		ui.WithHeight(35),
		ui.WithTitle("Focus Switching Demo"),
	)
	if err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}

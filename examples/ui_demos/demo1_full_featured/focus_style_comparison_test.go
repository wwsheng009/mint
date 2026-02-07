package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/components/button"
	"github.com/wwsheng009/mint/runtime/paint"
)

// TestAllFocusStyles compares all focus style options
func TestAllFocusStyles(t *testing.T) {
	fmt.Println("\n=== Comparing All Focus Styles ===")

	styles := []struct {
		name        string
		focusStyle  app.ButtonFocusStyle
		description string
	}{
		{"Bracket", app.FocusStyleBracket, "Bright yellow FG, '>' indicator"},
		{"Underline", app.FocusStyleUnderline, "Bright yellow FG, underline"},
		{"Bold", app.FocusStyleBold, "Bold only, preserves colors"},
		{"Reverse", app.FocusStyleReverse, "Inverted FG/BG colors"},
	}

	for _, styleInfo := range styles {
		fmt.Printf("--- FocusStyle: %s (%s) ---\n", styleInfo.name, styleInfo.description)

		btn := app.ButtonBuilder("[Test Button]").
			Variant(app.ButtonVariantPrimary).
			FocusStyle(styleInfo.focusStyle).
			Build()

		btnVNode, ok := btn.(*button.ButtonVNode)
		if !ok {
			t.Fatal("Expected ButtonVNode")
		}

		// Without focus
		btnVNode.SetFocus(false)
		cmdsNoFocus := btnVNode.Paint(0, 0)
		if len(cmdsNoFocus) > 0 {
			cmd := cmdsNoFocus[0]
			ssm := paint.NewStyleStateMachine()
			ansiNoFocus := ssm.Update(cmd.Style)
			fmt.Printf("  No Focus: Text=%-20q ANSI=%-40s FG=%s BG=%s\n",
				cmd.Text, ansiNoFocus, cmd.Style.FG, cmd.Style.BG)
		}

		// With focus
		btnVNode.SetFocus(true)
		cmdsWithFocus := btnVNode.Paint(0, 0)
		if len(cmdsWithFocus) > 0 {
			cmd := cmdsWithFocus[0]
			ssm2 := paint.NewStyleStateMachine()
			ansiFocus := ssm2.Update(cmd.Style)
			fmt.Printf("  With Focus: Text=%-20q ANSI=%-40s FG=%s BG=%s\n",
				cmd.Text, ansiFocus, cmd.Style.FG, cmd.Style.BG)
		}

		fmt.Println()
	}
}

// TestFocusVisualDifference calculates how different the styles are
func TestFocusVisualDifference(t *testing.T) {
	fmt.Println("\n=== Visual Difference Analysis ===")

	styles := []struct {
		name       string
		focusStyle app.ButtonFocusStyle
	}{
		{"Bracket", app.FocusStyleBracket},
		{"Underline", app.FocusStyleUnderline},
		{"Bold", app.FocusStyleBold},
		{"Reverse", app.FocusStyleReverse},
	}

	for _, styleInfo := range styles {
		btn := app.ButtonBuilder("[Test]").
			Variant(app.ButtonVariantPrimary).
			FocusStyle(styleInfo.focusStyle).
			Build()

		btnVNode, _ := btn.(*button.ButtonVNode)

		btnVNode.SetFocus(false)
		cmdsNoFocus := btnVNode.Paint(0, 0)

		btnVNode.SetFocus(true)
		cmdsWithFocus := btnVNode.Paint(0, 0)

		if len(cmdsNoFocus) > 0 && len(cmdsWithFocus) > 0 {
			noFocusStyle := cmdsNoFocus[0].Style
			focusStyle := cmdsWithFocus[0].Style

			// Check what changed
			fgChanged := noFocusStyle.FG != focusStyle.FG
			bgChanged := noFocusStyle.BG != focusStyle.BG
			boldChanged := noFocusStyle.IsBold() != focusStyle.IsBold()
			underlineChanged := noFocusStyle.IsUnderline() != focusStyle.IsUnderline()

			fmt.Printf("%s:\n", styleInfo.name)
			if fgChanged {
				fmt.Printf("  ✓ FG changed: %s -> %s\n", noFocusStyle.FG, focusStyle.FG)
			}
			if bgChanged {
				fmt.Printf("  ✓ BG changed: %s -> %s\n", noFocusStyle.BG, focusStyle.BG)
			}
			if boldChanged {
				fmt.Printf("  ✓ Bold changed: %v -> %v\n", noFocusStyle.IsBold(), focusStyle.IsBold())
			}
			if underlineChanged {
				fmt.Printf("  ✓ Underline changed: %v -> %v\n", noFocusStyle.IsUnderline(), focusStyle.IsUnderline())
			}

			if !fgChanged && !bgChanged && !boldChanged && !underlineChanged {
				fmt.Printf("  ✗ No visual change!\n")
			}
			fmt.Println()
		}
	}
}

// TestPrimaryVariantFocus focuses on Primary variant specifically
func TestPrimaryVariantFocus(t *testing.T) {
	fmt.Println("\n=== Primary Variant Focus Analysis ===")

	// Test each focus style with Primary variant
	styles := []struct {
		name       string
		focusStyle app.ButtonFocusStyle
	}{
		{"Bracket", app.FocusStyleBracket},
		{"Underline", app.FocusStyleUnderline},
		{"Bold", app.FocusStyleBold},
		{"Reverse", app.FocusStyleReverse},
	}

	for _, styleInfo := range styles {
		btn := app.ButtonBuilder("[Primary]").
			Variant(app.ButtonVariantPrimary).
			FocusStyle(styleInfo.focusStyle).
			Build()

		btnVNode, _ := btn.(*button.ButtonVNode)

		// Primary variant: BG=PRIMARY (#88c0d0), FG=BG (#2e3440)
		btnVNode.SetFocus(false)
		cmdsNoFocus := btnVNode.Paint(0, 0)

		btnVNode.SetFocus(true)
		cmdsWithFocus := btnVNode.Paint(0, 0)

		if len(cmdsNoFocus) > 0 && len(cmdsWithFocus) > 0 {
			noFocus := cmdsNoFocus[0]
			withFocus := cmdsWithFocus[0]

			fmt.Printf("%s style:\n", styleInfo.name)
			fmt.Printf("  Without focus: FG=%s, BG=%s, Text=%q\n",
				noFocus.Style.FG, noFocus.Style.BG, noFocus.Text)
			fmt.Printf("  With focus:    FG=%s, BG=%s, Text=%q\n",
				withFocus.Style.FG, withFocus.Style.BG, withFocus.Text)

			// Calculate visual difference score
			diffScore := 0
			if noFocus.Style.FG != withFocus.Style.FG {
				diffScore += 3 // FG change is most visible
				fmt.Printf("  ✓ FG change (+3)\n")
			}
			if noFocus.Style.BG != withFocus.Style.BG {
				diffScore += 2 // BG change is visible
				fmt.Printf("  ✓ BG change (+2)\n")
			}
			if noFocus.Style.IsBold() != withFocus.Style.IsBold() {
				diffScore += 1 // Bold change is subtle
				fmt.Printf("  ✓ Bold change (+1)\n")
			}
			if noFocus.Style.IsUnderline() != withFocus.Style.IsUnderline() {
				diffScore += 1 // Underline change is subtle
				fmt.Printf("  ✓ Underline change (+1)\n")
			}

			fmt.Printf("  Visual difference score: %d/7\n", diffScore)
			if diffScore >= 3 {
				fmt.Printf("  ✓ Good visibility\n")
			} else if diffScore >= 1 {
				fmt.Printf("  ⚠ Moderate visibility\n")
			} else {
				fmt.Printf("  ✗ Poor visibility\n")
			}
			fmt.Println()
		}
	}
}

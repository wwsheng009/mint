// Package main demonstrates the constraint tracer integration with Panel.
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	panel "github.com/wwsheng009/mint/ui/components/panel"
)

func main() {
	// ========================================
	// Step 1: Enable constraint tracer
	// ========================================
	layout.EnableTracer()
	fmt.Println("=== Constraint Tracer Enabled ===")
	fmt.Println()

	// ========================================
	// Step 2: Create a simple Panel with explicit dimensions
	// ========================================
	fmt.Println("=== Test 1: Panel with explicit dimensions (50x10) ===")
	fmt.Println()

	panelNode := panel.New().
		SetTitle("Test Panel").
		SetContent(newtext.New("This is test content inside the panel.\nThe panel has explicit dimensions.")).
		SetWidth(50).
		SetHeight(10).
		Rounded()

	// Get the instance directly from VNode
	instance := panelNode.CreateInstance()
	if panelInst, ok := instance.(*panel.PanelInstance); ok {
		// Set path for trace
		panelInst.SetPath("/root/test1")

		// Measure with constraints
		constraints := layout.UnboundedConstraints()
		size := panelInst.Measure(constraints)
		fmt.Printf("Input constraints: unbounded\n")
		fmt.Printf("Measured size: %dx%d\n", size.Width, size.Height)
		fmt.Println()
	}

	// ========================================
	// Step 3: Panel with tight constraints
	// ========================================
	fmt.Println("=== Test 2: Panel with tight constraints (40x8) ===")
	fmt.Println()

	panelNode2 := panel.New().
		SetTitle("Constrained Panel").
		SetContent(newtext.New("Content constrained by parent.")).
		SetFlex(1) // Will expand

	instance2 := panelNode2.CreateInstance()
	if panelInst2, ok := instance2.(*panel.PanelInstance); ok {
		// Set path for trace
		panelInst2.SetPath("/root/test2")

		// Measure with tight constraints
		tightConstraints := layout.NewConstraints(0, 40, 0, 8)
		size2 := panelInst2.Measure(tightConstraints)
		fmt.Printf("Input constraints: 0-%d x 0-%d (tight)\n", tightConstraints.MaxWidth, tightConstraints.MaxHeight)
		fmt.Printf("Measured size: %dx%d\n", size2.Width, size2.Height)
		fmt.Println()
	}

	// ========================================
	// Step 4: Panel with nested content (auto height)
	// ========================================
	fmt.Println("=== Test 3: Nested Panel in Stack (auto height) ===")
	fmt.Println()

	nestedPanel := panel.New().
		SetTitle("Nested Panel").
		SetContent(
			newstack.New(newstack.Column).
				SetChildrenList([]rtui.VNode{
					newtext.New("First line of nested content"),
					newtext.New("Second line of nested content"),
					newtext.New("Third line of nested content"),
				}),
		).
		SetWidth(60).
		SetHeight(0) // Height 0 means auto height

	instance3 := nestedPanel.CreateInstance()
	if panelInst3, ok := instance3.(*panel.PanelInstance); ok {
		// Set path for trace
		panelInst3.SetPath("/root/test3")

		// Measure with constraints
		constraints3 := layout.NewConstraints(0, 100, 0, 100)
		size3 := panelInst3.Measure(constraints3)
		fmt.Printf("Input constraints: 0-%d x 0-%d (loose)\n", constraints3.MaxWidth, constraints3.MaxHeight)
		fmt.Printf("Measured size: %dx%d (auto height)\n", size3.Width, size3.Height)
		fmt.Println()
	}

	// ========================================
	// Step 5: Panel with large content requiring wrap
	// ========================================
	fmt.Println("=== Test 4: Panel with text wrapping ===")
	fmt.Println()

	wrapPanel := panel.New().
		SetTitle("Text Wrap Panel").
		SetContent(newtext.New("This is a very long text that should wrap to multiple lines when displayed inside a panel.")).
		SetWidth(40).
		SetHeight(0) // Auto height

	instance4 := wrapPanel.CreateInstance()
	if panelInst4, ok := instance4.(*panel.PanelInstance); ok {
		// Set path for trace
		panelInst4.SetPath("/root/test4")

		// Measure with constraints
		constraints4 := layout.NewConstraints(0, 50, 0, 50)
		size4 := panelInst4.Measure(constraints4)
		fmt.Printf("Input constraints: 0-%d x 0-%d\n", constraints4.MaxWidth, constraints4.MaxHeight)
		fmt.Printf("Measured size: %dx%d\n", size4.Width, size4.Height)
		fmt.Println()
	}

	// ========================================
	// Step 6: Output constraint trace
	// ========================================
	fmt.Println("======================================================================")
	fmt.Println("                        CONSTRAINT PROPAGATION TRACE                  ")
	fmt.Println("======================================================================")
	fmt.Println()

	// Dump trace log
	traceLog := layout.DumpTrace()
	fmt.Println(traceLog)

	// Disable tracer
	layout.DisableTracer()

	fmt.Println("======================================================================")
	fmt.Println("                        TRACE DEMO COMPLETE                           ")
	fmt.Println("======================================================================")
}

// Package main demonstrates the constraint tracer integration.
// Note: Panel now uses composition-based implementation (VNode->Border->Stack)
// and no longer supports CreateInstance(). The constraint tracer can be used
// with other components that implement InstanceFactory.
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/layout"
)

func main() {
	// ========================================
	// Step 1: Enable constraint tracer
	// ========================================
	layout.EnableTracer()
	fmt.Println("=== Constraint Tracer Enabled ===")
	fmt.Println()

	// ========================================
	// Info: Panel is now composition-based
	// ========================================
	fmt.Println("=== Panel Composition-Based Architecture ===")
	fmt.Println()
	fmt.Println("Panel now uses composition:")
	fmt.Println("  Panel VNode -> Border VNode -> Stack VNode")
	fmt.Println()
	fmt.Println("This means:")
	fmt.Println("  - No separate PanelInstance needed")
	fmt.Println("  - Border component handles all layout state")
	fmt.Println("  - CreateInstance() is no longer available")
	fmt.Println()
	fmt.Println("For constraint tracing demos, use components")
	fmt.Println("that implement InstanceFactory directly.")
	fmt.Println()

	// ========================================
	// Step 2: Output constraint trace
	// ========================================
	fmt.Println("======================================================================")
	fmt.Println("                        CONSTRAINT PROPAGATION DEMO                   ")
	fmt.Println("======================================================================")
	fmt.Println()

	// Dump trace log
	traceLog := layout.DumpTrace()
	fmt.Println(traceLog)

	// Disable tracer
	layout.DisableTracer()

	fmt.Println("======================================================================")
	fmt.Println("                        DEMO COMPLETE                                 ")
	fmt.Println("======================================================================")

	// Exit with note
	fmt.Println("")
	fmt.Println("NOTE: The original tracer demo used Panel.CreateInstance()")
	fmt.Println("      which is no longer supported. Panel is now a composition")
	fmt.Println("      component that delegates to Border and Stack.")
	os.Exit(0)
}

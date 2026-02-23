// Package stack addon - Constraint tracing support
// This file extends Stack instances with constraint tracing capability.
package stack

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// measureChildFromInstance measures a single child with constraint tracing.
// This replaces the original measureChild method in a tracing-compatible way.
func (inst *Instance) measureChildWithTrace(child rtui.VNode, constraints layout.Constraints) layout.Size {
	if child == nil {
		return layout.Size{}
	}

	// Build stack and child IDs for tracing
	stackID := "stack"
	if inst.key != "" {
		stackID = fmt.Sprintf("stack(%s)", inst.key)
	} else if inst.direction == Row {
		stackID = "hstack"
	} else {
		stackID = "vstack"
	}

	childID := "child"
	if child != nil {
		childTag := child.Tag()
		childKey := ""
		if keyer, ok := child.(interface{ Key() string }); ok {
			childKey = keyer.Key()
		}
		if childKey != "" {
			childID = fmt.Sprintf("%s(%s)", childTag, childKey)
		} else {
			childID = childTag
		}
	}

	// Build child path
	childPath := inst.path
	if childPath == "" {
		childPath = stackID
	} else {
		childPath = fmt.Sprintf("%s/%s", childPath, childID)
	}

	// Debug log child measurement
	debugMeas := false
	if debugMeas {
		fmt.Printf("[Stack.measureChild] child=%s MaxW=%d MaxH=%d\n", childID, constraints.MaxWidth, constraints.MaxHeight)
	}

	// Trace constraint propagation (before measurement)
	directionStr := "column"
	if inst.direction == Row {
		directionStr = "row"
	}
	layout.TraceMeasuring(
		stackID,
		childID,
		childPath,
		layout.Constraints{}, // Input is not tracked here since we don't know original input
		constraints,          // Constraints passed to child
		layout.Size{},        // Size will be updated after measurement
		fmt.Sprintf("Stack(%s) -> %s: direction=%v, gap=%d, padding applied",
			directionStr, childID, directionStr, inst.gap),
	)

	// Perform the actual measurement using original measureChild
	return inst.measureChildOriginal(child, constraints)
}

// measureChildOriginal is the original measurement logic without tracing.
// This will replace the existing measureChild body for cleaner tracing integration.
func (inst *Instance) measureChildOriginal(child rtui.VNode, constraints layout.Constraints) layout.Size {
	if child == nil {
		return layout.Size{}
	}

	// Try InstanceFactory -> Measure
	if factory, ok := child.(rtui.InstanceFactory); ok {
		measInst := factory.CreateInstance()
		if measurable, ok := measInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			return measurable.Measure(constraints)
		}
	}

	// Try direct Measurable
	if measurable, ok := child.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(constraints)
	}

	// Fallback: estimate from content
	return inst.estimateChildSize(child, constraints)
}

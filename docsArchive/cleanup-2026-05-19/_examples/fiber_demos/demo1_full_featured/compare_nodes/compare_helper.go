// Helper functions for VNode vs Fiber tree comparison
package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/ui"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// ComparisonResult holds the results of comparing VNode and Fiber trees
type ComparisonResult struct {
	VNodeCount      int
	FiberCount      int
	MismatchCount   int
	Matches         []string
	Mismatches      []string
	Warnings        []string
	StructureValid  bool
	PreservationOK  bool
}

// CompareTrees compares a VNode tree with a Fiber tree
func CompareTrees(vnode ui.VNode, fiber *rtui.Fiber) *ComparisonResult {
	result := &ComparisonResult{
		Matches:    make([]string, 0),
		Mismatches: make([]string, 0),
		Warnings:   make([]string, 0),
	}

	// Count VNodes
	result.VNodeCount = countVNodes(vnode)

	// Count Fibers
	result.FiberCount = rtui.CountFibers(fiber)

	// Compare tree structure
	vnodeList := collectVNodes(vnode)
	fiberList := collectFibers(fiber)

	// Basic count comparison
	if result.VNodeCount != result.FiberCount {
		result.MismatchCount++
		result.Mismatches = append(result.Mismatches,
			fmt.Sprintf("Node count mismatch: VNode=%d, Fiber=%d",
				result.VNodeCount, result.FiberCount))
	}

	// Compare nodes by position
	maxCount := len(vnodeList)
	if len(fiberList) > maxCount {
		maxCount = len(fiberList)
	}

	for i := 0; i < maxCount; i++ {
		if i < len(vnodeList) && i < len(fiberList) {
			compareSingleNode(vnodeList[i], fiberList[i], result, i)
		} else if i < len(vnodeList) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("VNode[%d] missing in Fiber: %s", i, vnodeList[i].Tag()))
			result.MismatchCount++
		} else {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Fiber[%d] missing in VNode: %s", i, fiberList[i].Tag))
			result.MismatchCount++
		}
	}

	return result
}

// compareSingleNode compares a single VNode with a Fiber node
func compareSingleNode(vnode ui.VNode, fiber *rtui.Fiber, result *ComparisonResult, index int) {
	vtag := vnode.Tag()
	ftag := fiber.Tag

	// Compare Tag
	if vtag == ftag {
		result.Matches = append(result.Matches,
			fmt.Sprintf("[%d] Tag match: %s", index, vtag))
	} else {
		result.Mismatches = append(result.Mismatches,
			fmt.Sprintf("[%d] Tag mismatch: VNode=%s, Fiber=%s", index, vtag, ftag))
		result.MismatchCount++
	}

	// Compare Type
	vType := vnode.Type().String()
	fType := fiber.Type.String()
	if vType == fType {
		result.Matches = append(result.Matches,
			fmt.Sprintf("[%d] Type match: %s", index, vType))
	} else {
		result.Mismatches = append(result.Mismatches,
			fmt.Sprintf("[%d] Type mismatch: VNode=%s, Fiber=%s", index, vType, fType))
		result.MismatchCount++
	}

	// Compare Key
	vkey := vnode.Key()
	fkey := fiber.Key
	if vkey == fkey {
		result.Matches = append(result.Matches,
			fmt.Sprintf("[%d] Key match: %q", index, vkey))
	} else {
		if vkey != "" || fkey != "" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("[%d] Key mismatch: VNode=%q, Fiber=%q", index, vkey, fkey))
		}
	}

	// Check children count
	vchildren := vnode.Children()
	fchildren := getFiberChildren(fiber)
	if len(vchildren) == len(fchildren) {
		result.Matches = append(result.Matches,
			fmt.Sprintf("[%d] Children count match: %d", index, len(vchildren)))
	} else {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("[%d] Children count mismatch: VNode=%d, Fiber=%d",
				index, len(vchildren), len(fchildren)))
	}
}

// PrintComparisonResult prints detailed comparison results
func PrintComparisonResult(result *ComparisonResult) {
	fmt.Println("========================================")
	fmt.Println("     VNODE VS FIBER COMPARISON REPORT")
	fmt.Println("========================================")

	fmt.Printf("\n--- Overview ---\n")
	fmt.Printf("  VNode Count:     %d\n", result.VNodeCount)
	fmt.Printf("  Fiber Count:     %d\n", result.FiberCount)
	fmt.Printf("  Mismatches:      %d\n", result.MismatchCount)

	if len(result.Matches) > 0 {
		fmt.Printf("\n--- Matches (%d) ---\n", len(result.Matches))
		for i := 0; i < len(result.Matches) && i < 10; i++ {
			fmt.Printf("  ✓ %s\n", result.Matches[i])
		}
		if len(result.Matches) > 10 {
			fmt.Printf("  ... and %d more matches\n", len(result.Matches)-10)
		}
	}

	if len(result.Mismatches) > 0 {
		fmt.Printf("\n--- Mismatches (%d) ---\n", len(result.Mismatches))
		for _, m := range result.Mismatches {
			fmt.Printf("  ✗ %s\n", m)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\n--- Warnings (%d) ---\n", len(result.Warnings))
		for i := 0; i < len(result.Warnings) && i < 5; i++ {
			fmt.Printf("  ! %s\n", result.Warnings[i])
		}
		if len(result.Warnings) > 5 {
			fmt.Printf("  ... and %d more warnings\n", len(result.Warnings)-5)
		}
	}

	fmt.Println("========================================")
}

// CheckFiberPreservation checks that Fiber preserves VNode information
func CheckFiberPreservation(fiber *rtui.Fiber, depth int) {
	if fiber == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Check basic properties
	fmt.Printf("%s✓ Fiber #%d (Tag=%s, Type=%s)\n",
		indent, fiber.NodeID, fiber.Tag, fiber.Type.String())

	// Check NodeID
	if fiber.NodeID == 0 {
		fmt.Printf("%s  ⚠ Warning: NodeID is 0\n", indent)
	}

	// Check Style (Style is a struct, not a pointer)
	style := fiber.Style
	styleCount := 0
	if style.FG != "" {
		styleCount++
	}
	if style.BG != "" {
		styleCount++
	}
	if style.IsBold() {
		styleCount++
	}
	if style.IsItalic() {
		styleCount++
	}
	if style.IsUnderline() {
		styleCount++
	}
	if style.IsReverse() {
		styleCount++
	}
	if style.IsBlink() {
		styleCount++
	}
	fmt.Printf("%s  ✓ Style: %d properties set\n", indent, styleCount)

	// Check Props
	if fiber.Props != nil && len(fiber.Props) > 0 {
		fmt.Printf("%s  ✓ Props: %d keys\n", indent, len(fiber.Props))
	}

	// Check Instance
	if fiber.GetInstance() != nil {
		fmt.Printf("%s  ✓ Instance: %T\n", indent, fiber.GetInstance())
	} else {
		fmt.Printf("%s  ℹ No Instance (may be text/element)\n", indent)
	}

	// Check Layer
	fmt.Printf("%s  ✓ Layer: %d (%s)\n", indent, fiber.Layer, getLayerName(fiber.Layer))

	// Recursively check children
	for child := fiber.Child; child != nil; child = child.Sibling {
		CheckFiberPreservation(child, depth+1)
	}
}

// PrintSummary prints a summary of the comparison
func PrintSummary(result *ComparisonResult) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("              SUMMARY")
	fmt.Println(strings.Repeat("=", 50))

	if result.MismatchCount == 0 {
		fmt.Println("  ✓ SUCCESS: VNode and Fiber trees match!")
	} else {
		fmt.Printf("  ⚠ PARTIAL: %d mismatch(es) found\n", result.MismatchCount)
	}

	fmt.Printf("\n  Statistical Summary:\n")
	fmt.Printf("    - Total VNodes:   %d\n", result.VNodeCount)
	fmt.Printf("    - Total Fibers:   %d\n", result.FiberCount)
	fmt.Printf("    - Matches:        %d\n", len(result.Matches))
	fmt.Printf("    - Mismatches:     %d\n", len(result.Mismatches))
	fmt.Printf("    - Warnings:       %d\n", len(result.Warnings))

	fmt.Println(strings.Repeat("=", 50))
}

// Helper functions

func countVNodes(vnode ui.VNode) int {
	if vnode == nil {
		return 0
	}
	count := 1
	children := vnode.Children()
	for _, child := range children {
		count += countVNodes(child)
	}
	return count
}

func collectVNodes(vnode ui.VNode) []ui.VNode {
	var nodes []ui.VNode
	var collect func(ui.VNode)
	collect = func(v ui.VNode) {
		if v == nil {
			return
		}
		nodes = append(nodes, v)
		for _, child := range v.Children() {
			collect(child)
		}
	}
	collect(vnode)
	return nodes
}

func collectFibers(fiber *rtui.Fiber) []*rtui.Fiber {
	var fibers []*rtui.Fiber
	rtui.WalkFiberDepthFirst(fiber, func(f *rtui.Fiber) bool {
		fibers = append(fibers, f)
		return true
	})
	return fibers
}

func getFiberChildren(fiber *rtui.Fiber) []*rtui.Fiber {
	var children []*rtui.Fiber
	for child := fiber.Child; child != nil; child = child.Sibling {
		children = append(children, child)
	}
	return children
}

func getLayerName(layer rtui.Layer) string {
	switch layer {
	case rtui.LayerBase:
		return "Base"
	case rtui.LayerOverlay:
		return "Overlay"
	case rtui.LayerModal:
		return "Modal"
	case rtui.LayerTooltip:
		return "Tooltip"
	case rtui.LayerInspector:
		return "Inspector"
	default:
		return fmt.Sprintf("Unknown(%d)", layer)
	}
}

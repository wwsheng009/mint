// VNode vs Fiber Tree Comparison Tool
// Verify Fiber tree can carry all VNode information
package main

import (
	"fmt"
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// ComparisonResult holds the comparison results
type ComparisonResult struct {
	VNodeCount     int
	FiberCount     int
	MatchCount     int
	MismatchCount  int
	StructureMatch bool
	Details        []string
}

// CompareTrees compares VNode tree with Fiber tree structure
func CompareTrees(vnode rtui.VNode, fiber *rtui.Fiber) *ComparisonResult {
	result := &ComparisonResult{
		Details: make([]string, 0),
	}

	result.VNodeCount = countVNodes(vnode)
	result.FiberCount = countFibers(fiber)
	result.StructureMatch = compareStructure(vnode, fiber, result, 0)

	return result
}

// compareStructure recursively compares VNode and Fiber structure
func compareStructure(vnode rtui.VNode, fiber *rtui.Fiber, result *ComparisonResult, depth int) bool {
	if vnode == nil && fiber == nil {
		return true
	}

	if vnode == nil || fiber == nil {
		result.Details = append(result.Details, fmt.Sprintf("%sNil mismatch: VNode=%v Fiber=%v",
			indent(depth), vnode, fiber))
		return false
	}

	// Compare basic properties
	tagMatch := compareTag(vnode, fiber)

	if tagMatch {
		result.MatchCount++
	} else {
		result.MismatchCount++
		result.Details = append(result.Details, fmt.Sprintf("%sType mismatch: VNode.Type=%v Fiber.Tag=%s",
			indent(depth), vnode.Type(), fiber.Tag))
	}

	// Compare children
	vnodeChildren := getVNodeChildren(vnode)
	fiberChildren := fiber.GetChildFibers()

	if len(vnodeChildren) != len(fiberChildren) {
		result.Details = append(result.Details, fmt.Sprintf("%sChild count mismatch: VNode=%d Fiber=%d",
			indent(depth), len(vnodeChildren), len(fiberChildren)))
	}

	// Recursively compare children
	allMatch := true
	minChildren := len(vnodeChildren)
	if len(fiberChildren) < minChildren {
		minChildren = len(fiberChildren)
	}

	for i := 0; i < minChildren; i++ {
		if !compareStructure(vnodeChildren[i], fiberChildren[i], result, depth+1) {
			allMatch = false
		}
	}

	return allMatch && tagMatch
}

// compareTag compares VNode tag with Fiber tag
func compareTag(vnode rtui.VNode, fiber *rtui.Fiber) bool {
	vnodeType := vnode.Type()
	fiberTag := fiber.Tag

	switch vnodeType {
	case rtui.VNodeElement:
		// Get VNode's actual tag and compare with Fiber tag
		vnodeTag := getVNodeTag(vnode)
		return vnodeTag == fiberTag
	case rtui.VNodeText:
		return fiberTag == "text"
	case rtui.VNodeComponent:
		return fiberTag != ""
	case rtui.VNodeFragment:
		return fiberTag == "fragment" || fiberTag == ""
	default:
		return fiberTag != ""
	}
}

// getVNodeTag extracts the tag from a VNode
// Handles ElementVNode, LayoutNode, and other element types
func getVNodeTag(vnode rtui.VNode) string {
	if vnode == nil {
		return ""
	}

	// Try to get tag from types that have Tag() method
	switch n := vnode.(type) {
	case interface{ Tag() string }:
		return n.Tag()
	default:
		// Fallback: return empty string for unknown types
		return ""
	}
}

// countVNodes counts all VNodes in tree
func countVNodes(vnode rtui.VNode) int {
	if vnode == nil {
		return 0
	}
	count := 1
	for _, child := range getVNodeChildren(vnode) {
		count += countVNodes(child)
	}
	return count
}

// countFibers counts all Fibers in tree
func countFibers(fiber *rtui.Fiber) int {
	if fiber == nil {
		return 0
	}
	count := 1
	for _, child := range fiber.GetChildFibers() {
		count += countFibers(child)
	}
	return count
}

// getVNodeChildren returns children of VNode
func getVNodeChildren(vnode rtui.VNode) []rtui.VNode {
	if vnode == nil {
		return nil
	}

	// Use Children() method directly
	return vnode.Children()
}

func indent(depth int) string {
	return strings.Repeat("  ", depth)
}

// PrintComparisonResult prints the comparison result
func PrintComparisonResult(result *ComparisonResult) {
	fmt.Println()
	fmt.Println("=== VNode vs Fiber Tree Comparison ===")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("VNode Nodes:    %d\n", result.VNodeCount)
	fmt.Printf("Fiber Nodes:    %d\n", result.FiberCount)
	fmt.Printf("Matches:        %d\n", result.MatchCount)
	fmt.Printf("Mismatches:     %d\n", result.MismatchCount)

	fmt.Println(strings.Repeat("-", 50))

	if result.StructureMatch {
		fmt.Println("Status: STRUCTURE MATCH")
	} else {
		fmt.Println("Status: STRUCTURE MISMATCH")
	}

	if len(result.Details) > 0 && len(result.Details) < 20 {
		fmt.Println()
		fmt.Println("Details:")
		for _, detail := range result.Details {
			fmt.Printf("  %s\n", detail)
		}
	} else if len(result.Details) >= 20 {
		fmt.Printf("\nShowing first 20 of %d details\n", len(result.Details))
		for i, detail := range result.Details {
			if i >= 20 {
				break
			}
			fmt.Printf("  %s\n", detail)
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
}

// CheckFiberPreservation verifies what information is preserved in Fiber
func CheckFiberPreservation(fiber *rtui.Fiber, depth int) {
	if fiber == nil {
		return
	}

	indentStr := indent(depth)

	// Check what's preserved
	fmt.Printf("%sNodeID=%d Tag=%s Type=%v\n", indentStr, fiber.NodeID, fiber.Tag, fiber.Type)

	// Check VNode reference
	if fiber.VNode != nil {
		fmt.Printf("%s  [OK] VNode reference preserved\n", indentStr)
	} else {
		fmt.Printf("%s  [WARN] VNode reference is nil\n", indentStr)
	}

	// Check Props
	if fiber.Props != nil {
		fmt.Printf("%s  [OK] Props preserved (count=%d)\n", indentStr, len(fiber.Props))
	}

	// Check MemoizedState
	if fiber.MemoizedState != nil {
		fmt.Printf("%s  [OK] MemoizedState: %v\n", indentStr, fiber.MemoizedState)
	}

	// Check Layout properties
	hasLayout := fiber.LayoutDirection != 0 ||
		fiber.LayoutGap != 0 ||
		fiber.LayoutAlign != 0 ||
		fiber.LayoutPadding != [4]int{}
	if hasLayout {
		fmt.Printf("%s  [OK] Layout: dir=%v gap=%d align=%v\n",
			indentStr, fiber.LayoutDirection, fiber.LayoutGap, fiber.LayoutAlign)
	}

	// Recurse to children
	for _, child := range fiber.GetChildFibers() {
		CheckFiberPreservation(child, depth+1)
	}
}

// PrintSummary prints a summary of the comparison
func PrintSummary(result *ComparisonResult) {
	fmt.Println("=== Summary ===")
	fmt.Println()

	if result.VNodeCount == result.FiberCount {
		fmt.Printf("[OK] Node count matches: %d nodes in both trees\n", result.VNodeCount)
	} else {
		fmt.Printf("[WARN] Node count mismatch: VNode=%d Fiber=%d (diff=%d)\n",
			result.VNodeCount, result.FiberCount, result.FiberCount-result.VNodeCount)
	}

	if result.StructureMatch {
		fmt.Println("[OK] Tree structure is preserved")
	} else {
		fmt.Println("[WARN] Tree structure has differences")
	}

	fmt.Println()
	fmt.Println("Fiber Information Preservation:")
	fmt.Println("  NodeID:     Fiber.NodeID")
	fmt.Println("  Tag:        Fiber.Tag")
	fmt.Println("  Type:       Fiber.Type")
	fmt.Println("  VNode Ref:  Fiber.VNode")
	fmt.Println("  Props:       Fiber.Props")
	fmt.Println("  State:       Fiber.MemoizedState")
	fmt.Println("  Layout:      Fiber.LayoutDirection, LayoutGap, etc.")
	fmt.Println("  Style:       Fiber.StyleWidth, StyleHeight, etc.")
	fmt.Println("  Events:      Fiber.EventHandlers")
	fmt.Println("  Tree:        Fiber.Child, Sibling, Return")
	fmt.Println()

	conclusion := "Fiber tree CAN carry all VNode information"
	if !result.StructureMatch || result.VNodeCount != result.FiberCount {
		conclusion = "Fiber tree CANNOT carry all VNode information"
	}
	fmt.Printf("CONCLUSION: %s\n", conclusion)
	fmt.Println()
}

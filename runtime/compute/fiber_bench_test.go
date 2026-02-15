// Package compute provides benchmarks for Fiber-first layout
package compute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Fiber-First Layout Benchmarks
// =============================================================================

// BenchmarkVNodeLayout measures VNode-based layout performance
func BenchmarkVNodeLayout(b *testing.B) {
	engine := NewEngine()

	// Create a moderately complex VNode tree
	vnodeTree := createBenchmarkTree(10) // 10 levels deep

	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth:  0, MinHeight: 0,
		MaxWidth: 200, MaxHeight: 200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Layout(vnodeTree, fiberTree, constraints)
	}
}

// BenchmarkFiberLayout measures Fiber-first layout performance
func BenchmarkFiberLayout(b *testing.B) {
	engine := NewEngine()

	// Create same tree structure
	vnodeTree := createBenchmarkTree(10)
	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth: 0, MinHeight: 0,
		MaxWidth: 200, MaxHeight: 200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.LayoutFiber(fiberTree, constraints)
	}
}

// BenchmarkFiberVsVNodeLayout compares both approaches
func BenchmarkFiberVsVNodeLayout(b *testing.B) {
	engine := NewEngine()

	vnodeTree := createBenchmarkTree(10)
	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth: 0, MinHeight: 0,
		MaxWidth: 200, MaxHeight: 200,
	}

	b.Run("VNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = engine.Layout(vnodeTree, fiberTree, constraints)
		}
	})

	b.Run("Fiber", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = engine.LayoutFiber(fiberTree, constraints)
		}
	})
}

// BenchmarkDeepFiberLayout tests performance with deep nesting
func BenchmarkDeepFiberLayout(b *testing.B) {
	engine := NewEngine()

	vnodeTree := createBenchmarkTree(20) // 20 levels deep
	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth: 0, MinHeight: 0,
		MaxWidth: 200, MaxHeight: 200,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.LayoutFiber(fiberTree, constraints)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// createBenchmarkTree creates a nested VNode tree for benchmarking
// depth: number of nesting levels
func createBenchmarkTree(depth int) rtui.VNode {
	var buildFunc func(currentDepth int) rtui.VNode

	buildFunc = func(currentDepth int) rtui.VNode {
		if currentDepth >= depth {
			// Leaf node
			return rtui.Element("text").Prop("content", "leaf").Build()
		}

		// Create nested HStack with multiple children
		children := []rtui.VNode{
			rtui.Element("text").Prop("content", "a").Build(),
			rtui.Element("text").Prop("content", "b").Build(),
			buildFunc(currentDepth + 1),
		}

		return rtui.Element("hstack").Children(children...).Build()
	}

	return buildFunc(0)
}

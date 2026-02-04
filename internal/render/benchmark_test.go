// Package render provides performance benchmarks for VNode rendering.
package render

import (
	"fmt"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/framework/event"
)

// =============================================================================
// VNode Creation Benchmarks
// =============================================================================

// BenchmarkVNode_SimpleText benchmarks simple text VNode creation
func BenchmarkVNode_SimpleText(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.Element("text").Prop("content", "Hello, World!").Build()
	}
}

// BenchmarkVNode_NestedElements benchmarks nested element creation
func BenchmarkVNode_NestedElements(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.VStack(
			rtui.HStack(
				rtui.Element("text").Prop("content", "A").Build(),
				rtui.Element("text").Prop("content", "B").Build(),
			),
			rtui.Element("text").Prop("content", "C").Build(),
		)
	}
}

// BenchmarkVNode_DeepNesting benchmarks deeply nested VNode creation
func BenchmarkVNode_DeepNesting(b *testing.B) {
	depth := 10
	var nested func(int) rtui.VNode
	nested = func(d int) rtui.VNode {
		if d == 0 {
			return rtui.Element("text").Prop("content", "Leaf").Build()
		}
		return rtui.VStack(
			rtui.Element("text").Prop("content", "Level").Build(),
			nested(d-1),
		)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nested(depth)
	}
}

// BenchmarkVNode_WideHStack benchmarks wide horizontal stack
func BenchmarkVNode_WideHStack(b *testing.B) {
	width := 50
	children := make([]rtui.VNode, width)
	for i := 0; i < width; i++ {
		children[i] = rtui.Element("text").Prop("content", "Item").Build()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.HStack(children...)
	}
}

// BenchmarkVNode_WideVStack benchmarks wide vertical stack
func BenchmarkVNode_WideVStack(b *testing.B) {
	height := 50
	children := make([]rtui.VNode, height)
	for i := 0; i < height; i++ {
		children[i] = rtui.Element("text").Prop("content", "Line").Build()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.VStack(children...)
	}
}

// BenchmarkVNode_Fragment benchmarks Fragment creation
func BenchmarkVNode_Fragment(b *testing.B) {
	children := []rtui.VNode{
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.Fragment(children...)
	}
}

// =============================================================================
// VNodeRenderer Measure Benchmarks
// =============================================================================

func setupRendererForBenchmark(b *testing.B) (*NonFiberRenderer, *FiberRenderer) {
	node := NewDeclarativeNodeFromFunc(func() rtui.VNode {
		return rtui.Element("text").Prop("content", "test").Build()
	})
	nonFiberRenderer := node.GetRenderer().(*NonFiberRenderer)
	fiberRenderer := NewFiberRenderer(nil)
	return nonFiberRenderer, fiberRenderer
}

// BenchmarkMeasure_NonFiber_Text benchmarks NonFiberRenderer.Measure for text
func BenchmarkMeasure_NonFiber_Text(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.Element("text").Prop("content", "Hello, World!").Build()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Measure(vnode)
	}
}

// BenchmarkMeasure_NonFiber_HStack benchmarks measuring HStack
func BenchmarkMeasure_NonFiber_HStack(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.HStack(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Measure(vnode)
	}
}

// BenchmarkMeasure_NonFiber_VStack benchmarks measuring VStack
func BenchmarkMeasure_NonFiber_VStack(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.VStack(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Measure(vnode)
	}
}

// BenchmarkMeasure_NonFiber_Fragment benchmarks measuring Fragment
func BenchmarkMeasure_NonFiber_Fragment(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.Fragment(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Measure(vnode)
	}
}

// BenchmarkMeasure_Fiber_Text benchmarks FiberRenderer.Measure for text
func BenchmarkMeasure_Fiber_Text(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.Element("text").Prop("content", "Hello, World!").Build()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Measure(vnode)
	}
}

// BenchmarkMeasure_Fiber_HStack benchmarks FiberRenderer measuring HStack
func BenchmarkMeasure_Fiber_HStack(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.HStack(
		rtui.Element("text").Prop("content", "A").Build(),
		rtui.Element("text").Prop("content", "B").Build(),
		rtui.Element("text").Prop("content", "C").Build(),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.Measure(vnode)
	}
}

// BenchmarkMeasure_Comparison_NonFiber_Vs_Fiber compares both renderers
func BenchmarkMeasure_Comparison_NonFiber_Vs_Fiber(b *testing.B) {
	vnode := rtui.Element("text").Prop("content", "Hello").Build()

	b.Run("NonFiber", func(b *testing.B) {
		_, renderer := setupRendererForBenchmark(b)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = renderer.Measure(vnode)
		}
	})

	b.Run("Fiber", func(b *testing.B) {
		_, renderer := setupRendererForBenchmark(b)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = renderer.Measure(vnode)
		}
	})
}

// =============================================================================
// CollectFocusable Benchmarks
// =============================================================================

// BenchmarkCollectFocusable_Empty benchmarks collecting from empty tree
func BenchmarkCollectFocusable_Empty(b *testing.B) {
	root := rtui.VStack()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.CollectFocusable(root)
	}
}

// BenchmarkCollectFocusable_Single benchmarks collecting single focusable
func BenchmarkCollectFocusable_Single(b *testing.B) {
	root := rtui.VStack(
		rtui.Element("button").Prop("label", "Click").Build(),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.CollectFocusable(root)
	}
}

// BenchmarkCollectFocusable_Small benchmarks collecting 10 focusable nodes
func BenchmarkCollectFocusable_Small(b *testing.B) {
	nodes := make([]rtui.VNode, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = rtui.Element("button").Prop("label", "Button").Build()
	}
	root := rtui.VStack(nodes...)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.CollectFocusable(root)
	}
}

// BenchmarkCollectFocusable_Medium benchmarks collecting 50 focusable nodes
func BenchmarkCollectFocusable_Medium(b *testing.B) {
	nodes := make([]rtui.VNode, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = rtui.Element("button").Prop("label", "Button").Build()
	}
	root := rtui.VStack(nodes...)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.CollectFocusable(root)
	}
}

// BenchmarkCollectFocusable_Large benchmarks collecting 100 focusable nodes
func BenchmarkCollectFocusable_Large(b *testing.B) {
	nodes := make([]rtui.VNode, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = rtui.Element("button").Prop("label", "Button").Build()
	}
	root := rtui.HStack(nodes...)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.CollectFocusable(root)
	}
}

// BenchmarkCollectFocusable_Nested benchmarks collecting from nested structure
func BenchmarkCollectFocusable_Nested(b *testing.B) {
	// Create nested structure with focusable at different levels
	nodes := make([]rtui.VNode, 30)
	for i := 0; i < 30; i++ {
		nodes[i] = rtui.Element("button").Prop("label", "Button").Build()
	}
	root := rtui.VStack(
		rtui.HStack(nodes[0:10]...),
		rtui.HStack(nodes[10:20]...),
		rtui.HStack(nodes[20:30]...),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.CollectFocusable(root)
	}
}

// =============================================================================
// FocusManager Benchmarks
// =============================================================================

// BenchmarkFocusManager_SetFocusable benchmarks SetFocusable with preservation
func BenchmarkFocusManager_SetFocusable(b *testing.B) {
	m := ui.NewVNodeFocusManager()
	nodes := make([]ui.FocusableVNode, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = createMockFocusable("btn", i)
	}
	m.SetFocusable(nodes)
	m.SetFocusByIndex(5) // Set focus to middle

	newNodes := make([]ui.FocusableVNode, 10)
	for i := 0; i < 10; i++ {
		newNodes[i] = createMockFocusable("btn", i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetFocusable(newNodes)
	}
}

// BenchmarkFocusManager_FocusNext benchmarks FocusNext
func BenchmarkFocusManager_FocusNext(b *testing.B) {
	m := ui.NewVNodeFocusManager()
	nodes := make([]ui.FocusableVNode, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = createMockFocusable("btn", i)
	}
	m.SetFocusable(nodes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.FocusNext()
	}
}

// BenchmarkFocusManager_FocusPrev benchmarks FocusPrev
func BenchmarkFocusManager_FocusPrev(b *testing.B) {
	m := ui.NewVNodeFocusManager()
	nodes := make([]ui.FocusableVNode, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = createMockFocusable("btn", i)
	}
	m.SetFocusable(nodes)
	m.SetFocusByIndex(25) // Start from middle

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.FocusPrev()
	}
}

// BenchmarkFocusManager_SetFocusByID benchmarks SetFocusByID
func BenchmarkFocusManager_SetFocusByID(b *testing.B) {
	m := ui.NewVNodeFocusManager()
	nodes := make([]ui.FocusableVNode, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = createMockFocusable("btn", i)
	}
	m.SetFocusable(nodes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.SetFocusByID("btn-50")
	}
}

// BenchmarkFocusManager_HandleEvent_Tab benchmarks Tab key handling
func BenchmarkFocusManager_HandleEvent_Tab(b *testing.B) {
	m := ui.NewVNodeFocusManager()
	nodes := make([]ui.FocusableVNode, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = createMockFocusable("btn", i)
	}
	m.SetFocusable(nodes)

	tabEvent := &event.KeyEvent{Special: event.KeyTab}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.HandleEvent(tabEvent)
	}
}

// =============================================================================
// GetLayoutInfo Benchmarks
// =============================================================================

// BenchmarkGetLayoutInfo_HStack benchmarks GetLayoutInfo for HStack
func BenchmarkGetLayoutInfo_HStack(b *testing.B) {
	vnode := rtui.HStack()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.GetLayoutInfo(vnode)
	}
}

// BenchmarkGetLayoutInfo_VStack benchmarks GetLayoutInfo for VStack
func BenchmarkGetLayoutInfo_VStack(b *testing.B) {
	vnode := rtui.VStack()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.GetLayoutInfo(vnode)
	}
}

// BenchmarkGetLayoutInfo_Element benchmarks GetLayoutInfo for plain element
func BenchmarkGetLayoutInfo_Element(b *testing.B) {
	vnode := rtui.Element("div").Build()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.GetLayoutInfo(vnode)
	}
}

// =============================================================================
// GetTextContent Benchmarks
// =============================================================================

// BenchmarkGetTextContent_WithContent benchmarks GetTextContent with content prop
func BenchmarkGetTextContent_WithContent(b *testing.B) {
	vnode := rtui.Element("text").Prop("content", "Hello, World!").Build()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.GetTextContent(vnode)
	}
}

// BenchmarkGetTextContent_WithoutContent benchmarks GetTextContent without content
func BenchmarkGetTextContent_WithoutContent(b *testing.B) {
	vnode := rtui.Element("div").Build()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rtui.GetTextContent(vnode)
	}
}

// =============================================================================
// Props Operations Benchmarks
// =============================================================================

// BenchmarkProps_GetString benchmarks Props.GetString
func BenchmarkProps_GetString(b *testing.B) {
	props := rtui.Props{
		"id":          "test-id",
		"class":       "test-class",
		"data-value":  "42",
		"data-label":  "test",
		"width":       10,
		"height":      5,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = props.GetString("id")
	}
}

// BenchmarkProps_Set benchmarks Props.Set
func BenchmarkProps_Set(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		props := make(rtui.Props)
		props.Set("key", "value")
		props.Set("number", 42)
		_ = props
	}
}

// =============================================================================
// Style Operations Benchmarks
// =============================================================================

// BenchmarkStyle_Creation benchmarks Style creation
func BenchmarkStyle_Creation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = style.NewStyle()
	}
}

// BenchmarkStyle_CreationWithProps benchmarks Style creation with properties
func BenchmarkStyle_CreationWithProps(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = style.NewStyle().
			Foreground(style.Red).
			Background(style.Blue).
			Bold(true).
			Underline(true)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// createMockFocusable creates a mock focusable node for benchmarking
func createMockFocusable(prefix string, index int) ui.FocusableVNode {
	return &mockFocusableNode{
		ElementVNode: rtui.NewElement("button"),
		id:           fmt.Sprintf("%s-%d", prefix, index),
		label:        fmt.Sprintf("Button %d", index),
		isFocusable:  true,
	}
}

// mockFocusableNode is a lightweight mock for benchmarking
type mockFocusableNode struct {
	*rtui.ElementVNode
	id          string
	label       string
	isFocusable bool
	hasFocus    bool
}

func (m *mockFocusableNode) SetFocus(hasFocus bool) {
	m.hasFocus = hasFocus
}

func (m *mockFocusableNode) IsFocusable() bool {
	return m.isFocusable
}

func (m *mockFocusableNode) GetFocusID() string {
	return m.id
}

func (m *mockFocusableNode) Label() string {
	return m.label
}

// =============================================================================
// Parallel Benchmarks
// =============================================================================

// BenchmarkParallel_Measure benchmarks parallel measurement
func BenchmarkParallel_Measure(b *testing.B) {
	_, renderer := setupRendererForBenchmark(b)
	vnode := rtui.Element("text").Prop("content", "Hello").Build()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = renderer.Measure(vnode)
		}
	})
}

// BenchmarkParallel_CollectFocusable benchmarks parallel CollectFocusable
func BenchmarkParallel_CollectFocusable(b *testing.B) {
	nodes := make([]rtui.VNode, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = rtui.Element("button").Prop("label", "Button").Build()
	}
	root := rtui.VStack(nodes...)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ui.CollectFocusable(root)
		}
	})
}

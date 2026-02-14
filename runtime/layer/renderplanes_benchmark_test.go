package layer

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// BenchmarkBuildFromFiber_100Nodes tests building RenderPlanes from 100 nodes
func BenchmarkBuildFromFiber_100Nodes(b *testing.B) {
	fiberTree := createTestFiberTree(100, 5, rtui.LayerBase)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildFromFiber(fiberTree)
	}
}

// BenchmarkBuildFromFiber_1000Nodes tests building RenderPlanes from 1000 nodes
func BenchmarkBuildFromFiber_1000Nodes(b *testing.B) {
	fiberTree := createTestFiberTree(1000, 10, rtui.LayerBase)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildFromFiber(fiberTree)
	}
}

// BenchmarkBuildFromFiber_10000Nodes tests building RenderPlanes from 10000 nodes
func BenchmarkBuildFromFiber_10000Nodes(b *testing.B) {
	fiberTree := createTestFiberTree(10000, 10, rtui.LayerBase)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildFromFiber(fiberTree)
	}
}

// BenchmarkBuildFromFiber_MultiLayer tests building RenderPlanes with multiple layers
func BenchmarkBuildFromFiber_MultiLayer(b *testing.B) {
	fiberTree := createMultiLayerFiberTree(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildFromFiber(fiberTree)
	}
}

// BenchmarkRenderPlanesAddToLayer tests adding boxes to layers
func BenchmarkRenderPlanesAddToLayer(b *testing.B) {
	rp := NewRenderPlanes()
	box := &compute.ComputedBox{
		Box:    runtime.Box{X: 0, Y: 0, Width: 10, Height: 10},
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.NodeID = uint64(i)
		rp.AddToLayer(rtui.LayerBase, box)
	}
}

// BenchmarkRenderPlanesIterate tests iterating over planes
func BenchmarkRenderPlanesIterate(b *testing.B) {
	rp := NewRenderPlanes()
	fiberTree := createTestFiberTree(1000, 10, rtui.LayerBase)
	rp = BuildFromFiber(fiberTree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		rp.Iterate(func(layer rtui.Layer, box *compute.ComputedBox) bool {
			count++
			return true
		})
		_ = count
	}
}

// BenchmarkRenderPlanesIterateReverse tests iterating in reverse order
func BenchmarkRenderPlanesIterateReverse(b *testing.B) {
	rp := NewRenderPlanes()
	fiberTree := createTestFiberTree(1000, 10, rtui.LayerBase)
	rp = BuildFromFiber(fiberTree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		rp.IterateReverse(func(layer rtui.Layer, box *compute.ComputedBox) bool {
			count++
			return true
		})
		_ = count
	}
}

// BenchmarkRenderPlanesHasLayer tests layer existence checks
func BenchmarkRenderPlanesHasLayer(b *testing.B) {
	rp := NewRenderPlanes()
	fiberTree := createMultiLayerFiberTree(1000)
	rp = BuildFromFiber(fiberTree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rp.HasLayer(rtui.LayerModal)
		_ = rp.HasLayer(rtui.LayerOverlay)
		_ = rp.HasLayer(rtui.LayerBase)
	}
}

// BenchmarkRenderPlanesGetHighestLayer tests finding highest layer
func BenchmarkRenderPlanesGetHighestLayer(b *testing.B) {
	rp := NewRenderPlanes()
	fiberTree := createMultiLayerFiberTree(1000)
	rp = BuildFromFiber(fiberTree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rp.GetHighestLayer()
	}
}

// BenchmarkRenderPlanesCountBoxes tests counting boxes
func BenchmarkRenderPlanesCountBoxes(b *testing.B) {
	rp := NewRenderPlanes()
	fiberTree := createTestFiberTree(1000, 10, rtui.LayerBase)
	rp = BuildFromFiber(fiberTree)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rp.CountBoxes()
	}
}

// Helper function to create a test fiber tree with specified node count
func createTestFiberTree(nodeCount, branchingFactor int, defaultLayer rtui.Layer) *rtui.Fiber {
	if nodeCount <= 0 {
		return nil
	}

	root := &rtui.Fiber{
		Layer:  defaultLayer,
		NodeID: 1,
	}
	current := root

	for i := 2; i <= nodeCount; i++ {
		// Create child
		child := &rtui.Fiber{
			Layer:  defaultLayer,
			NodeID: uint64(i),
		}

		if current.Child == nil {
			current.Child = child
		} else {
			// Find last sibling and attach
			sibling := current.Child
			for sibling.Sibling != nil {
				sibling = sibling.Sibling
			}
			sibling.Sibling = child
		}

		// Move to next level when we reach branching factor
		if i%branchingFactor == 0 {
			if child.Child == nil && i+1 <= nodeCount {
				// Make next node a child of this one
				current = child
			}
		}
	}

	return root
}

// Helper function to create a multi-layer fiber tree
func createMultiLayerFiberTree(nodeCount int) *rtui.Fiber {
	if nodeCount <= 0 {
		return nil
	}

	layers := []rtui.Layer{
		rtui.LayerBase,
		rtui.LayerOverlay,
		rtui.LayerModal,
		rtui.LayerTooltip,
		rtui.LayerInspector,
	}

	root := &rtui.Fiber{
		Layer:  rtui.LayerBase,
		NodeID: 1,
	}
	current := root

	for i := 2; i <= nodeCount; i++ {
		layer := layers[i%len(layers)]

		child := &rtui.Fiber{
			Layer:  layer,
			NodeID: uint64(i),
		}

		if current.Child == nil {
			current.Child = child
		} else {
			sibling := current.Child
			for sibling.Sibling != nil {
				sibling = sibling.Sibling
			}
			sibling.Sibling = child
		}

		if i%10 == 0 && child.Child == nil && i+1 <= nodeCount {
			current = child
		}
	}

	return root
}

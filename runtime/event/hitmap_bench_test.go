package event

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
)

// ============================================================================
// Phase 1-7: 基准测试 - 性能基准
// ============================================================================

// BenchmarkHitMap_Build_100 测试构建 100 个节点的 HitMap
func BenchmarkHitMap_Build_100(b *testing.B) {
	nodes := createFlatNodeTree(100)
	root := createRoot(nodes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHitMap(root)
	}
}

// BenchmarkHitMap_Build_1000 测试构建 1000 个节点的 HitMap
func BenchmarkHitMap_Build_1000(b *testing.B) {
	nodes := createFlatNodeTree(1000)
	root := createRoot(nodes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHitMap(root)
	}
}

// BenchmarkHitMap_Build_10000 测试构建 10000 个节点的 HitMap
func BenchmarkHitMap_Build_10000(b *testing.B) {
	nodes := createFlatNodeTree(10000)
	root := createRoot(nodes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHitMap(root)
	}
}

// BenchmarkHitMap_HitTest_100 测试 100 个节点的命中测试性能
func BenchmarkHitMap_HitTest_100(b *testing.B) {
	nodes := createFlatNodeTree(100)
	root := createRoot(nodes)
	hitMap := BuildHitMap(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := (i * 7) % 1000
		y := (i * 13) % 1000
		hitMap.HitTest(x, y)
	}
}

// BenchmarkHitMap_HitTest_1000 测试 1000 个节点的命中测试性能
func BenchmarkHitMap_HitTest_1000(b *testing.B) {
	nodes := createFlatNodeTree(1000)
	root := createRoot(nodes)
	hitMap := BuildHitMap(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := (i * 7) % 1000
		y := (i * 13) % 1000
		hitMap.HitTest(x, y)
	}
}

// BenchmarkHitMap_FindByID_100 测试 100 个节点的 ID 查找性能
func BenchmarkHitMap_FindByID_100(b *testing.B) {
	nodes := createFlatNodeTree(100)
	root := createRoot(nodes)
	hitMap := BuildHitMap(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeID := fmt.Sprintf("node-%d", i%100)
		hitMap.FindByID(nodeID)
	}
}

// BenchmarkHitMap_FindByID_1000 测试 1000 个节点的 ID 查找性能
func BenchmarkHitMap_FindByID_1000(b *testing.B) {
	nodes := createFlatNodeTree(1000)
	root := createRoot(nodes)
	hitMap := BuildHitMap(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeID := fmt.Sprintf("node-%d", i%1000)
		hitMap.FindByID(nodeID)
	}
}

// BenchmarkHitMap_FindAllAt_100 测试 100 个节点的 FindAllAt 性能
func BenchmarkHitMap_FindAllAt_100(b *testing.B) {
	nodes := createFlatNodeTree(100)
	root := createRoot(nodes)
	hitMap := BuildHitMap(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := (i * 7) % 1000
		y := (i * 13) % 1000
		hitMap.FindAllAt(x, y)
	}
}

// BenchmarkHitMap_FindAllAt_1000 测试 1000 个节点的 FindAllAt 性能
func BenchmarkHitMap_FindAllAt_1000(b *testing.B) {
	nodes := createFlatNodeTree(1000)
	root := createRoot(nodes)
	hitMap := BuildHitMap(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := (i * 7) % 1000
		y := (i * 13) % 1000
		hitMap.FindAllAt(x, y)
	}
}

// BenchmarkHitMap_LocalXY 测试局部坐标转换性能
func BenchmarkHitMap_LocalXY(b *testing.B) {
	node := &mockNode{
		id:       "test",
		nodeType: "test",
		x:        100,
		y:        100,
		width:    200,
		height:   200,
	}

	hitMap := BuildHitMap(node)
	entry := hitMap.FindByID("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := 100 + (i % 200)
		y := 100 + (i % 200)
		entry.LocalXY(x, y)
	}
}

// 辅助函数

// createFlatNodeTree 创建一个扁平的节点树
func createFlatNodeTree(count int) []layout.Node {
	nodes := make([]layout.Node, count)
	for i := 0; i < count; i++ {
		nodes[i] = &mockNode{
			id:       fmt.Sprintf("node-%d", i),
			nodeType: "leaf",
			x:        (i % 50) * 20,
			y:        (i / 50) * 20,
			width:    10,
			height:   10,
		}
	}
	return nodes
}

// createRoot 创建一个包含子节点的根节点
func createRoot(children []layout.Node) layout.Node {
	return &mockNode{
		id:       "root",
		nodeType: "root",
		x:        0,
		y:        0,
		width:    1000,
		height:   1000,
		children: children,
	}
}

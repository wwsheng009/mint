package component_fixtures_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/examples/component_fixtures"
	rtlayout "github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestLayout_WithComponentFixtures 使用组件测试数据进行布局测试
func TestLayout_WithComponentFixtures(t *testing.T) {
	fixtures := component_fixtures.StandardFixtures()
	engine := rtlayout.NewEngine()
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			// 构建VNode
			vnode := fixture.Build()
			assert.NotNil(t, vnode, "VNode build should succeed")

			// 转换为layout.Node
			layoutNode := rtui.AsLayoutNode(vnode)
			assert.NotNil(t, layoutNode, "Layout node conversion should succeed")

			// 执行布局
			result := engine.Layout(layoutNode, constraints)
			assert.NotNil(t, result, "Layout result should not be nil")
			assert.NotNil(t, result.Root, "Root layout box should not be nil")

			// 验证布局结果
			t.Logf("Fixture: %s - Root box: %dx%d at (%d,%d)",
				fixture.Name,
				result.Root.Width,
				result.Root.Height,
				result.Root.X,
				result.Root.Y,
			)

			// 验证基本约束
			assert.GreaterOrEqual(t, result.Root.Width, constraints.MinWidth,
				"Width should be >= MinWidth")
			assert.LessOrEqual(t, result.Root.Width, constraints.MaxWidth,
				"Width should be <= MaxWidth")
			assert.GreaterOrEqual(t, result.Root.Height, constraints.MinHeight,
				"Height should be >= MinHeight")
			assert.LessOrEqual(t, result.Root.Height, constraints.MaxHeight,
				"Height should be <= MaxHeight")

			// 统计节点数
			nodeCount := component_fixtures.CountNodes(vnode)
			boxCount := len(result.Boxes)
			t.Logf("Node count: %d, Layout box count: %d", nodeCount, boxCount)
		})
	}
}

// TestLayout_Constraints 测试不同约束下的布局行为
func TestLayout_Constraints(t *testing.T) {
	tests := []struct {
		name        string
		fixtureName string
		constraints rtlayout.Constraints
	}{
		{
			name:        "unbounded",
			fixtureName: "simple_vstack",
			constraints: rtlayout.UnboundedConstraints(),
		},
		{
			name:        "tight_80x24",
			fixtureName: "simple_vstack",
			constraints: rtlayout.TightConstraints(80, 24),
		},
		{
			name:        "loose_10x10",
			fixtureName: "simple_vstack",
			constraints: rtlayout.LooseConstraints(10, 10),
		},
		{
			name:        "flex_layout_unbounded",
			fixtureName: "flex_layout",
			constraints: rtlayout.UnboundedConstraints(),
		},
		{
			name:        "flex_layout_constrained",
			fixtureName: "flex_layout",
			constraints: rtlayout.NewConstraints(0, 100, 0, 50),
		},
		{
			name:        "nested_layout",
			fixtureName: "nested_layout",
			constraints: rtlayout.NewConstraints(40, 80, 10, 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := component_fixtures.GetFixture(tt.fixtureName)
			assert.NotNil(t, fixture, "Fixture should exist")

			vnode := fixture.Build()
			layoutNode := rtui.AsLayoutNode(vnode)

			engine := rtlayout.NewEngine()
			result := engine.Layout(layoutNode, tt.constraints)

			assert.NotNil(t, result)
			assert.NotNil(t, result.Root)

			// 验证约束被遵守
			if tt.constraints.MinWidth > 0 {
				assert.GreaterOrEqual(t, result.Root.Width, tt.constraints.MinWidth)
			}
			if tt.constraints.MaxWidth < rtlayout.MaxInt {
				assert.LessOrEqual(t, result.Root.Width, tt.constraints.MaxWidth)
			}
			if tt.constraints.MinHeight > 0 {
				assert.GreaterOrEqual(t, result.Root.Height, tt.constraints.MinHeight)
			}
			if tt.constraints.MaxHeight < rtlayout.MaxInt {
				assert.LessOrEqual(t, result.Root.Height, tt.constraints.MaxHeight)
			}

			t.Logf("Layout result: %dx%d at (%d,%d)",
				result.Root.Width, result.Root.Height,
				result.Root.X, result.Root.Y,
			)
		})
	}
}

// TestLayout_Demo1FullApp 测试完整的Demo1应用布局
func TestLayout_Demo1FullApp(t *testing.T) {
	// 使用自定义配置
	vnode := component_fixtures.BuildDemo1App(
		component_fixtures.WithCount(42),
		component_fixtures.WithInput("test input"),
		component_fixtures.WithItems([]string{"A", "B", "C"}),
		component_fixtures.WithSize(120, 40),
	)

	layoutNode := rtui.AsLayoutNode(vnode)
	engine := rtlayout.NewEngine()
	constraints := rtlayout.NewConstraints(120, 120, 40, 40)

	result := engine.Layout(layoutNode, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)

	// 验证尺寸符合约束
	assert.Equal(t, 120, result.Root.Width, "Width should match constraint")
	assert.LessOrEqual(t, result.Root.Height, 40, "Height should be <= 40")

	t.Logf("Demo1 app layout: %dx%d, boxes: %d",
		result.Root.Width, result.Root.Height, len(result.Boxes))

	// 验证缓存统计
	stats := engine.GetStats()
	t.Logf("Cache stats - Hits: %d, Misses: %d",
		stats.CacheHits, stats.CacheMisses)
}

// TestLayout_Consistency 测试布局一致性
func TestLayout_Consistency(t *testing.T) {
	fixture := component_fixtures.GetFixture("demo1_full_app")
	assert.NotNil(t, fixture)

	vnode := fixture.Build()
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	engine := rtlayout.NewEngine()

	// 第一次布局
	result1 := engine.Layout(layoutNode, constraints)
	assert.NotNil(t, result1)

	// 第二次布局（应该命中缓存）
	result2 := engine.Layout(layoutNode, constraints)
	assert.NotNil(t, result2)

	// 验证结果一致
	assert.Equal(t, result1.Root.Width, result2.Root.Width, "Width should be consistent")
	assert.Equal(t, result1.Root.Height, result2.Root.Height, "Height should be consistent")

	stats := engine.GetStats()
	t.Logf("Cache stats after 2 layouts - Hits: %d, Misses: %d",
		stats.CacheHits, stats.CacheMisses)
}

// TestLayout_Invalidate 测试缓存失效
func TestLayout_Invalidate(t *testing.T) {
	fixture := component_fixtures.GetFixture("simple_vstack")
	assert.NotNil(t, fixture)

	vnode := fixture.Build()
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.UnboundedConstraints()

	engine := rtlayout.NewEngine()

	// 第一次布局
	result1 := engine.Layout(layoutNode, constraints)
	assert.NotNil(t, result1)

	// 使缓存失效
	engine.Invalidate()

	// 第二次布局（应该不命中缓存）
	result2 := engine.Layout(layoutNode, constraints)
	assert.NotNil(t, result2)

	// 验证结果一致
	assert.Equal(t, result1.Root.Width, result2.Root.Width)

	stats := engine.GetStats()
	// 应该有2次缓存未命中（第一次和失效后）
	assert.GreaterOrEqual(t, stats.CacheMisses, int64(1))
}

// TestLayout_BuildVNodeTree 测试动态构建的VNode树
func TestLayout_BuildVNodeTree(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		breadth  int
		minBoxes int
	}{
		{"shallow_wide", 2, 5, 25},   // 1 + 5 + 25 = 31
		{"deep_narrow", 4, 2, 8},     // 1 + 2 + 4 + 8 + 16 = 31
		{"balanced", 3, 3, 9},        // 1 + 3 + 9 + 27 = 40
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := component_fixtures.BuildVNodeTree(tt.depth, tt.breadth)
			layoutNode := rtui.AsLayoutNode(vnode)
			constraints := rtlayout.UnboundedConstraints()

			engine := rtlayout.NewEngine()
			result := engine.Layout(layoutNode, constraints)

			assert.NotNil(t, result)
			assert.NotNil(t, result.Root)
			assert.GreaterOrEqual(t, len(result.Boxes), tt.minBoxes,
				"Should have at least %d layout boxes", tt.minBoxes)

			t.Logf("Tree (depth=%d, breadth=%d): boxes=%d",
				tt.depth, tt.breadth, len(result.Boxes))
		})
	}
}

// TestLayout_BuildKeyedVNodeTree 测试带键的VNode树
func TestLayout_BuildKeyedVNodeTree(t *testing.T) {
	vnode := component_fixtures.BuildKeyedVNodeTree(2, 3, "root")
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	engine := rtlayout.NewEngine()
	result := engine.Layout(layoutNode, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)

	// 验证根节点有key
	if adapter, ok := layoutNode.(*rtui.VNodeAdapter); ok {
		id := adapter.ID()
		assert.Equal(t, "root", id, "Root should have key 'root'")
	}

	t.Logf("Keyed tree layout: boxes=%d", len(result.Boxes))
}

// TestLayout_FlexLayout 测试Flex布局
func TestLayout_FlexLayout(t *testing.T) {
	fixture := component_fixtures.GetFixture("flex_layout")
	assert.NotNil(t, fixture)

	vnode := fixture.Build()
	layoutNode := rtui.AsLayoutNode(vnode)

	tests := []struct {
		name        string
		constraints rtlayout.Constraints
	}{
		{
			name:        "unbounded",
			constraints: rtlayout.UnboundedConstraints(),
		},
		{
			name:        "constrained_width",
			constraints: rtlayout.NewConstraints(0, 100, 0, 50),
		},
		{
			name:        "tight_width",
			constraints: rtlayout.TightConstraints(80, 50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := rtlayout.NewEngine()
			result := engine.Layout(layoutNode, tt.constraints)

			assert.NotNil(t, result)
			assert.NotNil(t, result.Root)

			// 验证宽度约束
			if tt.constraints.MaxWidth < rtlayout.MaxInt {
				assert.LessOrEqual(t, result.Root.Width, tt.constraints.MaxWidth)
			}

			t.Logf("Flex layout %s: %dx%d",
				tt.name, result.Root.Width, result.Root.Height)
		})
	}
}

// TestLayout_NestedLayout 测试嵌套布局
func TestLayout_NestedLayout(t *testing.T) {
	fixture := component_fixtures.GetFixture("nested_layout")
	assert.NotNil(t, fixture)

	vnode := fixture.Build()
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.NewConstraints(40, 80, 10, 20)

	engine := rtlayout.NewEngine()
	result := engine.Layout(layoutNode, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)

	// 验证层级结构
	assert.Greater(t, len(result.Root.Children), 0, "Root should have children")

	// 递归验证所有box都有有效的尺寸
	validateBoxes(t, result.Root)

	t.Logf("Nested layout: root=%dx%d, total boxes=%d",
		result.Root.Width, result.Root.Height, len(result.Boxes))
}

// validateBoxes 递归验证所有box的尺寸有效
func validateBoxes(t *testing.T, box *rtlayout.LayoutBox) {
	t.Helper()

	assert.GreaterOrEqual(t, box.Width, 0, "Box width should be non-negative")
	assert.GreaterOrEqual(t, box.Height, 0, "Box height should be non-negative")

	for _, child := range box.Children {
		validateBoxes(t, child)
	}
}

// TestLayout_MixedKeyedTree 测试混合键树
func TestLayout_MixedKeyedTree(t *testing.T) {
	vnode := component_fixtures.BuildMixedKeyedTree(3, 2)
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.UnboundedConstraints()

	engine := rtlayout.NewEngine()
	result := engine.Layout(layoutNode, constraints)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)

	// 验证有足够的子节点（3 keyed + 2 non-keyed）
	assert.GreaterOrEqual(t, len(result.Root.Children), 5,
		"Should have at least 5 children")

	t.Logf("Mixed keyed tree: boxes=%d, children=%d",
		len(result.Boxes), len(result.Root.Children))
}

// BenchmarkLayout_Layout 性能基准测试
func BenchmarkLayout_Layout(b *testing.B) {
	fixtures := component_fixtures.StandardFixtures()
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	for _, fixture := range fixtures {
		b.Run(fixture.Name, func(b *testing.B) {
			vnode := fixture.Build()
			layoutNode := rtui.AsLayoutNode(vnode)
			engine := rtlayout.NewEngine()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.Layout(layoutNode, constraints)
			}
		})
	}
}

// BenchmarkLayout_Demo1App Demo1应用性能基准
func BenchmarkLayout_Demo1App(b *testing.B) {
	vnode := component_fixtures.BuildDemo1App(
		component_fixtures.WithCount(100),
		component_fixtures.WithItems([]string{
			"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
		}),
	)
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.NewConstraints(120, 120, 40, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := rtlayout.NewEngine()
		engine.Layout(layoutNode, constraints)
	}
}

// BenchmarkLayout_Caching 缓存性能基准
func BenchmarkLayout_Caching(b *testing.B) {
	fixture := component_fixtures.GetFixture("demo1_full_app")
	vnode := fixture.Build()
	layoutNode := rtui.AsLayoutNode(vnode)
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	engine := rtlayout.NewEngine()

	// 预热缓存
	engine.Layout(layoutNode, constraints)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(layoutNode, constraints)
	}
}

package compute

import (
	"testing"

	"github.com/wwsheng009/mint/internal/reconciler"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime"
)

// TestComplexLayoutValidation 测试复杂布局的验证
func TestComplexLayoutValidation(t *testing.T) {
	tests := []struct {
		name        string
		description string
		buildLayout func() (*ComputedLayout, error)
		expectValid bool
		checkItems   []string // 需要检查的验证项
	}{
		{
			name:        "DeeplyNestedLayout",
			description: "深度嵌套的布局结构",
			buildLayout: func() (*ComputedLayout, error) {
				// 构建一个5层深的嵌套布局
				leaf := rtui.Element("text").Prop("content", "Deep").Build()
				l4 := rtui.Element("bordered").Children(leaf).Build()
				l3 := rtui.Element("vstack").Children(l4).Build()
				l2 := rtui.Element("bordered").Children(l3).Build()
				l1 := rtui.Element("hstack").Children(l2).Build()
				root := rtui.Element("bordered").Children(l1).Build()

				fiber := reconciler.CreateFiberFromVNode(root)
				engine := NewEngine()
				constraints := runtime.BoxConstraints{
					MinWidth: 80, MaxWidth: 80,
					MinHeight: 24, MaxHeight: 24,
				}
				return engine.Layout(root, fiber, constraints)
			},
			expectValid: true,
			checkItems:  []string{"NodeCount", "UniqueNodeIDs", "ParentChildRelation"},
		},
		{
			name:        "WideHorizontalLayout",
			description: "宽水平布局（多个子节点）",
			buildLayout: func() (*ComputedLayout, error) {
				children := make([]rtui.VNode, 10)
				for i := 0; i < 10; i++ {
					children[i] = rtui.Element("text").
						Prop("content", "Item").
						Key(string(rune('A' + i))).
						Build()
				}
				hstack := rtui.Element("hstack").Children(children...).Build()
				root := rtui.Element("bordered").Children(hstack).Build()

				fiber := reconciler.CreateFiberFromVNode(root)
				engine := NewEngine()
				constraints := runtime.BoxConstraints{
					MinWidth: 100, MaxWidth: 100,
					MinHeight: 10, MaxHeight: 10,
				}
				return engine.Layout(root, fiber, constraints)
			},
			expectValid: true,
			checkItems:  []string{"NodeCount", "UniqueNodeIDs"},
		},
		{
			name:        "TallVerticalLayout",
			description: "高垂直布局（多个子节点）",
			buildLayout: func() (*ComputedLayout, error) {
				children := make([]rtui.VNode, 15)
				for i := 0; i < 15; i++ {
					children[i] = rtui.Element("text").
						Prop("content", "Line").
						Key(string(rune('A' + i))).
						Build()
				}
				vstack := rtui.Element("vstack").Children(children...).Build()
				root := rtui.Element("bordered").Children(vstack).Build()

				fiber := reconciler.CreateFiberFromVNode(root)
				engine := NewEngine()
				constraints := runtime.BoxConstraints{
					MinWidth: 40, MaxWidth: 40,
					MinHeight: 30, MaxHeight: 30,
				}
				return engine.Layout(root, fiber, constraints)
			},
			expectValid: true,
			checkItems:  []string{"NodeCount", "UniqueNodeIDs"},
		},
		{
			name:        "GridLayout",
			description: "网格布局（2x3）",
			buildLayout: func() (*ComputedLayout, error) {
				// 创建2行3列的网格
				rows := make([]rtui.VNode, 2)
				for r := 0; r < 2; r++ {
					cells := make([]rtui.VNode, 3)
					for c := 0; c < 3; c++ {
						cells[c] = rtui.Element("bordered").
							Children(
								rtui.Element("text").Prop("content", "Cell").Build(),
							).
							Key(string(rune('A'+r*3+c))).
							Build()
					}
					rows[r] = rtui.Element("hstack").Children(cells...).Build()
				}
				grid := rtui.Element("vstack").Children(rows...).Build()
				root := rtui.Element("bordered").Children(grid).Build()

				fiber := reconciler.CreateFiberFromVNode(root)
				engine := NewEngine()
				constraints := runtime.BoxConstraints{
					MinWidth: 80, MaxWidth: 80,
					MinHeight: 20, MaxHeight: 20,
				}
				return engine.Layout(root, fiber, constraints)
			},
			expectValid: true,
			checkItems:  []string{"NodeCount", "UniqueNodeIDs", "NoOverlaps"},
		},
		{
			name:        "MixedFlexLayout",
			description: "混合Flex布局（flex和固定尺寸混合）",
			buildLayout: func() (*ComputedLayout, error) {
				// 固定宽度 + flex子节点
				fixed := rtui.Element("bordered").
					Children(rtui.Element("text").Prop("content", "Fixed").Build()).
					Build()

				flex1 := rtui.Element("bordered").
					Children(rtui.Element("text").Prop("content", "Flex1").Build()).
					Build()

				flex2 := rtui.Element("bordered").
					Children(rtui.Element("text").Prop("content", "Flex2").Build()).
					Build()

				hstack := rtui.Element("hstack").Children(fixed, flex1, flex2).Build()
				root := rtui.Element("bordered").Children(hstack).Build()

				fiber := reconciler.CreateFiberFromVNode(root)
				engine := NewEngine()
				constraints := runtime.BoxConstraints{
					MinWidth: 80, MaxWidth: 80,
					MinHeight: 24, MaxHeight: 24,
				}
				return engine.Layout(root, fiber, constraints)
			},
			expectValid: true,
			checkItems:  []string{"NodeCount", "UniqueNodeIDs"},
		},
		{
			name:        "SidebarLayout",
			description: "侧边栏布局（类似IDE）",
			buildLayout: func() (*ComputedLayout, error) {
				// 左侧边栏
				sidebarItems := make([]rtui.VNode, 5)
				for i := 0; i < 5; i++ {
					sidebarItems[i] = rtui.Element("text").
						Prop("content", "Item").
						Build()
				}
				sidebar := rtui.Element("bordered").
					Children(rtui.Element("vstack").Children(sidebarItems...).Build()).
					Build()

				// 主内容区
				content := rtui.Element("bordered").
					Children(rtui.Element("text").Prop("content", "Content").Build()).
					Build()

				// 右侧面板
				panels := rtui.Element("bordered").
					Children(rtui.Element("text").Prop("content", "Panels").Build()).
					Build()

				main := rtui.Element("hstack").Children(sidebar, content, panels).Build()
				root := rtui.Element("bordered").Children(main).Build()

				fiber := reconciler.CreateFiberFromVNode(root)
				engine := NewEngine()
				constraints := runtime.BoxConstraints{
					MinWidth: 120, MaxWidth: 120,
					MinHeight: 40, MaxHeight: 40,
				}
				return engine.Layout(root, fiber, constraints)
			},
			expectValid: true,
			checkItems:  []string{"NodeCount", "UniqueNodeIDs", "NoOverlaps"},
		},
	}

	validator := NewLayoutValidator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			layout, err := tc.buildLayout()
			if err != nil {
				t.Fatalf("Failed to build layout: %v", err)
			}
			if layout == nil || layout.Root == nil {
				t.Fatal("Layout is nil")
			}

			// 验证布局
			result := validator.Validate(layout)

			t.Logf("=== %s ===", tc.description)
			t.Logf("Valid: %v", result.Valid)
			t.Logf("Total Issues: %d (Errors: %d, Warnings: %d)",
				len(result.Issues), result.ErrorCount(), result.WarningCount())

			// 打印问题摘要
			if len(result.Summary) > 0 {
				t.Logf("Issue Summary:")
				for issueType, count := range result.Summary {
					t.Logf("  %s: %d", issueType, count)
				}
			}

			// 检查预期
			if tc.expectValid && !result.Valid {
				// 如果预期有效但验证失败，打印详细信息
				for _, issue := range result.Issues {
					if issue.Severity == "error" {
						t.Errorf("Validation error: [%s] %s at %s", issue.Type, issue.Message, issue.Path)
					}
				}
			}

			// 统计节点
			nodeCount := countNodes(layout.Root)
			uniqueNodeIDs := countUniqueNodeIDs(layout.Root)
			t.Logf("Nodes: %d, Unique NodeIDs: %d", nodeCount, uniqueNodeIDs)

			// 验证节点ID唯一性
			if nodeCount != uniqueNodeIDs {
				t.Errorf("NodeID not unique: %d nodes but %d unique IDs", nodeCount, uniqueNodeIDs)
			}
		})
	}
}

// TestLayoutStressTest 布局压力测试
func TestLayoutStressTest(t *testing.T) {
	tests := []struct {
		name     string
		children int
		depth    int
	}{
		{"Wide_50", 50, 1},
		{"Deep_10", 2, 10},
		{"Balanced_20x5", 20, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := buildNestedLayout(tc.children, tc.depth)
			fiber := reconciler.CreateFiberFromVNode(root)

			engine := NewEngine()
			constraints := runtime.BoxConstraints{
				MinWidth: 100, MaxWidth: 100,
				MinHeight: 50, MaxHeight: 50,
			}

			layout, err := engine.Layout(root, fiber, constraints)
			if err != nil {
				t.Fatalf("Layout failed: %v", err)
			}

			// 验证
			validator := NewLayoutValidator()
			result := validator.Validate(layout)

			nodeCount := countNodes(layout.Root)
			uniqueNodeIDs := countUniqueNodeIDs(layout.Root)

			t.Logf("Children: %d, Depth: %d", tc.children, tc.depth)
			t.Logf("Total nodes: %d, Unique NodeIDs: %d", nodeCount, uniqueNodeIDs)
			t.Logf("Valid: %v, Issues: %d", result.Valid, len(result.Issues))

			// 节点ID必须唯一
			if nodeCount != uniqueNodeIDs {
				t.Errorf("NodeID duplication: %d nodes, %d unique IDs", nodeCount, uniqueNodeIDs)
			}
		})
	}
}

// buildNestedLayout 构建嵌套布局
func buildNestedLayout(children, depth int) rtui.VNode {
	if depth <= 0 {
		return rtui.Element("text").Prop("content", "X").Build()
	}

	childNodes := make([]rtui.VNode, children)
	for i := 0; i < children; i++ {
		childNodes[i] = buildNestedLayout(children/2, depth-1)
	}

	return rtui.Element("bordered").Children(childNodes...).Build()
}

// countNodes 统计节点数量
func countNodes(box *ComputedBox) int {
	if box == nil {
		return 0
	}
	count := 1
	for _, child := range box.Children {
		count += countNodes(child)
	}
	return count
}

// countUniqueNodeIDs 统计唯一的NodeID数量
func countUniqueNodeIDs(box *ComputedBox) int {
	ids := make(map[uint64]bool)
	var collect func(*ComputedBox)
	collect = func(b *ComputedBox) {
		if b == nil {
			return
		}
		ids[b.NodeID] = true
		for _, c := range b.Children {
			collect(c)
		}
	}
	collect(box)
	return len(ids)
}

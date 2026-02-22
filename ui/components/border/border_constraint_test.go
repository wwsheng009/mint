package border

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestBorder_ConstraintPropagation 测试约束传递逻辑
// 验证 computeChildConstraints 方法正确应用约束优先级规则
func TestBorder_ConstraintPropagation(t *testing.T) {
	tests := []struct {
		name                   string
		borderWidth            int
		borderHeight           int
		parentConstraints      layout.Constraints
		expectedChildConstraints layout.Constraints
	}{
		{
			name:     "Explicit width uses inner width (with border padding)",
			borderWidth: 20,
			borderHeight: 0,
			parentConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  50,
				MinHeight: 0,
				MaxHeight: 100,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  18,  // 20 - borderPadding(2)
				MaxWidth:  18,
				MinHeight: 0,
				MaxHeight: 98,  // 100 - borderPadding(2)
			},
		},
		{
			name:     "Auto width uses parent constraint (with border padding)",
			borderWidth: 0,  // auto
			borderHeight: 0,
			parentConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  50,
				MinHeight: 0,
				MaxHeight: 100,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  48,  // 50 - borderPadding(2)
				MinHeight: 0,
				MaxHeight: 98,  // 100 - borderPadding(2)
			},
		},
		{
			name:     "Explicit both dimensions",
			borderWidth: 30,
			borderHeight: 10,
			parentConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  100,
				MinHeight: 0,
				MaxHeight: 50,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  28,  // 30 - 2
				MaxWidth:  28,  // 30 - 2
				MinHeight: 8,   // 10 - 2
				MaxHeight: 8,   // 10 - 2
			},
		},
		{
			name:     "Explicit width exceeds parent MaxWidth (explicit takes priority)",
			borderWidth: 60,
			borderHeight: 0,
			parentConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  50,
				MinHeight: 0,
				MaxHeight: 100,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  58,  // 60 - 2 (explicit takes priority, even if exceeds parent)
				MaxWidth:  58,
				MinHeight: 0,
				MaxHeight: 98,
			},
		},
		{
			name:     "Parent constraint with MinWidth",
			borderWidth: 0,
			borderHeight: 0,
			parentConstraints: layout.Constraints{
				MinWidth:  20,
				MaxWidth:  50,
				MinHeight: 10,
				MaxHeight: 30,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  18,  // 20 - 2
				MaxWidth:  48,  // 50 - 2
				MinHeight: 8,   // 10 - 2
				MaxHeight: 28,  // 30 - 2
			},
		},
		{
			name:     "Tight parent constraint (equal min and max)",
			borderWidth: 0,
			borderHeight: 0,
			parentConstraints: layout.Constraints{
				MinWidth:  40,
				MaxWidth:  40,
				MinHeight: 20,
				MaxHeight: 20,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  38,  // 40 - 2
				MaxWidth:  38,
				MinHeight: 18,  // 20 - 2
				MaxHeight: 18,
			},
		},
		{
			name:     "Double border (still 1-char wide per border)",
			borderWidth: 0,
			borderHeight: 0,
			parentConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  50,
				MinHeight: 0,
				MaxHeight: 30,
			},
			expectedChildConstraints: layout.Constraints{
				MinWidth:  0,
				MaxWidth:  48,  // All border styles are 1-char wide
				MinHeight: 0,
				MaxHeight: 28,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			border := NewInstance(rtui.Props{
				"width":  tt.borderWidth,
				"height": tt.borderHeight,
			})

			childConstraints := border.computeChildConstraints(tt.parentConstraints)

			if childConstraints != tt.expectedChildConstraints {
				t.Errorf("Child constraints mismatch:\n  Expected: %+v\n  Got:      %+v",
					tt.expectedChildConstraints, childConstraints)
			}
		})
	}
}

// TestBorder_ConstraintPriority 测试约束优先级规则
// 规则：显式维度 > 父约束 > 自动测量
func TestBorder_ConstraintPriority(t *testing.T) {
	t.Run("Explicit width takes priority over parent constraint", func(t *testing.T) {
		border := NewInstance(rtui.Props{
			"width": 20,  // 显式宽度
		})

		parentConstraints := layout.Constraints{
			MinWidth:  40,
			MaxWidth:  60,
			MinHeight: 0,
			MaxHeight: 100,
		}

		childConstraints := border.computeChildConstraints(parentConstraints)

		// 显式宽度 20 应该优先，不受父约束 40-60 影响
		if childConstraints.MinWidth != 18 || childConstraints.MaxWidth != 18 {
			t.Errorf("Explicit width should take priority. Expected MinWidth=18, MaxWidth=18, got MinWidth=%d, MaxWidth=%d",
				childConstraints.MinWidth, childConstraints.MaxWidth)
		}
	})

	t.Run("Auto width uses parent constraint", func(t *testing.T) {
		border := NewInstance(rtui.Props{
			"width": 0,  // auto
		})

		parentConstraints := layout.Constraints{
			MinWidth:  10,
			MaxWidth:  50,
			MinHeight: 0,
			MaxHeight: 100,
		}

		childConstraints := border.computeChildConstraints(parentConstraints)

		// Auto 宽度应使用父约束
		if childConstraints.MinWidth != 8 || childConstraints.MaxWidth != 48 {
			t.Errorf("Auto width should use parent constraint. Expected MinWidth=8, MaxWidth=48, got MinWidth=%d, MaxWidth=%d",
				childConstraints.MinWidth, childConstraints.MaxWidth)
		}
	})

	t.Run("Parent constraint with explicit height", func(t *testing.T) {
		border := NewInstance(rtui.Props{
			"width":  0,
			"height": 15,  // 显式高度
		})

		parentConstraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  100,
			MinHeight: 5,
			MaxHeight: 30,
		}

		childConstraints := border.computeChildConstraints(parentConstraints)

		// 显式高度 15 应该优先
		if childConstraints.MinHeight != 13 || childConstraints.MaxHeight != 13 {
			t.Errorf("Explicit height should take priority. Expected MinHeight=13, MaxHeight=13, got MinHeight=%d, MaxHeight=%d",
				childConstraints.MinHeight, childConstraints.MaxHeight)
		}

		// 宽度应使用父约束 (auto)
		if childConstraints.MinWidth != 0 || childConstraints.MaxWidth != 98 {
			t.Errorf("Auto width should use parent constraint. Expected MinWidth=0, MaxWidth=98, got MinWidth=%d, MaxWidth=%d",
				childConstraints.MinWidth, childConstraints.MaxWidth)
		}
	})
}

// TestBorder_TracedConstraintPropagation 测试约束追踪
// 验证 TraceMeasuring 被正确调用
func TestBorder_TracedConstraintPropagation(t *testing.T) {
	// 启用追踪器
	layout.EnableTracer()
	defer layout.DisableTracer()

	// 清除之前的追踪数据
	layout.ClearTrace()

	// 创建 border 并测量
	child := &mockChildVNode{width: 20, height: 5}
	border := NewInstance(rtui.Props{
		"width":  0,  // auto，会触发子元素测量
		"child":  child,
		"key":    "test-border",
	})
	border.SetPath("/test")

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  50,
		MinHeight: 0,
		MaxHeight: 30,
	}

	border.Measure(constraints)

	// 检查追踪数据
	entries := layout.GetTraceEntries()
	if len(entries) == 0 {
		t.Fatal("Expected trace entries, got none")
	}

	// 查找相关的追踪条目
	found := false
	for _, entry := range entries {
		if entry.From == "border(test-border)" {
			found = true
			// 验证输入约束
			if entry.Input != constraints {
				t.Errorf("Input constraints mismatch. Expected %+v, got %+v", constraints, entry.Input)
			}
			// 验证输出约束（应该减去边框 padding）
			expectedOutput := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  48,  // 50 - 2
				MinHeight: 0,
				MaxHeight: 28,  // 30 - 2
			}
			if entry.Output != expectedOutput {
				t.Errorf("Output constraints mismatch. Expected %+v, got %+v", expectedOutput, entry.Output)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find trace entry from 'border(test-border)'")
	}
}

// TestBorder_BorderPaddingCalculation 测试边框内边距计算
func TestBorder_BorderPaddingCalculation(t *testing.T) {
	tests := []struct {
		name         string
		borderStyle  BorderStyle
		outerWidth   int
		innerWidth   int
	}{
		{"Single border", BorderSingle, 20, 18},
		{"Double border", BorderDouble, 20, 18},
		{"Rounded border", BorderRounded, 30, 28},
		{"Dashed border", BorderDashed, 40, 38},
		{"No border", BorderNone, 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			border := NewInstance(rtui.Props{
				"borderStyle": tt.borderStyle,
				"width":       tt.outerWidth,
			})

			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  tt.outerWidth,
				MinHeight: 0,
				MaxHeight: 100,
			}

			childConstraints := border.computeChildConstraints(constraints)

			borderWidth := GetBorderWidth(tt.borderStyle)
			expectedInnerWidth := tt.outerWidth - 2*borderWidth

			if childConstraints.MaxWidth != expectedInnerWidth {
				t.Errorf("Inner width mismatch. BorderStyle=%s, Outer=%d, Expected inner=%d, got %d",
					tt.borderStyle, tt.outerWidth, expectedInnerWidth, childConstraints.MaxWidth)
			}
		})
	}
}

package grid

import (
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Cell Borders 属性测试
// =============================================================================

func TestVNode_CellBordersDefaults(t *testing.T) {
	g := New()

	if g.showCellBorders != false {
		t.Errorf("预期的 showCellBorders 默认为 false, 得到 %v", g.showCellBorders)
	}
	if g.cellBorderStyle != "single" {
		t.Errorf("预期的 cellBorderStyle 默认为 'single', 得到 %v", g.cellBorderStyle)
	}
	if g.cellBorderRounded != false {
		t.Errorf("预期的 cellBorderRounded 默认为 false, 得到 %v", g.cellBorderRounded)
	}
	if g.cellBorderColor != "" {
		t.Errorf("预期的 cellBorderColor 默认为 '', 得到 %v", g.cellBorderColor)
	}
}

func TestVNode_SetCellBorderProperties(t *testing.T) {
	g := New().
		SetShowCellBorders(true).
		SetCellBorderStyle("double").
		SetCellBorderRounded(true).
		SetCellBorderColor("cyan")

	if !g.showCellBorders {
		t.Error("SetShowCellBorders(true) 失败")
	}
	if g.cellBorderStyle != "double" {
		t.Errorf("SetCellBorderStyle('double') 失败, 得到 %v", g.cellBorderStyle)
	}
	if !g.cellBorderRounded {
		t.Error("SetCellBorderRounded(true) 失败")
	}
	if g.cellBorderColor != "cyan" {
		t.Errorf("SetCellBorderColor('cyan') 失败, 得到 %v", g.cellBorderColor)
	}
}

func TestVNode_ShowCellBorders(t *testing.T) {
	g := New().ShowCellBorders()

	if !g.showCellBorders {
		t.Error("ShowCellBorders() 没有启用 cell 边框")
	}
	if g.cellBorderStyle != "single" {
		t.Errorf("ShowCellBorders() 应该使用 'single' 样式, 得到 %v", g.cellBorderStyle)
	}
}

func TestVNode_HideCellBorders(t *testing.T) {
	g := New().ShowCellBorders().HideCellBorders()

	if g.showCellBorders {
		t.Error("HideCellBorders() 没有隐藏 cell 边框")
	}
}

func TestVNode_SingleCellBorders(t *testing.T) {
	g := New().SingleCellBorders()

	if !g.showCellBorders {
		t.Error("SingleCellBorders() 没有启用 cell 边框")
	}
	if g.cellBorderStyle != "single" {
		t.Errorf("SingleCellBorders() 应该使用 'single' 样式, 得到 %v", g.cellBorderStyle)
	}
}

func TestVNode_DoubleCellBorders(t *testing.T) {
	g := New().DoubleCellBorders()

	if !g.showCellBorders {
		t.Error("DoubleCellBorders() 没有启用 cell 边框")
	}
	if g.cellBorderStyle != "double" {
		t.Errorf("DoubleCellBorders() 应该使用 'double' 样式, 得到 %v", g.cellBorderStyle)
	}
}

func TestVNode_LightCellBorders(t *testing.T) {
	g := New().LightCellBorders()

	if !g.showCellBorders {
		t.Error("LightCellBorders() 没有启用 cell 边框")
	}
	if g.cellBorderStyle != "light" {
		t.Errorf("LightCellBorders() 应该使用 'light' 样式, 得到 %v", g.cellBorderStyle)
	}
}

func TestVNode_RoundedCellBorders(t *testing.T) {
	g := New().RoundedCellBorders()

	if !g.showCellBorders {
		t.Error("RoundedCellBorders() 没有启用 cell 边框")
	}
	if !g.cellBorderRounded {
		t.Error("RoundedCellBorders() 没有启用圆角")
	}
}

// =============================================================================
// Props 测试
// =============================================================================

func TestVNode_CellBordersProps(t *testing.T) {
	g := New()
	props := rtui.Props{
		"showCellBorders":   true,
		"cellBorderStyle":   "double",
		"cellBorderRounded": false,
		"cellBorderColor":   "red",
	}

	g.SetProps(props)

	if g.showCellBorders != true {
		t.Error("Props showCellBorders 未正确设置")
	}
	if g.cellBorderStyle != "double" {
		t.Error("Props cellBorderStyle 未正确设置")
	}
	if g.cellBorderRounded != false {
		t.Error("Props cellBorderRounded 未正确设置")
	}
	if g.cellBorderColor != "red" {
		t.Error("Props cellBorderColor 未正确设置")
	}
}

// =============================================================================
// Instance Props 测试
// =============================================================================

func TestInstance_CellBordersProps(t *testing.T) {
	props := rtui.Props{
		"columns":           []Dimension{Flex{Factor: 1}, Flex{Factor: 1}},
		"rows":              []Dimension{Flex{Factor: 1}, Flex{Factor: 1}},
		"showCellBorders":   true,
		"cellBorderStyle":   "double",
		"cellBorderRounded": false,
		"cellBorderColor":   "blue",
	}

	inst := NewInstance(props)

	if inst.showCellBorders != true {
		t.Error("Instance: showCellBars 未正确设置")
	}
	if inst.cellBorderStyle != "double" {
		t.Error("Instance: cellBorderStyle 未正确设置")
	}
	if inst.cellBorderRounded != false {
		t.Error("Instance: cellBorderRounded 未正确设置")
	}
	if inst.cellBorderColor != "blue" {
		t.Error("Instance: cellBorderColor 未正确设置")
	}
}

// =============================================================================
// Measure 集成测试
// =============================================================================

func TestInstance_CellBordersMeasureSize(t *testing.T) {
	props := rtui.Props{
		"columns":         []Dimension{Flex{Factor: 1}, Flex{Factor: 1}},
		"rows":            []Dimension{Flex{Factor: 1}, Flex{Factor: 1}},
		"showCellBorders": false,
		"width":           20,
		"height":          10,
	}

	inst := NewInstance(props)

	// 创建 mock 对象来测试 Measure
	// 这里我们测试基本属性是否正确设置
	if inst.showCellBorders {
		t.Error("初始状态 showCellBorders 应该为 false")
	}

	// 测试 SetProps 是否更新 cell borders
	newProps := rtui.Props{
		"showCellBorders": true,
	}
	inst.SetProps(newProps)

	if !inst.showCellBorders {
		t.Error("SetProps 应该更新 showCellBorders")
	}
}

// =============================================================================
// 边框字符测试
// =============================================================================

func TestCellBorderChars(t *testing.T) {
	tests := []struct {
		style     string
		wantChars BorderChars
	}{
		{"single", cellBorderChars["single"]},
		{"double", cellBorderChars["double"]},
		{"light", cellBorderChars["light"]},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			chars := cellBorderChars[tt.style]

			if chars.horizontal != tt.wantChars.horizontal {
				t.Errorf("字符不匹配")
			}

			// 检查所有必需的字符是否存在
			if chars.horizontal == "" || chars.vertical == "" ||
				chars.topLeft == "" || chars.topRight == "" ||
				chars.bottomLeft == "" || chars.bottomRight == "" {
				t.Errorf("边框字符集 %v 缺少必需的字符", tt.style)
			}
		})
	}
}

func TestRoundedBorderChars(t *testing.T) {
	chars := roundedBorderChars

	if chars.topLeft != "╭" {
		t.Errorf("圆角边框 topLeft 字符应该是 '╭', 得到 '%s'", chars.topLeft)
	}
	if chars.topRight != "╮" {
		t.Errorf("圆角边框 topRight 字符应该是 '╮', 得到 '%s'", chars.topRight)
	}
	if chars.bottomLeft != "╰" {
		t.Errorf("圆角边框 bottomLeft 字符应该是 '╰', 得到 '%s'", chars.bottomLeft)
	}
	if chars.bottomRight != "╯" {
		t.Errorf("圆角边框 bottomRight 字符应该是 '╯', 得到 '%s'", chars.bottomRight)
	}
}

// =============================================================================
// 链式调用测试
// =============================================================================

func TestVNode_CellBordersChaining(t *testing.T) {
	g := New().
		SetColumns(Flex{Factor: 1}, Flex{Factor: 1}).
		SetRows(Flex{Factor: 1}, Flex{Factor: 1}).
		SingleCellBorders().
		SetGap(1, 1).
		SetChildrenAuto([]rtui.VNode{
			text.New("A1"),
			text.New("A2"),
			text.New("B1"),
			text.New("B2"),
		})

	// 验证所有调用都正确执行
	if len(g.columns) != 2 {
		t.Error("链式调用中 SetColumns 失败")
	}
	if len(g.rows) != 2 {
		t.Error("链式调用中 SetRows 失败")
	}
	if !g.showCellBorders {
		t.Error("链式调用中 SingleCellBorders 失败")
	}
	if g.columnGap != 1 || g.rowGap != 1 {
		t.Error("链式调用中 SetGap 失败")
	}
	if len(g.cells) != 4 {
		t.Error("链式调用中 SetChildrenAuto 失败")
	}
}

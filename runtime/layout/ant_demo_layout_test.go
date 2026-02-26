package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAntDesignSimulation 模拟 ant_design_demo 中的嵌套布局
func TestAntDesignSimulation(t *testing.T) {
	// 模拟 FormItem 的内部结构：
	// FormItem(VStack)
	//   ├── HStack (label + input)  // 使用 HStack 作为 HStack 的子节点
	//   │     ├── Label (Text)
	//   │     └── Input (InputBox)
	//   └── ErrorMsg (Text)

	// 创建叶子节点（实际内容节点）
	label := NewMockMeasurableNode("label", 15, 1)
	input := NewMockMeasurableNode("input", 30, 1)
	errorMsg := NewMockMeasurableNode("error", 45, 1)

	// 创建 HStack (label + input)
	hStack := NewFlexLayout("hstack", []Node{label, input})
	hStack.SetDirection(FlexRow)
	hStack.SetGap(2)

	// 创建 FormItem (VStack 包含 hStack + errorMessage)
	formItem := NewFlexLayout("form_item", []Node{hStack, errorMsg})
	formItem.SetDirection(FlexColumn)
	formItem.SetGap(1)

	// 创建 FormContent (VStack 包含多个 FormItem)
	formContent := NewFlexLayout("form_content", []Node{formItem})
	formContent.SetDirection(FlexColumn)
	formContent.SetGap(2)

	// 布局
	engine := NewEngine()
	constraints := Constraints{
		MinWidth:  100,
		MaxWidth:  100,
		MinHeight: 50,
		MaxHeight: 50,
	}
	result := engine.Layout(formContent, constraints)
	box := result.Root

	// 验证结构
	assert.NotNil(t, box, "Root box should not be nil")
	assert.Len(t, box.Children, 1, "FormContent should have 1 FormItem")

	formItemBox := box.Children[0]
	assert.Len(t, formItemBox.Children, 2, "FormItem should have HStack and ErrorMsg")

	hStackBox := formItemBox.Children[0]
	assert.Len(t, hStackBox.Children, 2, "HStack should have Label and Input")

	errorMsgBox := formItemBox.Children[1]

	// 验证尺寸
	t.Logf("=== 层级尺寸 ===")
	t.Logf("FormContent: Width=%d, Height=%d", box.Width, box.Height)
	t.Logf("  ├─ FormItem: Width=%d, Height=%d", formItemBox.Width, formItemBox.Height)
	t.Logf("  │  ├─ HStack: Width=%d, Height=%d", hStackBox.Width, hStackBox.Height)
	t.Logf("  │  │  ├─ Label: Width=%d, Height=%d", hStackBox.Children[0].Width, hStackBox.Children[0].Height)
	t.Logf("  │  │  └─ Input: Width=%d, Height=%d", hStackBox.Children[1].Width, hStackBox.Children[1].Height)
	t.Logf("  │  └─ ErrorMsg: Width=%d, Height=%d", errorMsgBox.Width, errorMsgBox.Height)
}

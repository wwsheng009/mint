package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/stretchr/testify/assert"
)

// TestAntDesignFormItemLayout 模拟 FormItem 的布局结构
// FormItem = VStack (gap=1)
//   ├── HStack (label + input + required)
//   └── HStack (indent + helpText)
func TestAntDesignFormItemLayout(t *testing.T) {
	// 创建叶子节点
	label := NewMockMeasurableNode("label", 10, 1)   // "Username: "
	input := NewMockMeasurableNode("input", 24, 1)   // Input box
	required := NewMockMeasurableNode("*", 1, 1)     // Required mark

	// 第一行：HStack (label + input + required)
	labelRow := layout.NewFlexLayout("label_row", []layout.Node{label, input, required})
	labelRow.SetDirection(layout.FlexRow)
	labelRow.SetGap(1)

	// 创建帮助文本和缩进
	indent := NewMockMeasurableNode("indent", 11, 1) // 10 spaces + 1 for alignment
	helpText := NewMockMeasurableNode("help", 50, 1) // "We'll never share your email"

	// 第二行：HStack (indent + helpText)
	helpRow := layout.NewFlexLayout("help_row", []layout.Node{indent, helpText})
	helpRow.SetDirection(layout.FlexRow)
	helpRow.SetGap(0)

	// FormItem = VStack (gap=1)
	formItem := layout.NewFlexLayout("form_item", []layout.Node{labelRow, helpRow})
	formItem.SetDirection(layout.FlexColumn)
	formItem.SetGap(1) // 关键：这是验证 Gap 是否生效的地方

	// 布局
	engine := layout.NewEngine()
	constraints := layout.Constraints{
		MinWidth:  70,
		MaxWidth:  70,
		MinHeight: 10,
		MaxHeight: 10,
	}
	result := engine.Layout(formItem, constraints)

	// 打印布局树
	t.Logf("\n========== FormItem 布局树 ==========")
	printLayoutTree(result.Root, 0, t)

	// 验证 FormItem 结构
	box := result.Root
	assert.Equal(t, 2, len(box.Children), "FormItem 应该有 2 个子节点 (labelRow + helpRow)")

	labelRowBox := box.Children[0]
	helpRowBox := box.Children[1]

	t.Logf("\n========== 布局位置验证 ==========")
	t.Logf("LabelRow:  Y=%d, Height=%d, EndY=%d", labelRowBox.Y, labelRowBox.Height, labelRowBox.Y+labelRowBox.Height)
	t.Logf("HelpRow:   Y=%d, Height=%d, EndY=%d", helpRowBox.Y, helpRowBox.Height, helpRowBox.Y+helpRowBox.Height)

	// 验证 Gap 是否生效
	// Gap(1) 意味着 helpRow.Y 应该 >= labelRow.Y + labelRowBox.Height + 1
	expectedMinY := labelRowBox.Y + labelRowBox.Height + 1
	if helpRowBox.Y < expectedMinY {
		t.Errorf("✗ Gap(1) 未生效! HelpRow Y=%d, 应该 >= %d", helpRowBox.Y, expectedMinY)
		t.Errorf("  这说明两行会重叠！")
	} else {
		actualGap := helpRowBox.Y - (labelRowBox.Y + labelRowBox.Height)
		t.Logf("✓ Gap 生效! 实际 Gap = %d", actualGap)
		assert.Equal(t, 1, actualGap, "Gap 应该是 1")
	}

	// 验证尺寸
	t.Logf("\n========== 尺寸验证 ==========")
	t.Logf("FormItem 总尺寸: W=%d, H=%d", box.Width, box.Height)
	t.Logf("LabelRow:  W=%d, H=%d (包含: label, input, required)", labelRowBox.Width, labelRowBox.Height)
	t.Logf("HelpRow:   W=%d, H=%d (包含: indent, helpText)", helpRowBox.Width, helpRowBox.Height)

	// 验证内容不重叠
	overlap := (helpRowBox.Y < labelRowBox.Y+labelRowBox.Height)
	if overlap {
		t.Errorf("✗ 检测到内容重叠!")
	} else {
		t.Logf("✓ 内容无重叠")
	}
}

// TestAntDesignFormLayout 模拟整个 Form 的布局
func TestAntDesignFormLayout(t *testing.T) {
	// 创建 3 个 FormItem
	formItem1 := createMockFormItem("username", "Username:", "Enter your username", 24, "We'll never share your email")
	formItem2 := createMockFormItem("email", "Email:", "example@domain.com", 24, "")
	formItem3 := createMockFormItem("password", "Password:", "Enter your password", 24, "At least 8 characters")

	// Form = VStack (gap=2) 包含多个 FormItem
	form := layout.NewFlexLayout("form", []layout.Node{formItem1, formItem2, formItem3})
	form.SetDirection(layout.FlexColumn)
	form.SetGap(2)

	// 布局
	engine := layout.NewEngine()
	constraints := layout.Constraints{
		MinWidth:  80,
		MaxWidth:  80,
		MinHeight: 25,
		MaxHeight: 25,
	}
	result := engine.Layout(form, constraints)

	// 打印完整布局树
	t.Logf("\n========== 完整 Form 布局树 ==========")
	printLayoutTree(result.Root, 0, t)

	// 验证 Form 结构
	formBox := result.Root

	t.Logf("\n========== Form 层级验证 ==========")
	t.Logf("Form: W=%d, H=%d", formBox.Width, formBox.Height)
	t.Logf("FormItem 数量: %d", len(formBox.Children))

	// 验证每个 FormItem 之间没有重叠
	t.Logf("\n========== FormItem 位置验证 ==========")
	for i, itemBox := range formBox.Children {
		t.Logf("FormItem %d: Y=%d, Height=%d, EndY=%d", i+1, itemBox.Y, itemBox.Height, itemBox.Y+itemBox.Height)

		if i > 0 {
			prevItem := formBox.Children[i-1]
			prevEndY := prevItem.Y + prevItem.Height
			if itemBox.Y < prevEndY {
				t.Errorf("✗ FormItem %d 与 %d 重叠!", i, i+1)
				t.Errorf("  FormItem %d: EndY=%d", i, prevEndY)
				t.Errorf("  FormItem %d: Y=%d", i+1, itemBox.Y)
			} else {
				gap := itemBox.Y - prevEndY
				t.Logf("  与上一个 FormItem 的 Gap: %d", gap)
			}
		}

		// 验证 FormItem 内部布局
		if len(itemBox.Children) >= 2 {
			labelRow := itemBox.Children[0]
			helpRow := itemBox.Children[1]
			prevEndY := labelRow.Y + labelRow.Height
			if helpRow.Y < prevEndY {
				t.Errorf("✗ FormItem %d 内部行重叠!", i+1)
				t.Errorf("  LabelRow: Y=%d, H=%d, EndY=%d", labelRow.Y, labelRow.Height, prevEndY)
				t.Errorf("  HelpRow: Y=%d", helpRow.Y)
			}
		}
	}
}

// createMockFormItem 创建一个模拟的 FormItem
func createMockFormItem(field, label, placeholder string, inputWidth int, helpText string) layout.Node {
	// 模拟子节点
	labelNode := NewMockMeasurableNode(label, 10, 1)
	inputNode := NewMockMeasurableNode("input", inputWidth, 1)

	// Label 行
	labelRow := layout.NewFlexLayout("label_row", []layout.Node{labelNode, inputNode})
	labelRow.SetDirection(layout.FlexRow)
	labelRow.SetGap(1)

	// Help 行（如果有帮助文本）
	helpNode := NewMockMeasurableNode(helpText, len(helpText), 1)
	indentNode := NewMockMeasurableNode("indent", 11, 1)

	helpRow := layout.NewFlexLayout("help_row", []layout.Node{indentNode, helpNode})
	helpRow.SetDirection(layout.FlexRow)
	helpRow.SetGap(0)

	// FormItem = VStack (gap=1)
	children := []layout.Node{labelRow}
	if helpText != "" {
		children = append(children, helpRow)
	}
	formItem := layout.NewFlexLayout(fmt.Sprintf("form_item_%s", field), children)
	formItem.SetDirection(layout.FlexColumn)
	formItem.SetGap(1)

	return formItem
}

// NewMockMeasurableNode 创建一个简单的可测量节点
type MockMeasurableNode struct {
	id       string
	baseW    int
	baseH    int
}

func NewMockMeasurableNode(id string, width, height int) layout.Node {
	return &MockMeasurableNode{
		id:    id,
		baseW: width,
		baseH: height,
	}
}

func (n *MockMeasurableNode) ID() string     { return n.id }
func (n *MockMeasurableNode) Type() string   { return "MockMeasurable" }
func (n *MockMeasurableNode) Children() []layout.Node { return nil }
func (n *MockMeasurableNode) GetPosition() (int, int) { return 0, 0 }
func (n *MockMeasurableNode) SetPosition(x, y int) {}
func (n *MockMeasurableNode) SetSize(w, h int)    {}
func (n *MockMeasurableNode) GetSize() (int, int) { return n.baseW, n.baseH }
func (n *MockMeasurableNode) GetWidth() int      { return n.baseW }
func (n *MockMeasurableNode) GetHeight() int     { return n.baseH }
func (n *MockMeasurableNode) Measure(c layout.Constraints) layout.Size {
	w, h := n.baseW, n.baseH
	if w > c.MaxWidth {
		w = c.MaxWidth
	}
	if h > c.MaxHeight {
		h = c.MaxHeight
	}
	return layout.Size{Width: w, Height: h}
}

// printLayoutTree 打印布局树
func printLayoutTree(box *layout.LayoutBox, indent int, t *testing.T) {
	if box == nil {
		return
	}

	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	t.Logf("%s[%s] W=%d H=%d X=%d Y=%d Children=%d",
		prefix, box.ID, box.Width, box.Height, box.X, box.Y, len(box.Children))

	for _, child := range box.Children {
		printLayoutTree(child, indent+1, t)
	}
}

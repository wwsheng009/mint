package render

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type mockCustomPaintNode struct {
	id    string
	tag   string
	style style.Style
	paint func(x, y int) []paint.DrawCmd
}

func (m *mockCustomPaintNode) ID() string               { return m.id }
func (m *mockCustomPaintNode) NodeType() paint.NodeType { return paint.NodeTypeElement }
func (m *mockCustomPaintNode) Tag() string              { return m.tag }
func (m *mockCustomPaintNode) Style() style.Style       { return m.style }
func (m *mockCustomPaintNode) SetStyle(s style.Style)   { m.style = s }
func (m *mockCustomPaintNode) TextContent() string      { return "" }
func (m *mockCustomPaintNode) Paint(x, y int) []paint.DrawCmd {
	if m.paint == nil {
		return nil
	}
	return m.paint(x, y)
}

// =============================================================================
// PaintEngine New API Tests
// =============================================================================

func TestPaintEngine_PaintPaintableLayouts(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建 PaintableLayouts
	layouts := make(paint.PaintableLayouts)

	// 创建一个简单的 root box
	rootBox := paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layouts[paint.RenderLayerBase] = paint.NewPaintableLayout(rootBox)

	// 调用新 API
	err := engine.PaintPaintableLayouts(layouts, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts() error = %v", err)
	}
}

func TestPaintEngine_PaintPaintableLayouts_Nil(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 测试 nil 布局
	err := engine.PaintPaintableLayouts(nil, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts(nil) error = %v", err)
	}

	// 测试空布局
	layouts := make(paint.PaintableLayouts)
	err = engine.PaintPaintableLayouts(layouts, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts(empty) error = %v", err)
	}
}

func TestPaintEngine_PaintPaintablePlanes(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建 PaintablePlanes
	planes := paint.NewPaintablePlanes()
	planes.AddToLayer(paint.RenderLayerBase, paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25))
	planes.AddToLayer(paint.RenderLayerModal, paint.NewPaintableBoxWithBounds(nil, 20, 10, 40, 5))

	// 调用新 API
	err := engine.PaintPaintablePlanes(planes, buffer)
	if err != nil {
		t.Fatalf("PaintPaintablePlanes() error = %v", err)
	}
}

func TestPaintEngine_PaintPaintablePlanes_RepaintsZeroSizedTooltipLayerNode(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(40, 10)

	rootNode := &mockCustomPaintNode{id: "root", tag: "root"}
	tooltipNode := &mockCustomPaintNode{
		id:  "tooltip",
		tag: "tooltip",
		paint: func(x, y int) []paint.DrawCmd {
			return []paint.DrawCmd{{X: 8, Y: 2, Text: "TIP", Style: style.NewStyle().Foreground(style.Yellow)}}
		},
	}
	bodyNode := NewMockPaintableTextNode("body", "body text")

	rootBox := paint.NewPaintableBoxWithBounds(rootNode, 0, 0, 40, 10)
	tooltipBox := paint.NewPaintableBoxWithBounds(tooltipNode, 0, 0, 0, 0)
	bodyBox := paint.NewPaintableBoxWithBounds(bodyNode, 0, 2, 20, 1)
	rootBox.Children = []*paint.PaintableBox{tooltipBox, bodyBox}
	tooltipBox.Parent = rootBox
	bodyBox.Parent = rootBox

	planes := paint.NewPaintablePlanes()
	planes.AddToLayer(paint.RenderLayerBase, rootBox)
	planes.AddToLayer(paint.RenderLayerTooltip, tooltipBox)

	if err := engine.PaintPaintablePlanes(planes, buffer); err != nil {
		t.Fatalf("PaintPaintablePlanes() error = %v", err)
	}

	got := buffer.GetContent(8, 2).Cluster
	if got != "T" {
		t.Fatalf("tooltip cell = %q, want %q (tooltip layer should repaint above base text)", got, "T")
	}
}

func TestPaintEngine_PaintPaintablePlanes_Nil(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 测试 nil planes
	err := engine.PaintPaintablePlanes(nil, buffer)
	if err != nil {
		t.Fatalf("PaintPaintablePlanes(nil) error = %v", err)
	}

	// 测试空 planes
	planes := paint.NewPaintablePlanes()
	err = engine.PaintPaintablePlanes(planes, buffer)
	if err != nil {
		t.Fatalf("PaintPaintablePlanes(empty) error = %v", err)
	}
}

func TestPaintEngine_PaintPaintableLayouts_MultiLayer(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建多层 PaintableLayouts
	layouts := make(paint.PaintableLayouts)

	// Base layer
	baseBox := paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layouts[paint.RenderLayerBase] = paint.NewPaintableLayout(baseBox)

	// Modal layer
	modalBox := paint.NewPaintableBoxWithBounds(nil, 20, 10, 40, 5)
	layouts[paint.RenderLayerModal] = paint.NewPaintableLayout(modalBox)

	// Tooltip layer
	tooltipBox := paint.NewPaintableBoxWithBounds(nil, 30, 5, 20, 3)
	layouts[paint.RenderLayerTooltip] = paint.NewPaintableLayout(tooltipBox)

	// 调用新 API
	err := engine.PaintPaintableLayouts(layouts, buffer)
	if err != nil {
		t.Fatalf("PaintPaintableLayouts() error = %v", err)
	}
}

func TestPaintEngine_PaintLayout(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 创建 PaintableLayout
	rootBox := paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25)
	layout := paint.NewPaintableLayout(rootBox)

	// 调用 PaintLayout
	err := engine.PaintLayout(layout, buffer)
	if err != nil {
		t.Fatalf("PaintLayout() error = %v", err)
	}
}

func TestPaintEngine_PaintLayout_Nil(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 测试 nil layout
	err := engine.PaintLayout(nil, buffer)
	if err != nil {
		t.Fatalf("PaintLayout(nil) error = %v", err)
	}

	// 测试空 layout
	layout := &paint.PaintableLayout{}
	err = engine.PaintLayout(layout, buffer)
	if err != nil {
		t.Fatalf("PaintLayout(empty) error = %v", err)
	}
}

// =============================================================================
// Modal Backdrop Tests
// =============================================================================

func TestPaintEngine_ModalBackdrop_Tracking(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(80, 25)

	// 第一次渲染：无 modal
	layoutsNoModal := make(paint.PaintableLayouts)
	layoutsNoModal[paint.RenderLayerBase] = paint.NewPaintableLayout(
		paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25),
	)
	engine.PaintPaintableLayouts(layoutsNoModal, buffer)

	// 第二次渲染：有 modal
	layoutsWithModal := make(paint.PaintableLayouts)
	layoutsWithModal[paint.RenderLayerBase] = paint.NewPaintableLayout(
		paint.NewPaintableBoxWithBounds(nil, 0, 0, 80, 25),
	)
	// 有效的 modal box 需要有效的 Node
	modalBox := paint.NewPaintableBoxWithBounds(nil, 20, 10, 40, 5)
	layoutsWithModal[paint.RenderLayerModal] = paint.NewPaintableLayout(modalBox)
	engine.PaintPaintableLayouts(layoutsWithModal, buffer)

	// 第三次渲染：modal 消失
	engine.PaintPaintableLayouts(layoutsNoModal, buffer)
}

func TestPaintEngine_PaintPaintablePlanes_IgnoresPortalHostForBackdrop(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(20, 8)

	baseNode := &mockCustomPaintNode{
		id:  "base",
		tag: "base",
		paint: func(x, y int) []paint.DrawCmd {
			return []paint.DrawCmd{{
				X:     5,
				Y:     5,
				Text:  "B",
				Style: style.NewStyle().Foreground(style.White),
			}}
		},
	}
	baseBox := paint.NewPaintableBoxWithBounds(baseNode, 0, 0, 20, 8)
	baseBox.Layer = int(paint.RenderLayerBase)

	hostVNode := rtui.NewElement("box").
		SetPortalRootId(rtui.DefaultModalPortalRootID).
		SetLayer(rtui.LayerModal)
	hostBox := paint.NewPaintableBoxWithBounds(NewVNodePaintableNode(hostVNode), 0, 0, 1, 1)
	hostBox.Layer = int(paint.RenderLayerModal)

	planes := paint.NewPaintablePlanes()
	planes.AddToLayer(paint.RenderLayerBase, baseBox)
	planes.AddToLayer(paint.RenderLayerModal, hostBox)

	if err := engine.PaintPaintablePlanes(planes, buffer); err != nil {
		t.Fatalf("PaintPaintablePlanes() error = %v", err)
	}

	cell := buffer.GetContent(5, 5)
	if cell.Style.FG != style.White {
		t.Fatalf("base cell FG = %q, want %q (portal host should not trigger backdrop)", cell.Style.FG, style.White)
	}
}

func TestPaintEngine_PaintPaintablePlanes_UsesVisibleModalBoxForBackdrop(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(30, 10)

	baseNode := &mockCustomPaintNode{
		id:  "base",
		tag: "base",
		paint: func(x, y int) []paint.DrawCmd {
			return []paint.DrawCmd{{
				X:     2,
				Y:     8,
				Text:  "B",
				Style: style.NewStyle().Foreground(style.White),
			}}
		},
	}
	baseBox := paint.NewPaintableBoxWithBounds(baseNode, 0, 0, 30, 10)
	baseBox.Layer = int(paint.RenderLayerBase)

	hostVNode := rtui.NewElement("box").
		SetPortalRootId(rtui.DefaultModalPortalRootID).
		SetLayer(rtui.LayerModal)
	hostBox := paint.NewPaintableBoxWithBounds(NewVNodePaintableNode(hostVNode), 0, 0, 1, 1)
	hostBox.Layer = int(paint.RenderLayerModal)

	dialogNode := &mockCustomPaintNode{
		id:  "dialog",
		tag: "dialog",
		paint: func(x, y int) []paint.DrawCmd {
			return []paint.DrawCmd{{
				X:     x + 1,
				Y:     y + 1,
				Text:  "M",
				Style: style.NewStyle().Foreground(style.White),
			}}
		},
	}
	dialogBox := paint.NewPaintableBoxWithBounds(dialogNode, 10, 2, 8, 4)
	dialogBox.Layer = int(paint.RenderLayerModal)
	dialogBox.ZIndex = 10

	planes := paint.NewPaintablePlanes()
	planes.AddToLayer(paint.RenderLayerBase, baseBox)
	planes.AddToLayer(paint.RenderLayerModal, hostBox)
	planes.AddToLayer(paint.RenderLayerModal, dialogBox)

	if err := engine.PaintPaintablePlanes(planes, buffer); err != nil {
		t.Fatalf("PaintPaintablePlanes() error = %v", err)
	}

	dialogCell := buffer.GetContent(11, 3)
	if dialogCell.Style.FG != style.White {
		t.Fatalf("dialog cell FG = %q, want %q (visible modal content should stay undimmed)", dialogCell.Style.FG, style.White)
	}

	outsideCell := buffer.GetContent(2, 8)
	if outsideCell.Style.FG != style.BrightBlack {
		t.Fatalf("outside cell FG = %q, want %q (background should be dimmed outside modal bounds)", outsideCell.Style.FG, style.BrightBlack)
	}
}

func TestPaintEngine_PaintPaintableLayouts_UsesVisibleModalChildForBackdrop(t *testing.T) {
	engine := NewPaintEngine()
	buffer := paint.NewBuffer(30, 10)

	baseNode := &mockCustomPaintNode{
		id:  "base",
		tag: "base",
		paint: func(x, y int) []paint.DrawCmd {
			return []paint.DrawCmd{{
				X:     1,
				Y:     8,
				Text:  "B",
				Style: style.NewStyle().Foreground(style.White),
			}}
		},
	}
	baseBox := paint.NewPaintableBoxWithBounds(baseNode, 0, 0, 30, 10)

	hostVNode := rtui.NewElement("box").
		SetPortalRootId(rtui.DefaultModalPortalRootID).
		SetLayer(rtui.LayerModal)
	hostBox := paint.NewPaintableBoxWithBounds(NewVNodePaintableNode(hostVNode), 0, 0, 1, 1)
	hostBox.Layer = int(paint.RenderLayerModal)

	dialogNode := &mockCustomPaintNode{
		id:  "dialog",
		tag: "dialog",
		paint: func(x, y int) []paint.DrawCmd {
			return []paint.DrawCmd{{
				X:     x + 1,
				Y:     y + 1,
				Text:  "M",
				Style: style.NewStyle().Foreground(style.White),
			}}
		},
	}
	dialogBox := paint.NewPaintableBoxWithBounds(dialogNode, 12, 2, 8, 4)
	dialogBox.Layer = int(paint.RenderLayerModal)
	dialogBox.ZIndex = 10
	hostBox.AddChild(dialogBox)

	layouts := paint.PaintableLayouts{
		paint.RenderLayerBase:  paint.NewPaintableLayout(baseBox),
		paint.RenderLayerModal: paint.NewPaintableLayout(hostBox),
	}

	if err := engine.PaintPaintableLayouts(layouts, buffer); err != nil {
		t.Fatalf("PaintPaintableLayouts() error = %v", err)
	}

	dialogCell := buffer.GetContent(13, 3)
	if dialogCell.Style.FG != style.White {
		t.Fatalf("dialog cell FG = %q, want %q", dialogCell.Style.FG, style.White)
	}

	outsideCell := buffer.GetContent(1, 8)
	if outsideCell.Style.FG != style.BrightBlack {
		t.Fatalf("outside cell FG = %q, want %q", outsideCell.Style.FG, style.BrightBlack)
	}
}

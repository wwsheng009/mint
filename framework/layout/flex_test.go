package layout

import (
	"testing"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/paint"
)

func TestNewColumn(t *testing.T) {
	col := NewColumn()

	if col.GetDirection() != FlexColumn {
		t.Errorf("expected FlexColumn direction, got %v", col.GetDirection())
	}
}

func TestNewRow(t *testing.T) {
	row := NewRow()

	if row.GetDirection() != FlexRow {
		t.Errorf("expected FlexRow direction, got %v", row.GetDirection())
	}
}

func TestFlex_Direction(t *testing.T) {
	flex := NewFlex()

	flex.Direction(FlexRow)
	if flex.GetDirection() != FlexRow {
		t.Errorf("expected FlexRow, got %v", flex.GetDirection())
	}

	flex.Direction(FlexColumn)
	if flex.GetDirection() != FlexColumn {
		t.Errorf("expected FlexColumn, got %v", flex.GetDirection())
	}
}

func TestFlex_Alignment(t *testing.T) {
	flex := NewFlex()

	flex.MainAlign(MainCenter)
	if flex.GetMainAlign() != MainCenter {
		t.Errorf("expected MainCenter, got %v", flex.GetMainAlign())
	}

	flex.CrossAlign(CrossEnd)
	if flex.GetCrossAlign() != CrossEnd {
		t.Errorf("expected CrossEnd, got %v", flex.GetCrossAlign())
	}
}

func TestFlex_Gap(t *testing.T) {
	flex := NewFlex()

	flex.Gap(5)
	if flex.GetGap() != 5 {
		t.Errorf("expected gap 5, got %d", flex.GetGap())
	}

	flex.CrossGap(3)
	if flex.GetCrossGap() != 3 {
		t.Errorf("expected cross gap 3, got %d", flex.GetCrossGap())
	}
}

func TestFlex_Padding(t *testing.T) {
	flex := NewFlex()

	flex.Padding(10)
	p := flex.GetPadding()
	if p.Top != 10 || p.Right != 10 || p.Bottom != 10 || p.Left != 10 {
		t.Errorf("expected uniform padding 10, got %+v", p)
	}

	flex.PaddingV(5)
	p = flex.GetPadding()
	if p.Top != 5 || p.Bottom != 5 {
		t.Errorf("expected vertical padding 5, got %+v", p)
	}

	flex.PaddingH(7)
	p = flex.GetPadding()
	if p.Left != 7 || p.Right != 7 {
		t.Errorf("expected horizontal padding 7, got %+v", p)
	}

	flex.PaddingLTRB(1, 2, 3, 4)
	p = flex.GetPadding()
	if p.Top != 2 || p.Right != 3 || p.Bottom != 4 || p.Left != 1 {
		t.Errorf("expected LTRB padding 1,2,3,4, got %+v", p)
	}
}

func TestFlex_AddChild(t *testing.T) {
	flex := NewRow()
	child1 := component.NewBaseComponent("child1")
	child2 := component.NewBaseComponent("child2")

	flex.AddChild(child1).AddChild(child2)

	if flex.ChildCount() != 2 {
		t.Errorf("expected 2 children, got %d", flex.ChildCount())
	}

	if flex.GetChild(0) != child1 {
		t.Error("expected first child to be child1")
	}
}

func TestFlex_AddChildren(t *testing.T) {
	flex := NewColumn()
	children := []component.Node{
		component.NewBaseComponent("child1"),
		component.NewBaseComponent("child2"),
		component.NewBaseComponent("child3"),
	}

	flex.AddChildren(children...)

	if flex.ChildCount() != 3 {
		t.Errorf("expected 3 children, got %d", flex.ChildCount())
	}
}

func TestFlex_Remove(t *testing.T) {
	flex := NewRow()
	child1 := component.NewBaseComponent("child1")
	child2 := component.NewBaseComponent("child2")
	child3 := component.NewBaseComponent("child3")

	flex.AddChildren(child1, child2, child3)
	flex.Remove(child2)

	if flex.ChildCount() != 2 {
		t.Errorf("expected 2 children after removal, got %d", flex.ChildCount())
	}

	if flex.GetChild(1) != child3 {
		t.Error("expected second child to be child3 after removal")
	}
}

func TestFlex_RemoveAt(t *testing.T) {
	flex := NewRow()
	children := []component.Node{
		component.NewBaseComponent("child1"),
		component.NewBaseComponent("child2"),
		component.NewBaseComponent("child3"),
	}
	flex.AddChildren(children...)

	flex.RemoveAt(1)

	if flex.ChildCount() != 2 {
		t.Errorf("expected 2 children after removal, got %d", flex.ChildCount())
	}

	if flex.GetChild(1) != children[2] {
		t.Error("expected second child to be child3 after removal")
	}
}

func TestFlex_ClearChildren(t *testing.T) {
	flex := NewRow()
	flex.AddChildren(
		component.NewBaseComponent("child1"),
		component.NewBaseComponent("child2"),
	)

	flex.ClearChildren()

	if flex.ChildCount() != 0 {
		t.Errorf("expected 0 children after clear, got %d", flex.ChildCount())
	}
}

func TestFlex_Flex(t *testing.T) {
	flex := NewRow()

	flex.Flex(0, FlexConfig{Grow: 1, Shrink: 1, Basis: 100})
	flex.Flex(1, FlexConfig{Grow: 2, Shrink: 0, Basis: 50})

	configs := flex.GetFlexChildren()

	if configs[0].Grow != 1 || configs[0].Basis != 100 {
		t.Errorf("expected first config Grow=1, Basis=100, got %+v", configs[0])
	}

	if configs[1].Grow != 2 || configs[1].Shrink != 0 {
		t.Errorf("expected second config Grow=2, Shrink=0, got %+v", configs[1])
	}
}

func TestFlex_FlexGrow(t *testing.T) {
	flex := NewRow()

	flex.FlexGrow(0, 2)

	configs := flex.GetFlexChildren()
	if configs[0].Grow != 2 {
		t.Errorf("expected Grow=2, got %d", configs[0].Grow)
	}
}

func TestFlex_FlexBasis(t *testing.T) {
	flex := NewRow()

	flex.FlexBasis(0, 100)

	configs := flex.GetFlexChildren()
	if configs[0].Basis != 100 {
		t.Errorf("expected Basis=100, got %d", configs[0].Basis)
	}
}

func TestFlex_ChainMethods(t *testing.T) {
	flex := NewRow().
		Direction(FlexColumn).
		MainAlign(MainCenter).
		CrossAlign(CrossCenter).
		Gap(10).
		Padding(5).
		Background("blue")

	if flex.GetDirection() != FlexColumn {
		t.Error("chain method Direction failed")
	}
	if flex.GetMainAlign() != MainCenter {
		t.Error("chain method MainAlign failed")
	}
	if flex.GetCrossAlign() != CrossCenter {
		t.Error("chain method CrossAlign failed")
	}
	if flex.GetGap() != 10 {
		t.Error("chain method Gap failed")
	}
}

func TestFlex_Measure_Empty(t *testing.T) {
	flex := NewColumn()

	w, h := flex.Measure(100, 100)

	if w != 0 || h != 0 {
		t.Errorf("expected empty flex to measure 0x0, got %dx%d", w, h)
	}
}

func TestFlex_Measure_WithPadding(t *testing.T) {
	flex := NewColumn().Padding(10)

	w, h := flex.Measure(100, 100)

	if w != 20 || h != 20 {
		t.Errorf("expected 10px padding to result in 20x20, got %dx%d", w, h)
	}
}

func TestFlex_Measure_WithChildren(t *testing.T) {
	// Create a simple mock child component
	mockChild := &mockMeasurableChild{width: 50, height: 20}

	flex := NewRow().AddChild(mockChild)

	w, h := flex.Measure(200, 100)

	if w < 50 {
		t.Errorf("expected width at least 50, got %d", w)
	}
	_ = h // Height is checked for minimum but not asserted
}

func TestColumn_Measure_Vertical(t *testing.T) {
	child1 := &mockMeasurableChild{width: 50, height: 20}
	child2 := &mockMeasurableChild{width: 50, height: 30}

	col := NewColumn().AddChildren(child1, child2)

	_, h := col.Measure(200, 200)

	// Height should be sum of children
	if h < 50 {
		t.Errorf("expected height at least 50 (20+30), got %d", h)
	}
}

func TestRow_Measure_Horizontal(t *testing.T) {
	child1 := &mockMeasurableChild{width: 30, height: 20}
	child2 := &mockMeasurableChild{width: 40, height: 20}

	row := NewRow().AddChildren(child1, child2)

	w, _ := row.Measure(200, 200)

	// Width should be sum of children
	if w < 70 {
		t.Errorf("expected width at least 70 (30+40), got %d", w)
	}
}

func TestFlex_Paint(t *testing.T) {
	flex := NewRow().
		Background("blue").
		AddChild(component.NewBaseComponent("child"))

	buf := paint.NewBuffer(80, 24)
	ctx := component.PaintContext{
		X:                0,
		Y:                0,
		AvailableWidth:  80,
		AvailableHeight: 24,
	}

	// Should not panic
	flex.Paint(ctx, buf)
}

// mockMeasurableChild is a mock component that implements Measure
type mockMeasurableChild struct {
	*component.BaseComponent
	width  int
	height int
}

func (m *mockMeasurableChild) Measure(maxWidth, maxHeight int) (int, int) {
	return m.width, m.height
}

package divider

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestNew(t *testing.T) {
	d := New()

	if d.Tag() != "divider" {
		t.Errorf("Tag() = %q, want %q", d.Tag(), "divider")
	}

	if d.Key() != "" {
		t.Errorf("Key() = %q, want empty", d.Key())
	}

	if d.DividerStyle() != StyleSolid {
		t.Errorf("DividerStyle() = %v, want StyleSolid", d.DividerStyle())
	}

	if d.Orientation() != Horizontal {
		t.Errorf("Orientation() = %v, want Horizontal", d.Orientation())
	}

	if d.Thickness() != 1 {
		t.Errorf("Thickness() = %d, want 1", d.Thickness())
	}
}

func TestVNode_ImplementsInterfaces(t *testing.T) {
	d := New()

	// Test VNode interface
	var _ rtui.VNode = d

	// Test InstanceFactory interface
	var _ rtui.InstanceFactory = d
}

func TestVNode_CreateInstance(t *testing.T) {
	d := New()
	d.SetKey("test-divider")
	d.SetLabel("Section")
	d.Double()

	inst := d.CreateInstance()
	if inst == nil {
		t.Fatal("CreateInstance() returned nil")
	}

	divInst, ok := inst.(*Instance)
	if !ok {
		t.Fatal("CreateInstance() did not return *Instance")
	}

	if divInst.label != "Section" {
		t.Errorf("label = %q, want %q", divInst.label, "Section")
	}

	if divInst.key != "test-divider" {
		t.Errorf("key = %q, want %q", divInst.key, "test-divider")
	}

	if divInst.dividerStyle != StyleDouble {
		t.Errorf("dividerStyle = %v, want StyleDouble", divInst.dividerStyle)
	}
}

func TestVNode_FluentAPI(t *testing.T) {
	d := New().
		SetLabel("Title").
		SetDividerStyle(StyleDashed).
		SetThickness(2).
		Vertical()

	if d.Label() != "Title" {
		t.Errorf("Label() = %q, want %q", d.Label(), "Title")
	}

	if d.DividerStyle() != StyleDashed {
		t.Errorf("DividerStyle() = %v, want StyleDashed", d.DividerStyle())
	}

	if d.Orientation() != Vertical {
		t.Errorf("Orientation() = %v, want Vertical", d.Orientation())
	}

	if d.Thickness() != 2 {
		t.Errorf("Thickness() = %d, want 2", d.Thickness())
	}
}

func TestVNode_ConvenienceMethods(t *testing.T) {
	d1 := New().Solid()
	if d1.DividerStyle() != StyleSolid {
		t.Error("Solid() did not set StyleSolid")
	}

	d2 := New().Dashed()
	if d2.DividerStyle() != StyleDashed {
		t.Error("Dashed() did not set StyleDashed")
	}

	d3 := New().Dotted()
	if d3.DividerStyle() != StyleDotted {
		t.Error("Dotted() did not set StyleDotted")
	}

	d4 := New().Double()
	if d4.DividerStyle() != StyleDouble {
		t.Error("Double() did not set StyleDouble")
	}

	d5 := New().Horizontal()
	if d5.Orientation() != Horizontal {
		t.Error("Horizontal() did not set Horizontal")
	}

	d6 := New().Vertical()
	if d6.Orientation() != Vertical {
		t.Error("Vertical() did not set Vertical")
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_Measure(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *VNode
		constraints layout.Constraints
		wantWidth   int
		wantHeight  int
	}{
		{
			name:        "Default horizontal divider",
			setup:       func() *VNode { return New() },
			constraints: layout.Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 10},
			wantWidth:   50,
			wantHeight:  1,
		},
		{
			name:        "Divider with thickness 2",
			setup:       func() *VNode { return New().SetThickness(2) },
			constraints: layout.Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 10},
			wantWidth:   50,
			wantHeight:  2,
		},
		{
			name:        "Divider with label",
			setup:       func() *VNode { return New().SetLabel("Section") },
			constraints: layout.Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 10},
			wantWidth:   50,
			wantHeight:  1,
		},
		{
			name:        "Divider without fillWidth",
			setup:       func() *VNode { return New().SetFillWidth(false) },
			constraints: layout.Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 10},
			wantWidth:   20, // Minimum width when not filling
			wantHeight:  1,
		},
		{
			name:        "Vertical divider",
			setup:       func() *VNode { return New().Vertical() },
			constraints: layout.Constraints{MinWidth: 0, MaxWidth: 50, MinHeight: 0, MaxHeight: 20},
			wantWidth:   1,
			wantHeight:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			inst := d.CreateInstance().(*Instance)
			size := inst.Measure(tt.constraints)

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

func TestInstance_Measure_WithStyleDimensions(t *testing.T) {
	d := New()
	s := style.Style{Width: 30, Height: 2}
	d.SetStyleProps(s)

	inst := d.CreateInstance().(*Instance)
	size := inst.Measure(layout.UnboundedConstraints())

	if size.Width != 30 {
		t.Errorf("Width = %d, want 30 (from style)", size.Width)
	}
	if size.Height != 2 {
		t.Errorf("Height = %d, want 2 (from style)", size.Height)
	}
}

func TestInstance_Paint(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *VNode
		bounds       [4]int
		wantContains string
	}{
		{
			name:         "Solid divider",
			setup:        func() *VNode { return New().Solid() },
			bounds:       [4]int{0, 0, 20, 1},
			wantContains: "─",
		},
		{
			name:         "Double divider",
			setup:        func() *VNode { return New().Double() },
			bounds:       [4]int{0, 0, 20, 1},
			wantContains: "═",
		},
		{
			name:         "Divider with label",
			setup:        func() *VNode { return New().SetLabel("Title").Solid() },
			bounds:       [4]int{0, 0, 20, 1},
			wantContains: "Title",
		},
		{
			name:         "Dashed divider",
			setup:        func() *VNode { return New().Dashed() },
			bounds:       [4]int{0, 0, 20, 1},
			wantContains: "-",
		},
		{
			name:         "Dotted divider",
			setup:        func() *VNode { return New().Dotted() },
			bounds:       [4]int{0, 0, 20, 1},
			wantContains: "·",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setup()
			inst := d.CreateInstance().(*Instance)
			inst.SetBounds(tt.bounds[0], tt.bounds[1], tt.bounds[2], tt.bounds[3])

			cmds := inst.Paint(tt.bounds[0], tt.bounds[1])
			if len(cmds) == 0 {
				t.Fatal("Paint() returned no commands")
			}

			if !strings.Contains(cmds[0].Text, tt.wantContains) {
				t.Errorf("Text = %q, want to contain %q", cmds[0].Text, tt.wantContains)
			}
		})
	}
}

func TestInstance_Paint_LabelCentered(t *testing.T) {
	d := New().SetLabel("OK").Solid()
	inst := d.CreateInstance().(*Instance)
	inst.SetBounds(0, 0, 10, 1)

	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("Paint() returned no commands")
	}

	text := cmds[0].Text
	// Label should be centered
	// "─OK─" or similar pattern
	if !strings.Contains(text, "OK") {
		t.Errorf("Text = %q, should contain 'OK'", text)
	}
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"label": "Original",
	})

	// Update props
	changed := inst.SetProps(rtui.Props{
		"label":        "Updated",
		"dividerStyle": StyleDouble,
	})

	if !changed {
		t.Error("SetProps() returned false, want true")
	}

	if inst.label != "Updated" {
		t.Errorf("label = %q, want %q", inst.label, "Updated")
	}

	if inst.dividerStyle != StyleDouble {
		t.Errorf("dividerStyle = %v, want StyleDouble", inst.dividerStyle)
	}

	// Set same props again
	changed = inst.SetProps(rtui.Props{
		"label":        "Updated",
		"dividerStyle": StyleDouble,
	})

	if changed {
		t.Error("SetProps() returned true for unchanged props, want false")
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder(t *testing.T) {
	d := NewBuilder().
		Key("test-key").
		Label("Section").
		Double().
		Thickness(2).
		Build()

	vnode, ok := d.(*VNode)
	if !ok {
		t.Fatal("Build() did not return *VNode")
	}

	if vnode.Label() != "Section" {
		t.Errorf("Label() = %q, want %q", vnode.Label(), "Section")
	}

	if vnode.Key() != "test-key" {
		t.Errorf("Key() = %q, want %q", vnode.Key(), "test-key")
	}

	if vnode.DividerStyle() != StyleDouble {
		t.Errorf("DividerStyle() = %v, want StyleDouble", vnode.DividerStyle())
	}

	if vnode.Thickness() != 2 {
		t.Errorf("Thickness() = %d, want 2", vnode.Thickness())
	}
}

func TestBuilder_BuildInstance(t *testing.T) {
	inst := NewBuilder().
		Key("test-key").
		Label("Title").
		BuildInstance()

	if inst == nil {
		t.Fatal("BuildInstance() returned nil")
	}

	if inst.Key() != "test-key" {
		t.Errorf("Key() = %q, want %q", inst.Key(), "test-key")
	}
}

// =============================================================================
// Convenience Function Tests
// =============================================================================

func TestD(t *testing.T) {
	d := D()

	vnode, ok := d.(*VNode)
	if !ok {
		t.Fatal("D() did not return *VNode")
	}

	if vnode.DividerStyle() != StyleSolid {
		t.Errorf("Default style should be StyleSolid")
	}
}

func TestH(t *testing.T) {
	d1 := H()
	vnode1, ok := d1.(*VNode)
	if !ok {
		t.Fatal("H() did not return *VNode")
	}
	if vnode1.Orientation() != Horizontal {
		t.Error("H() should return horizontal divider")
	}

	d2 := H("With Label")
	vnode2, ok := d2.(*VNode)
	if !ok {
		t.Fatal("H(label) did not return *VNode")
	}
	if vnode2.Label() != "With Label" {
		t.Errorf("Label() = %q, want %q", vnode2.Label(), "With Label")
	}
}

func TestV(t *testing.T) {
	d := V()

	vnode, ok := d.(*VNode)
	if !ok {
		t.Fatal("V() did not return *VNode")
	}

	if vnode.Orientation() != Vertical {
		t.Error("V() should return vertical divider")
	}
}

func TestWithLabel(t *testing.T) {
	d := WithLabel("Custom Label")

	vnode, ok := d.(*VNode)
	if !ok {
		t.Fatal("WithLabel() did not return *VNode")
	}

	if vnode.Label() != "Custom Label" {
		t.Errorf("Label() = %q, want %q", vnode.Label(), "Custom Label")
	}
}

func TestSection(t *testing.T) {
	d := Section("Section Title")

	vnode, ok := d.(*VNode)
	if !ok {
		t.Fatal("Section() did not return *VNode")
	}

	if !strings.Contains(vnode.Label(), "Section Title") {
		t.Errorf("Label() = %q, should contain 'Section Title'", vnode.Label())
	}

	if vnode.DividerStyle() != StyleDouble {
		t.Error("Section() should use StyleDouble")
	}
}

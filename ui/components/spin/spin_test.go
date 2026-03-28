package spin

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestVNode_New(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "spin" {
		t.Errorf("Tag = %q, want \"spin\"", v.Tag())
	}
	if !v.Spinning() {
		t.Error("Default Spinning should be true")
	}
	if v.Size() != SizeDefault {
		t.Errorf("Default Size = %d, want SizeDefault", v.Size())
	}
	if v.Tip() != "" {
		t.Errorf("Default Tip = %q, want empty", v.Tip())
	}
	if v.Delay() != 0 {
		t.Errorf("Default Delay = %d, want 0", v.Delay())
	}
}

func TestVNode_Setters(t *testing.T) {
	v := New().
		SetSpinning(false).
		SetTip("Loading...").
		SetSize(SizeLarge).
		SetDelay(200)

	if v.Spinning() {
		t.Error("Spinning should be false")
	}
	if v.Tip() != "Loading..." {
		t.Errorf("Tip = %q, want \"Loading...\"", v.Tip())
	}
	if v.Size() != SizeLarge {
		t.Errorf("Size = %d, want SizeLarge", v.Size())
	}
	if v.Delay() != 200 {
		t.Errorf("Delay = %d, want 200", v.Delay())
	}
}

func TestVNode_SizeHelpers(t *testing.T) {
	tests := []struct {
		name     string
		setFn    func(*VNode) *VNode
		wantSize Size
	}{
		{"Small", func(v *VNode) *VNode { return v.Small() }, SizeSmall},
		{"Default", func(v *VNode) *VNode { return v.Default() }, SizeDefault},
		{"Large", func(v *VNode) *VNode { return v.Large() }, SizeLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			tt.setFn(v)
			if v.Size() != tt.wantSize {
				t.Errorf("Size = %d, want %d", v.Size(), tt.wantSize)
			}
		})
	}
}

func TestVNode_ToProps(t *testing.T) {
	v := New().
		SetSpinning(true).
		SetTip("Please wait").
		SetSize(SizeSmall).
		SetDelay(100)
	v.SetKey("s1")

	props := v.Props()
	if props[propKey] != "s1" {
		t.Errorf("props[key] = %v, want \"s1\"", props[propKey])
	}
	if props[propSpinning] != true {
		t.Errorf("props[spinning] = %v, want true", props[propSpinning])
	}
	if props[propTip] != "Please wait" {
		t.Errorf("props[tip] = %v, want \"Please wait\"", props[propTip])
	}
	if props[propSize] != SizeSmall {
		t.Errorf("props[size] = %v, want SizeSmall", props[propSize])
	}
	if props[propDelay] != 100 {
		t.Errorf("props[delay] = %v, want 100", props[propDelay])
	}
}

func TestVNode_NewInstance(t *testing.T) {
	v := New().SetTip("Loading")
	inst := v.NewInstance()
	if inst == nil {
		t.Fatal("NewInstance returned nil")
	}
	if inst.tip != "Loading" {
		t.Errorf("Instance tip = %q, want \"Loading\"", inst.tip)
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestInstance_NewInstance(t *testing.T) {
	props := rtui.Props{
		propKey:      "k1",
		propSpinning: true,
		propTip:      "wait",
		propSize:     SizeLarge,
		propDelay:    50,
	}
	inst := NewInstance(props)
	if inst.Key() != "k1" {
		t.Errorf("Key = %q, want \"k1\"", inst.Key())
	}
	if !inst.spinning {
		t.Error("spinning should be true")
	}
	if inst.tip != "wait" {
		t.Errorf("tip = %q, want \"wait\"", inst.tip)
	}
	if inst.size != SizeLarge {
		t.Errorf("size = %d, want SizeLarge", inst.size)
	}
	if inst.delay != 50 {
		t.Errorf("delay = %d, want 50", inst.delay)
	}
	if !inst.WantsTick() {
		t.Error("spinning instance should want tick updates")
	}
}

func TestInstance_Lifecycle(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	if !inst.IsDirty() {
		t.Error("new instance should be dirty")
	}
	inst.MarkClean()
	if inst.IsDirty() {
		t.Error("after MarkClean should not be dirty")
	}
	inst.MarkDirty()
	if !inst.IsDirty() {
		t.Error("after MarkDirty should be dirty")
	}
	// lifecycle hooks should not panic
	inst.OnMount()
	inst.OnUnmount()
	inst.Destroy()
}

func TestInstance_SetProps(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.MarkClean()

	inst.SetProps(rtui.Props{
		propSpinning: false,
		propTip:      "done",
		propSize:     SizeSmall,
		propDelay:    300,
	})

	if inst.spinning {
		t.Error("spinning should be false after SetProps")
	}
	if inst.tip != "done" {
		t.Errorf("tip = %q, want \"done\"", inst.tip)
	}
	if inst.size != SizeSmall {
		t.Errorf("size = %d, want SizeSmall", inst.size)
	}
	if inst.delay != 300 {
		t.Errorf("delay = %d, want 300", inst.delay)
	}
	if !inst.IsDirty() {
		t.Error("should be dirty after SetProps")
	}
}

func TestInstance_GetProps(t *testing.T) {
	inst := &Instance{
		key:      "x",
		spinning: true,
		tip:      "loading",
		size:     SizeDefault,
		delay:    0,
	}
	props := inst.GetProps()
	if props[propKey] != "x" {
		t.Errorf("GetProps key = %v", props[propKey])
	}
	if props[propSpinning] != true {
		t.Errorf("GetProps spinning = %v", props[propSpinning])
	}
	if props[propTip] != "loading" {
		t.Errorf("GetProps tip = %v", props[propTip])
	}
}

func TestInstance_Measure(t *testing.T) {
	tests := []struct {
		name       string
		spinning   bool
		tip        string
		size       Size
		wantHeight int
		wantMinW   int
	}{
		{"not spinning", false, "", SizeDefault, 0, 0},
		{"spinning no tip", true, "", SizeDefault, 1, 1},
		{"spinning with tip", true, "Loading", SizeDefault, 2, 1},
		{"small no tip", true, "", SizeSmall, 1, 1},
		{"large no tip", true, "", SizeLarge, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &Instance{spinning: tt.spinning, tip: tt.tip, size: tt.size}
			size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 24})
			if size.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", size.Height, tt.wantHeight)
			}
			if tt.spinning && size.Width < tt.wantMinW {
				t.Errorf("Width = %d, want >= %d", size.Width, tt.wantMinW)
			}
		})
	}
}

func TestInstance_Paint_NotSpinning(t *testing.T) {
	inst := &Instance{spinning: false}
	cmds := inst.Paint(0, 0)
	if len(cmds) != 0 {
		t.Errorf("expected 0 draw cmds when not spinning, got %d", len(cmds))
	}
}

func TestInstance_Paint_Spinning(t *testing.T) {
	inst := &Instance{spinning: true, size: SizeDefault, frame: 0}
	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("expected draw cmds when spinning")
	}
	if cmds[0].X != 0 || cmds[0].Y != 0 {
		t.Errorf("Paint position = (%d,%d), want (0,0)", cmds[0].X, cmds[0].Y)
	}
	if cmds[0].Text == "" {
		t.Error("Paint text should not be empty")
	}
	if inst.frame != 0 {
		t.Error("Paint should not mutate frame state")
	}
}

func TestInstance_Paint_WithTip(t *testing.T) {
	inst := &Instance{spinning: true, tip: "Loading...", size: SizeDefault, frame: 0}
	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("expected draw cmds")
	}
	if !strings.Contains(cmds[0].Text, "Loading...") {
		t.Errorf("Paint text %q does not contain tip", cmds[0].Text)
	}
}

func TestInstance_TickFrame(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.MarkClean()
	initFrame := inst.frame
	inst.TickFrame()
	if inst.frame != (initFrame+1)%len(framesDefault) {
		t.Errorf("frame = %d, want %d", inst.frame, (initFrame+1)%len(framesDefault))
	}
	if !inst.IsDirty() {
		t.Error("TickFrame should mark instance dirty")
	}
}

func TestInstance_Tick(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.MarkClean()

	start := time.Unix(0, 0)
	if changed := inst.Tick(start); changed {
		t.Fatal("first tick should only prime the loop")
	}
	if changed := inst.Tick(start.Add(spinFrameInterval)); !changed {
		t.Fatal("tick should advance the spinner after one frame interval")
	}
	if inst.frame == 0 {
		t.Fatal("frame should advance after tick")
	}
	if !inst.IsDirty() {
		t.Fatal("Tick should mark spinner dirty")
	}
}

func TestInstance_DelayPreventsImmediatePaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propSpinning: true,
		propDelay:    50,
	})
	if cmds := inst.Paint(0, 0); len(cmds) != 0 {
		t.Fatalf("expected no paint output before delay elapses, got %d cmds", len(cmds))
	}

	start := time.Unix(0, 0)
	inst.Tick(start)
	if changed := inst.Tick(start.Add(40 * time.Millisecond)); changed {
		t.Fatal("spinner should remain hidden before delay elapses")
	}
	if cmds := inst.Paint(0, 0); len(cmds) != 0 {
		t.Fatalf("expected no paint output at 40ms, got %d cmds", len(cmds))
	}
	if changed := inst.Tick(start.Add(60 * time.Millisecond)); !changed {
		t.Fatal("spinner should become visible after delay")
	}
	if cmds := inst.Paint(0, 0); len(cmds) == 0 {
		t.Fatal("expected paint output after delay elapses")
	}
}

func TestInstance_SetBounds(t *testing.T) {
	inst := &Instance{}
	inst.SetBounds(10, 20, 30, 5)
	if inst.bounds != [4]int{10, 20, 30, 5} {
		t.Errorf("bounds = %v, want [10 20 30 5]", inst.bounds)
	}
}

func TestInstance_AllFrames(t *testing.T) {
	// verify all frame indices cycle without panic
	for _, size := range []Size{SizeSmall, SizeDefault, SizeLarge} {
		inst := &Instance{spinning: true, size: size}
		var frames []string
		switch size {
		case SizeSmall:
			frames = framesSmall
		case SizeLarge:
			frames = framesLarge
		default:
			frames = framesDefault
		}
		for i := 0; i < len(frames)*2; i++ {
			inst.frame = i
			cmds := inst.Paint(0, 0)
			if len(cmds) == 0 {
				t.Errorf("size=%d frame=%d: expected paint cmd", size, i)
			}
		}
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_Defaults(t *testing.T) {
	v := NewBuilder().Build()
	if !v.Spinning() {
		t.Error("Builder default Spinning should be true")
	}
	if v.Size() != SizeDefault {
		t.Errorf("Builder default Size = %d, want SizeDefault", v.Size())
	}
}

func TestBuilder_Fluent(t *testing.T) {
	v := NewBuilder().
		Key("b1").
		Spinning(false).
		Tip("Please wait").
		Size(SizeLarge).
		Delay(500).
		Build()

	if v.Key() != "b1" {
		t.Errorf("Key = %q, want \"b1\"", v.Key())
	}
	if v.Spinning() {
		t.Error("Spinning should be false")
	}
	if v.Tip() != "Please wait" {
		t.Errorf("Tip = %q, want \"Please wait\"", v.Tip())
	}
	if v.Size() != SizeLarge {
		t.Errorf("Size = %d, want SizeLarge", v.Size())
	}
	if v.Delay() != 500 {
		t.Errorf("Delay = %d, want 500", v.Delay())
	}
}

func TestBuilder_SizeHelpers(t *testing.T) {
	if NewBuilder().Small().Build().Size() != SizeSmall {
		t.Error("Small() should set SizeSmall")
	}
	if NewBuilder().Default().Build().Size() != SizeDefault {
		t.Error("Default() should set SizeDefault")
	}
	if NewBuilder().Large().Build().Size() != SizeLarge {
		t.Error("Large() should set SizeLarge")
	}
}

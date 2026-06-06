package pageviewport

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNodeChildrenPreservesChildTree(t *testing.T) {
	child := rtui.NewElement("button")
	node := NewBuilder().Child(child).Width(20).Height(5).ScrollOffset(2).BuildVNode()

	children := node.Children()
	if len(children) != 1 || children[0] != child {
		t.Fatalf("children = %#v, want original child", children)
	}
	props := node.Props()
	if props.GetInt(propWidth) != 20 || props.GetInt(propHeight) != 5 || props.GetInt(propScrollOffset) != 2 {
		t.Fatalf("props width/height/offset = %v/%v/%v", props[propWidth], props[propHeight], props[propScrollOffset])
	}
}

func TestInstanceMeasureAndScrollViewport(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 30, propHeight: 7, propScrollOffset: -2})
	size := inst.Measure(layout.NewConstraints(0, 80, 0, 24))
	if size.Width != 30 || size.Height != 7 {
		t.Fatalf("size = %dx%d, want 30x7", size.Width, size.Height)
	}
	viewport := inst.GetScrollViewport()
	if !viewport.Enabled || viewport.Width != 30 || viewport.Height != 7 || viewport.ScrollOffset != 0 {
		t.Fatalf("viewport = %#v", viewport)
	}
}

func TestInstanceHandlesUncontrolledPageScroll(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 30, propHeight: 4})
	inst.SetScrollViewportMetrics(30, 10, 30, 4)

	if !inst.HandleAction(action.NewAction(action.ActionNavigatePageDown)) {
		t.Fatal("PageDown should be handled in uncontrolled mode")
	}
	if got := inst.GetScrollViewport().ScrollOffset; got != 4 {
		t.Fatalf("offset after PageDown = %d, want 4", got)
	}
	if !inst.HandleAction(action.NewAction(action.ActionNavigateEnd)) {
		t.Fatal("End should be handled in uncontrolled mode")
	}
	if got := inst.GetScrollViewport().ScrollOffset; got != 6 {
		t.Fatalf("offset after End = %d, want max 6", got)
	}
	if !inst.HandleAction(action.NewAction(action.ActionNavigatePageUp)) {
		t.Fatal("PageUp should be handled in uncontrolled mode")
	}
	if got := inst.GetScrollViewport().ScrollOffset; got != 2 {
		t.Fatalf("offset after PageUp = %d, want 2", got)
	}
}

func TestInstanceHandlesUncontrolledWheelScroll(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 30, propHeight: 4})
	inst.SetScrollViewportMetrics(30, 10, 30, 4)
	mouse := runtimemsg.NewMouseMsgWithDelta(0, 0, -1, runtimemsg.MouseActionWheel)

	if !inst.HandleAction(action.NewActionWithPayload(action.ActionScroll, mouse)) {
		t.Fatal("wheel scroll down should be handled in uncontrolled mode")
	}
	if got := inst.GetScrollViewport().ScrollOffset; got != 1 {
		t.Fatalf("offset after wheel down = %d, want 1", got)
	}
}

func TestInstanceControlledOffsetDoesNotHandleScrollActions(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 30, propHeight: 4, propScrollOffset: 3})
	inst.SetScrollViewportMetrics(30, 10, 30, 4)

	if inst.HandleAction(action.NewAction(action.ActionNavigatePageDown)) {
		t.Fatal("controlled PageViewport should not mutate internal offset")
	}
	if got := inst.GetScrollViewport().ScrollOffset; got != 3 {
		t.Fatalf("controlled offset = %d, want 3", got)
	}
}

func TestInstancePostPaintDrawsScrollIndicatorOnlyWhenOverflowing(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 10, propHeight: 4})
	inst.SetBounds(2, 3, 10, 4)
	inst.SetScrollViewportMetrics(10, 8, 10, 4)

	cmds := inst.PostPaint(2, 3)
	if len(cmds) != 4 {
		t.Fatalf("indicator commands = %d, want one per viewport row", len(cmds))
	}
	if cmds[0].X != 11 || cmds[0].Y != 3 || cmds[0].Text != "#" {
		t.Fatalf("top indicator cmd = %+v, want thumb at right edge", cmds[0])
	}
	if cmds[len(cmds)-1].Text != "v" {
		t.Fatalf("bottom indicator = %q, want down affordance", cmds[len(cmds)-1].Text)
	}

	inst.SetScrollViewportMetrics(10, 4, 10, 4)
	if cmds := inst.PostPaint(2, 3); len(cmds) != 0 {
		t.Fatalf("indicator commands without overflow = %#v, want none", cmds)
	}
}

func TestInstancePostPaintCanBeHidden(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 10, propHeight: 4, propShowIndicator: false})
	inst.SetBounds(2, 3, 10, 4)
	inst.SetScrollViewportMetrics(10, 8, 10, 4)
	if cmds := inst.PostPaint(2, 3); len(cmds) != 0 {
		t.Fatalf("hidden indicator commands = %#v, want none", cmds)
	}
}

func TestInstancePostPaintReflectsScrolledPosition(t *testing.T) {
	inst := NewInstance(rtui.Props{propWidth: 10, propHeight: 4, propScrollOffset: 4})
	inst.SetBounds(2, 3, 10, 4)
	inst.SetScrollViewportMetrics(10, 8, 10, 4)

	cmds := inst.PostPaint(2, 3)
	texts := collectDrawTexts(cmds)
	if texts != "^|##" {
		t.Fatalf("indicator texts = %q, want scrolled thumb with up affordance", texts)
	}
}

func collectDrawTexts(cmds []paint.DrawCmd) string {
	out := ""
	for _, cmd := range cmds {
		out += cmd.Text
	}
	return out
}

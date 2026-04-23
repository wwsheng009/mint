package notification

import (
	"testing"
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode Tests
// =============================================================================

func TestNew_Defaults(t *testing.T) {
	v := New()
	if v.key != "" {
		t.Errorf("key: got %q, want empty", v.key)
	}
	if v.notificationType != NotificationInfo {
		t.Errorf("type: got %d, want NotificationInfo", v.notificationType)
	}
	if !v.closable {
		t.Error("closable: want true by default")
	}
	if v.placement != PlacementTopRight {
		t.Errorf("placement: got %d, want PlacementTopRight", v.placement)
	}
	if v.duration != 0 {
		t.Errorf("duration: got %v, want 0 (persistent)", v.duration)
	}
}

func TestVNode_TypedSetters(t *testing.T) {
	v := New().
		SetTitle("Hello").
		SetMessage("World").
		Success().
		SetDuration(3 * time.Second).
		SetPlacement(PlacementBottomRight)

	if v.title != "Hello" {
		t.Errorf("title: got %q", v.title)
	}
	if v.message != "World" {
		t.Errorf("message: got %q", v.message)
	}
	if v.notificationType != NotificationSuccess {
		t.Errorf("type: got %d, want NotificationSuccess", v.notificationType)
	}
	if v.duration != 3*time.Second {
		t.Errorf("duration: got %v", v.duration)
	}
	if v.placement != PlacementBottomRight {
		t.Errorf("placement: got %d", v.placement)
	}
}

func TestVNode_InterfaceMethodsReturnVNode(t *testing.T) {
	v := New()
	var iface rtui.VNode = v
	if got := iface.SetKey("k1"); got == nil {
		t.Error("SetKey returned nil")
	}
	if v.Key() != "k1" {
		t.Errorf("Key: got %q, want k1", v.Key())
	}
}

func TestVNode_Props_RoundTrip(t *testing.T) {
	v := New().
		SetTitle("T").
		SetMessage("M").
		Error().
		SetDuration(5 * time.Second)
	v.SetClosable(false)

	props := v.Props()

	v2 := New()
	v2.SetProps(props)

	if v2.title != "T" {
		t.Errorf("title: got %q", v2.title)
	}
	if v2.message != "M" {
		t.Errorf("message: got %q", v2.message)
	}
	if v2.notificationType != NotificationError {
		t.Errorf("type: got %d", v2.notificationType)
	}
	if v2.duration != 5*time.Second {
		t.Errorf("duration: got %v", v2.duration)
	}
	if v2.closable {
		t.Error("closable: want false")
	}
}

func TestVNode_Layer(t *testing.T) {
	v := New()
	if v.GetLayer() != rtui.LayerOverlay {
		t.Errorf("layer: got %v, want LayerOverlay", v.GetLayer())
	}
}

// =============================================================================
// Instance Tests
// =============================================================================

func TestNewInstance_Defaults(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	if inst.key != "" {
		t.Errorf("key: got %q", inst.key)
	}
	if inst.notificationType != NotificationInfo {
		t.Errorf("type: got %d", inst.notificationType)
	}
	if !inst.closable {
		t.Error("closable: want true by default")
	}
	if inst.IsVisible() {
		t.Error("visible: want false before Show")
	}
	if inst.WantsTick() {
		t.Error("persistent hidden notification should not want ticks")
	}
}

func TestInstance_ShowHide(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	if inst.IsVisible() {
		t.Error("should not be visible initially")
	}
	inst.Show()
	if !inst.IsVisible() {
		t.Error("should be visible after Show")
	}
	inst.Hide()
	if inst.IsVisible() {
		t.Error("should not be visible after Hide")
	}
}

func TestInstance_IsExpired_Persistent(t *testing.T) {
	inst := NewInstance(rtui.Props{propDuration: time.Duration(0)})
	inst.showAt(time.Unix(0, 0))
	if inst.IsExpired() {
		t.Error("persistent notification should never expire")
	}
	if inst.WantsTick() {
		t.Error("persistent notification should not want tick updates")
	}
}

func TestInstance_IsExpired_Timed(t *testing.T) {
	inst := NewInstance(rtui.Props{propDuration: 10 * time.Millisecond})
	start := time.Unix(0, 0)
	inst.showAt(start)
	if !inst.WantsTick() {
		t.Fatal("timed notification should want tick updates after Show")
	}
	if inst.IsExpired() {
		t.Error("should not be expired immediately")
	}
	if changed := inst.Tick(start.Add(5 * time.Millisecond)); changed {
		t.Fatal("notification should not change before duration elapses")
	}
	if changed := inst.Tick(start.Add(20 * time.Millisecond)); !changed {
		t.Fatal("notification should report a state change when it expires")
	}
	if !inst.IsExpired() {
		t.Error("should be expired after duration-driven tick")
	}
	if inst.IsVisible() {
		t.Error("expired notification should become hidden")
	}
	if inst.WantsTick() {
		t.Error("expired notification should stop requesting ticks")
	}
}

func TestInstance_HideStopsTicking(t *testing.T) {
	inst := NewInstance(rtui.Props{propDuration: 50 * time.Millisecond})
	inst.showAt(time.Unix(0, 0))
	inst.Hide()

	if inst.WantsTick() {
		t.Fatal("hidden notification should not want ticks")
	}
	if inst.IsExpired() {
		t.Fatal("manual hide should not mark notification expired")
	}
}

func TestInstance_SetProps_DoesNotResetDeadlineWhenDurationUnchanged(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMessage:  "before",
		propDuration: 10 * time.Millisecond,
	})
	start := time.Unix(0, 0)
	inst.showAt(start)

	changed := inst.SetProps(rtui.Props{
		propMessage:  "after",
		propDuration: 10 * time.Millisecond,
	})
	if !changed {
		t.Fatal("SetProps should report changed when message changes")
	}
	if changed := inst.Tick(start.Add(20 * time.Millisecond)); !changed {
		t.Fatal("notification should still expire on the original deadline")
	}
	if !inst.IsExpired() {
		t.Fatal("notification should expire after its original duration")
	}
}

func TestInstance_Paint_Hidden(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:   "T",
		propMessage: "M",
	})
	// not shown — Paint should return nil
	cmds := inst.Paint(0, 0)
	if len(cmds) != 0 {
		t.Errorf("hidden Paint: got %d cmds, want 0", len(cmds))
	}
}

func TestInstance_Paint_Visible(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "Alert",
		propMessage:  "Something happened",
		propClosable: true,
	})
	inst.SetBounds(0, 0, 40, 3)
	inst.Show()
	cmds := inst.Paint(0, 0)
	// title + message + close hint = 3 rows
	if len(cmds) != 3 {
		t.Errorf("visible Paint: got %d cmds, want 3", len(cmds))
	}
}

func TestInstance_Paint_NoTitle_NoClose(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propMessage:  "msg",
		propClosable: false,
	})
	inst.SetBounds(0, 0, 40, 2)
	inst.Show()
	cmds := inst.Paint(0, 0)
	// auto title (type label) + message = 2 rows; closable=false so no hint
	if len(cmds) != 2 {
		t.Errorf("no-title Paint: got %d cmds, want 2", len(cmds))
	}
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestBuilder_Fluent(t *testing.T) {
	inst := NewBuilder("msg").
		Title("title").
		Warning().
		Duration(2 * time.Second).
		Placement(PlacementTopLeft).
		Closable(false).
		BuildInstance()

	if inst.title != "title" {
		t.Errorf("title: got %q", inst.title)
	}
	if inst.notificationType != NotificationWarning {
		t.Errorf("type: got %d", inst.notificationType)
	}
	if inst.duration != 2*time.Second {
		t.Errorf("duration: got %v", inst.duration)
	}
	if inst.placement != PlacementTopLeft {
		t.Errorf("placement: got %d", inst.placement)
	}
	if inst.closable {
		t.Error("closable: want false")
	}
}

func TestBuilder_Build_ReturnsVNode(t *testing.T) {
	v := NewBuilder("hello").Info().Build()
	if v == nil {
		t.Fatal("Build returned nil")
	}
	if v.message != "hello" {
		t.Errorf("message: got %q", v.message)
	}
}

// =============================================================================
// Manager Tests
// =============================================================================

func TestManager_AddAndCount(t *testing.T) {
	m := NewManager()
	if m.Count() != 0 {
		t.Error("empty manager should have count 0")
	}
	m.Info("Title", "message")
	if m.Count() != 1 {
		t.Errorf("count: got %d, want 1", m.Count())
	}
	if m.IsEmpty() {
		t.Error("IsEmpty: want false")
	}
}

func TestManager_DismissAll(t *testing.T) {
	m := NewManager()
	m.Info("A", "a")
	m.Success("B", "b")
	m.DismissAll()
	if m.Count() != 0 {
		t.Errorf("after DismissAll count: got %d, want 0", m.Count())
	}
}

func TestManager_Dismiss_ByIndex(t *testing.T) {
	m := NewManager()
	m.Info("A", "a")
	m.Info("B", "b")
	m.Info("C", "c")
	m.Dismiss(1)
	if m.Count() != 2 {
		t.Errorf("after Dismiss(1) count: got %d, want 2", m.Count())
	}
	if m.Active()[0].title != "A" || m.Active()[1].title != "C" {
		t.Error("wrong notifications remain after Dismiss(1)")
	}
}

func TestManager_Dismiss_OutOfRange(t *testing.T) {
	m := NewManager()
	m.Info("A", "a")
	m.Dismiss(5)  // should not panic
	m.Dismiss(-1) // should not panic
	if m.Count() != 1 {
		t.Errorf("count: got %d, want 1", m.Count())
	}
}

func TestManager_Tick_RemovesExpired(t *testing.T) {
	m := NewManager()
	m.InfoTimed("T", "m", 10*time.Millisecond)
	m.Info("Persistent", "stays")
	start := time.Unix(0, 0)
	m.notifications[0].showAt(start)
	m.tickAt(start.Add(20 * time.Millisecond))
	if m.Count() != 1 {
		t.Errorf("after Tick count: got %d, want 1", m.Count())
	}
	if m.Active()[0].title != "Persistent" {
		t.Errorf("wrong notification remains: %q", m.Active()[0].title)
	}
}

func TestManager_AddFromVNode(t *testing.T) {
	m := NewManager()
	v := NewBuilder("msg").Title("vnode title").Error().Build()
	m.AddFromVNode(v)
	if m.Count() != 1 {
		t.Errorf("count: got %d, want 1", m.Count())
	}
	if m.Active()[0].title != "vnode title" {
		t.Errorf("title: got %q", m.Active()[0].title)
	}
}

func TestManager_Render(t *testing.T) {
	m := NewManager()
	m.Info("Title1", "Message1")
	m.Success("Title2", "Message2")
	cmds := m.Render(0, 0, 40)
	// each notification: title + message (closable=true adds close hint = 3 rows each)
	if len(cmds) == 0 {
		t.Error("Render: got 0 cmds")
	}
}

// =============================================================================
// Helpers Tests
// =============================================================================

func TestGetNotificationTypeProp(t *testing.T) {
	props := rtui.Props{propNotificationType: NotificationError}
	got := getNotificationTypeProp(props, NotificationInfo)
	if got != NotificationError {
		t.Errorf("got %d, want NotificationError", got)
	}
}

func TestGetNotificationTypeProp_Default(t *testing.T) {
	got := getNotificationTypeProp(rtui.Props{}, NotificationWarning)
	if got != NotificationWarning {
		t.Errorf("got %d, want NotificationWarning", got)
	}
}

func TestGetDurationProp(t *testing.T) {
	props := rtui.Props{propDuration: 5 * time.Second}
	got := getDurationProp(props, 0)
	if got != 5*time.Second {
		t.Errorf("got %v", got)
	}
}

func TestGetPlacementProp(t *testing.T) {
	props := rtui.Props{propPlacement: PlacementBottomLeft}
	got := getPlacementProp(props, PlacementTopRight)
	if got != PlacementBottomLeft {
		t.Errorf("got %d", got)
	}
}

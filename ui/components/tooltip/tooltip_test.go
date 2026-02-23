package tooltip

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Tooltip VNode Tests
// =============================================================================

func TestNewTooltip(t *testing.T) {
	content := newtext.New("button")
	tooltip := New(content, "Click me")

	if tooltip.Tag() != "tooltip" {
		t.Errorf("Expected tag 'tooltip', got '%s'", tooltip.Tag())
	}

	if tooltip.Text() != "Click me" {
		t.Errorf("Expected text 'Click me', got '%s'", tooltip.Text())
	}

	if tooltip.Position() != PositionAuto {
		t.Errorf("Expected position PositionAuto, got %v", tooltip.Position())
	}

	if tooltip.Delay() != 500*time.Millisecond {
		t.Errorf("Expected delay 500ms, got %v", tooltip.Delay())
	}
}

func TestTooltipBuilder(t *testing.T) {
	content := newtext.New("button")
	tooltip := NewBuilder(content, "Help").
		Key("tooltip1").
		Position(PositionTop).
		Delay(1000 * time.Millisecond).
		Build()

	tooltipVNode, ok := tooltip.(*VNode)
	if !ok {
		t.Fatal("Expected *VNode")
	}

	if tooltipVNode.Key() != "tooltip1" {
		t.Errorf("Expected key 'tooltip1', got '%s'", tooltipVNode.Key())
	}

	if tooltipVNode.Position() != PositionTop {
		t.Errorf("Expected position PositionTop, got %v", tooltipVNode.Position())
	}

	if tooltipVNode.Delay() != 1000*time.Millisecond {
		t.Errorf("Expected delay 1000ms, got %v", tooltipVNode.Delay())
	}
}

func TestTooltipPositionShortcuts(t *testing.T) {
	content := newtext.New("button")

	tests := []struct {
		name     string
		builder  *Builder
		expected Position
	}{
		{"Top", NewBuilder(content, "text").Top(), PositionTop},
		{"Bottom", NewBuilder(content, "text").Bottom(), PositionBottom},
		{"Left", NewBuilder(content, "text").Left(), PositionLeft},
		{"Right", NewBuilder(content, "text").Right(), PositionRight},
		{"Auto", NewBuilder(content, "text").Auto(), PositionAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.builder.Build().(*VNode)
			if vnode.Position() != tt.expected {
				t.Errorf("Expected position %v, got %v", tt.expected, vnode.Position())
			}
		})
	}
}

func TestTooltipStyle(t *testing.T) {
	content := newtext.New("button")
	tooltip := NewBuilder(content, "Help").
		FgColor("white").
		BgColor("blue").
		Build()

	tooltipVNode := tooltip.(*VNode)
	s := tooltipVNode.Style()

	if s.FG != style.Color("white") {
		t.Errorf("Expected FG white, got %v", s.FG)
	}

	if s.BG != style.Color("blue") {
		t.Errorf("Expected BG blue, got %v", s.BG)
	}
}

func TestTooltipConvenienceFunc(t *testing.T) {
	content := newtext.New("button")
	tooltip := Tooltip(content, "Convenient")

	if tooltip == nil {
		t.Fatal("Expected non-nil tooltip")
	}

	if tooltip.Text() != "Convenient" {
		t.Errorf("Expected text 'Convenient', got '%s'", tooltip.Text())
	}
}

// =============================================================================
// Tooltip Instance Tests
// =============================================================================

func TestNewTooltipInstance(t *testing.T) {
	props := rtui.Props{
		"key":      "test",
		"text":     "Test tooltip",
		"position": PositionTop,
		"delay":    100 * time.Millisecond,
	}

	inst := NewInstance(props)

	if inst.Key() != "test" {
		t.Errorf("Expected key 'test', got '%s'", inst.Key())
	}

	if inst.text != "Test tooltip" {
		t.Errorf("Expected text 'Test tooltip', got '%s'", inst.text)
	}

	if inst.position != PositionTop {
		t.Errorf("Expected position PositionTop, got %v", inst.position)
	}
}

func TestTooltipShowHide(t *testing.T) {
	inst := NewInstance(rtui.Props{"text": "Test"})

	// Initially not visible
	if inst.visible {
		t.Error("Expected invisible initially")
	}

	// Show
	inst.Show()
	if !inst.visible {
		t.Error("Expected visible after Show()")
	}

	// Hide
	inst.Hide()
	if inst.visible {
		t.Error("Expected invisible after Hide()")
	}
}

func TestTooltipCalculatePosition(t *testing.T) {
	inst := NewInstance(rtui.Props{"text": "Test"})

	tests := []struct {
		name          string
		position      Position
		anchorX, anchorY, anchorW, anchorH int
		expectedX, expectedY int
	}{
		// Test text is "Test" (4 chars), tooltip width = 4 + 2 = 6
		{"Top", PositionTop, 10, 10, 20, 5, 17, 8},    // X = 10 + 10 - 3 = 17
		{"Bottom", PositionBottom, 10, 10, 20, 5, 17, 16}, // X = 10 + 10 - 3 = 17
		{"Left", PositionLeft, 10, 10, 20, 5, 3, 12},     // X = 10 - 6 - 1 = 3
		{"Right", PositionRight, 10, 10, 20, 5, 31, 12},  // X = 10 + 20 + 1 = 31
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst.position = tt.position
			inst.SetAnchorBounds(tt.anchorX, tt.anchorY, tt.anchorW, tt.anchorH)
			x, y := inst.CalculatePosition()

			if x != tt.expectedX || y != tt.expectedY {
				t.Errorf("Expected position (%d, %d), got (%d, %d)", tt.expectedX, tt.expectedY, x, y)
			}
		})
	}
}

// =============================================================================
// Toast VNode Tests
// =============================================================================

func TestNewToast(t *testing.T) {
	toast := NewToast("Test message")

	if toast.Tag() != "toast" {
		t.Errorf("Expected tag 'toast', got '%s'", toast.Tag())
	}

	if toast.Message() != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", toast.Message())
	}

	if toast.ToastType() != ToastInfo {
		t.Errorf("Expected type ToastInfo, got %v", toast.ToastType())
	}

	if toast.Duration() != 3000*time.Millisecond {
		t.Errorf("Expected duration 3000ms, got %v", toast.Duration())
	}
}

func TestToastBuilder(t *testing.T) {
	toast := NewToastBuilder("Test message").
		Key("toast1").
		Title("Info").
		Type(ToastSuccess).
		Duration(5000 * time.Millisecond).
		Build()

	toastVNode, ok := toast.(*ToastVNode)
	if !ok {
		t.Fatal("Expected *ToastVNode")
	}

	if toastVNode.Key() != "toast1" {
		t.Errorf("Expected key 'toast1', got '%s'", toastVNode.Key())
	}

	if toastVNode.Title() != "Info" {
		t.Errorf("Expected title 'Info', got '%s'", toastVNode.Title())
	}

	if toastVNode.ToastType() != ToastSuccess {
		t.Errorf("Expected type ToastSuccess, got %v", toastVNode.ToastType())
	}

	if toastVNode.Duration() != 5000*time.Millisecond {
		t.Errorf("Expected duration 5000ms, got %v", toastVNode.Duration())
	}
}

func TestToastTypeShortcuts(t *testing.T) {
	tests := []struct {
		name     string
		builder  *ToastBuilder
		expected ToastType
	}{
		{"Info", NewToastBuilder("text").Info(), ToastInfo},
		{"Success", NewToastBuilder("text").Success(), ToastSuccess},
		{"Warning", NewToastBuilder("text").Warning(), ToastWarning},
		{"Error", NewToastBuilder("text").Error(), ToastError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.builder.Build().(*ToastVNode)
			if vnode.ToastType() != tt.expected {
				t.Errorf("Expected type %v, got %v", tt.expected, vnode.ToastType())
			}
		})
	}
}

func TestToastConvenienceFuncs(t *testing.T) {
	info := Info("Info message")
	if info.ToastType() != ToastInfo {
		t.Errorf("Expected type ToastInfo, got %v", info.ToastType())
	}

	success := Success("Success message")
	if success.ToastType() != ToastSuccess {
		t.Errorf("Expected type ToastSuccess, got %v", success.ToastType())
	}

	warning := Warning("Warning message")
	if warning.ToastType() != ToastWarning {
		t.Errorf("Expected type ToastWarning, got %v", warning.ToastType())
	}

	error_ := Error("Error message")
	if error_.ToastType() != ToastError {
		t.Errorf("Expected type ToastError, got %v", error_.ToastType())
	}
}

// =============================================================================
// Toast Instance Tests
// =============================================================================

func TestNewToastInstance(t *testing.T) {
	props := rtui.Props{
		"key":       "test",
		"title":     "Test",
		"message":   "Test toast",
		"toastType": ToastSuccess,
		"duration":  2000 * time.Millisecond,
	}

	inst := NewToastInstance(props)

	if inst.Key() != "test" {
		t.Errorf("Expected key 'test', got '%s'", inst.Key())
	}

	if inst.title != "Test" {
		t.Errorf("Expected title 'Test', got '%s'", inst.title)
	}

	if inst.message != "Test toast" {
		t.Errorf("Expected message 'Test toast', got '%s'", inst.message)
	}

	if inst.toastType != ToastSuccess {
		t.Errorf("Expected type ToastSuccess, got %v", inst.toastType)
	}

	if !inst.visible {
		t.Error("Expected visible initially")
	}
}

func TestToastShowHide(t *testing.T) {
	inst := NewToastInstance(rtui.Props{"message": "Test"})

	// Initially visible
	if !inst.visible {
		t.Error("Expected visible initially")
	}

	// Hide
	inst.Hide()
	if inst.visible {
		t.Error("Expected invisible after Hide()")
	}

	// Show
	inst.Show()
	if !inst.visible {
		t.Error("Expected visible after Show()")
	}
}

func TestToastExpiration(t *testing.T) {
	// Short duration for testing
	duration := 50 * time.Millisecond
	inst := NewToastInstance(rtui.Props{
		"message":  "Test",
		"duration": duration,
	})

	if inst.IsExpired() {
		t.Error("Should not be expired immediately")
	}

	// Wait for expiration (longer than duration to ensure expiration)
	time.Sleep(duration + 50*time.Millisecond)

	// Debug: check actual time
	now := time.Now()
	timeSinceCreation := now.Sub(inst.createdAt)
	timeUntilExpire := inst.expireAt.Sub(now)

	t.Logf("Duration: %v, Created: %v, ExpireAt: %v", inst.toastDuration, inst.createdAt.Format("15:04:05.000"), inst.expireAt.Format("15:04:05.000"))
	t.Logf("Time since creation: %v, Time until expire: %v", timeSinceCreation, timeUntilExpire)

	if !inst.IsExpired() {
		t.Errorf("Should be expired after duration+%v", duration)
	}
}

// =============================================================================
// Toast Manager Tests
// =============================================================================

func TestNewToastManager(t *testing.T) {
	tm := NewManager()

	if tm.Count() != 0 {
		t.Errorf("Expected 0 toasts, got %d", tm.Count())
	}

	if !tm.IsEmpty() {
		t.Error("Expected empty manager")
	}
}

func TestToastManagerAdd(t *testing.T) {
	tm := NewManager()
	toast := NewToastBuilder("Test").BuildInstance()

	tm.Add(toast)

	if tm.Count() != 1 {
		t.Errorf("Expected 1 toast, got %d", tm.Count())
	}

	if tm.IsEmpty() {
		t.Error("Expected non-empty manager")
	}
}

func TestToastManagerConvenience(t *testing.T) {
	tm := NewManager()

	tm.Info("Info message")
	tm.Success("Success message")
	tm.Warning("Warning message")
	tm.Error("Error message")

	if tm.Count() != 4 {
		t.Errorf("Expected 4 toasts, got %d", tm.Count())
	}

	// Check types
	toasts := tm.GetToasts()
	types := []ToastType{ToastInfo, ToastSuccess, ToastWarning, ToastError}
	for i, toast := range toasts {
		if toast.toastType != types[i] {
			t.Errorf("Toast %d: expected type %v, got %v", i, types[i], toast.toastType)
		}
	}
}

func TestToastManagerClear(t *testing.T) {
	tm := NewManager()
	tm.Info("Test")
	tm.Info("Test2")

	if tm.Count() != 2 {
		t.Errorf("Expected 2 toasts, got %d", tm.Count())
	}

	tm.Clear()

	if tm.Count() != 0 {
		t.Errorf("Expected 0 toasts after clear, got %d", tm.Count())
	}
}

func TestToastManagerRemove(t *testing.T) {
	tm := NewManager()
	toast1 := NewToastBuilder("Test1").BuildInstance()
	toast2 := NewToastBuilder("Test2").BuildInstance()

	tm.Add(toast1)
	tm.Add(toast2)

	if tm.Count() != 2 {
		t.Errorf("Expected 2 toasts, got %d", tm.Count())
	}

	tm.Remove(toast1)

	if tm.Count() != 1 {
		t.Errorf("Expected 1 toast after remove, got %d", tm.Count())
	}

	if tm.GetToasts()[0] != toast2 {
		t.Error("Expected remaining toast to be toast2")
	}
}

func TestToastManagerHideAndRemove(t *testing.T) {
	tm := NewManager()
	toast := NewToastBuilder("Test").BuildInstance()
	tm.Add(toast)

	if !toast.visible {
		t.Error("Expected toast to be visible")
	}

	tm.HideAndRemove(toast)

	if toast.visible {
		t.Error("Expected toast to be hidden")
	}

	if tm.Count() != 0 {
		t.Errorf("Expected 0 toasts after HideAndRemove, got %d", tm.Count())
	}
}

func TestToastManagerCleanExpired(t *testing.T) {
	tm := NewManager()

	// Add active toast (long duration)
	active := NewToastBuilder("Active message").Duration(5000 * time.Millisecond).BuildInstance()
	tm.Add(active)

	// Add expired toast (very short duration)
	expired := NewToastBuilder("Expired").Duration(50 * time.Millisecond).BuildInstance()
	tm.Add(expired)
	time.Sleep(100 * time.Millisecond) // Wait for expiration

	if tm.Count() != 2 {
		t.Errorf("Expected 2 toasts before clean, got %d", tm.Count())
	}

	tm.CleanExpired()

	if tm.Count() != 1 {
		t.Errorf("Expected 1 toast after clean expired, got %d", tm.Count())
	}

	if tm.GetToasts()[0].message != "Active message" {
		t.Errorf("Remaining toast should be the active one, got '%s'", tm.GetToasts()[0].message)
	}

	if tm.GetToasts()[0].IsExpired() {
		t.Error("Remaining toast should not be expired")
	}
}

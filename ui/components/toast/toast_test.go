package toast

import (
	"testing"
	"time"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

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
	if !inst.WantsTick() {
		t.Error("visible timed toast should want tick updates")
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
	duration := 50 * time.Millisecond
	inst := NewToastInstance(rtui.Props{
		"message":  "Test",
		"duration": duration,
	})
	start := time.Unix(0, 0)
	inst.showAt(start)

	if inst.IsExpired() {
		t.Error("Should not be expired immediately")
	}
	if changed := inst.Tick(start.Add(20 * time.Millisecond)); changed {
		t.Fatal("toast should not expire before duration elapses")
	}
	if changed := inst.Tick(start.Add(duration + 10*time.Millisecond)); !changed {
		t.Fatal("toast should report a state change when it expires")
	}
	if !inst.IsExpired() {
		t.Errorf("Should be expired after duration")
	}
	if inst.visible {
		t.Error("Expired toast should become hidden")
	}
	if inst.WantsTick() {
		t.Error("Expired toast should stop requesting ticks")
	}
}

func TestToastHideStopsTicking(t *testing.T) {
	inst := NewToastInstance(rtui.Props{
		"message":  "Test",
		"duration": 50 * time.Millisecond,
	})
	inst.Hide()

	if inst.WantsTick() {
		t.Fatal("hidden toast should not want ticks")
	}
	if inst.IsExpired() {
		t.Fatal("manual hide should not mark toast expired")
	}
}

func TestToastManagerTickRemovesExpired(t *testing.T) {
	tm := NewManager()
	active := NewToastBuilder("Active").Duration(5 * time.Second).BuildInstance()
	expired := NewToastBuilder("Expired").Duration(50 * time.Millisecond).BuildInstance()
	tm.Add(active)
	tm.Add(expired)

	start := time.Unix(0, 0)
	active.showAt(start)
	expired.showAt(start)
	tm.tickAt(start.Add(100 * time.Millisecond))

	if tm.Count() != 1 {
		t.Fatalf("count after Tick = %d, want 1", tm.Count())
	}
	if tm.GetToasts()[0].Message() != "Active" {
		t.Fatalf("remaining toast = %q, want Active", tm.GetToasts()[0].Message())
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
	start := time.Unix(0, 0)
	active.showAt(start)
	expired.showAt(start)

	if tm.Count() != 2 {
		t.Errorf("Expected 2 toasts before clean, got %d", tm.Count())
	}

	tm.tickAt(start.Add(100 * time.Millisecond))

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

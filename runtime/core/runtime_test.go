package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/focus"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/platform"
)

// =============================================================================
// Mock implementations
// =============================================================================

type mockPlatform struct {
	mu           sync.Mutex
	initCalled   bool
	closed       bool
	width        int
	height       int
	inputs       []platform.RawInput
	writtenCells []string
	lastString    string
	screenCleared bool
}

func newMockPlatform() *mockPlatform {
	return &mockPlatform{
		width:   80,
		height: 24,
	}
}

func (m *mockPlatform) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initCalled = true
	return nil
}

func (m *mockPlatform) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockPlatform) Size() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.width, m.height
}

func (m *mockPlatform) ReadInput() *platform.RawInput {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.inputs) == 0 {
		return nil
	}

	input := m.inputs[0]
	m.inputs = m.inputs[1:]
	return &input
}

func (m *mockPlatform) WriteString(s string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastString = s
	m.writtenCells = append(m.writtenCells, s)
	return len(s), nil
}

func (m *mockPlatform) Flush() error {
	return nil
}

func (m *mockPlatform) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.screenCleared = true
	return nil
}

func (m *mockPlatform) SetNormalMode() {}
func (m *mockPlatform) ShowCursor() {}
func (m *mockPlatform) ExitAltScreen() {}
func (m *mockPlatform) EnableEcho() {}

func (m *mockPlatform) SendInput(input *platform.RawInput) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inputs = append(m.inputs, *input)
}

func (m *mockPlatform) SetSize(w, h int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.width = w
	m.height = h
}

// mockLayoutNode is a simple layout node for testing
type mockLayoutNode struct {
	id          string
	x, y        int
	width       int
	height      int
	children    []layout.Node
}

func (m *mockLayoutNode) ID() string                       { return m.id }
func (m *mockLayoutNode) Type() string                      { return "mock" }
func (m *mockLayoutNode) GetPosition() (int, int)          { return m.x, m.y }
func (m *mockLayoutNode) SetPosition(x, y int)              { m.x = x; m.y = y }
func (m *mockLayoutNode) GetSize() (int, int)               { return m.width, m.height }
func (m *mockLayoutNode) GetWidth() int                     { return m.width }
func (m *mockLayoutNode) GetHeight() int                    { return m.height }
func (m *mockLayoutNode) SetSize(w, h int)                  { m.width = w; m.height = h }
func (m *mockLayoutNode) Children() []layout.Node          { return m.children }
func (m *mockLayoutNode) Parent() layout.Node              { return nil }
func (m *mockLayoutNode) SetParent(parent layout.Node)     {}
func (m *mockLayoutNode) Constraints() layout.Constraints {
	return layout.Constraints{MaxWidth: 80, MaxHeight: 24}
}
func (m *mockLayoutNode) SetConstraints(c layout.Constraints) {}

// =============================================================================
// NewRuntime Tests
// =============================================================================

func TestNewRuntime(t *testing.T) {
	pf := newMockPlatform()
	rt := NewRuntime(pf)

	if rt == nil {
		t.Fatal("NewRuntime should not return nil")
	}

	// Check platform was set by checking it's not nil
	rt.mu.RLock()
	pfSet := rt.platform != nil
	rt.mu.RUnlock()
	if !pfSet {
		t.Error("platform should be set")
	}

	if rt.layoutEngine == nil {
		t.Error("layoutEngine should be initialized")
	}

	if rt.focusManager == nil {
		t.Error("focusManager should be initialized")
	}

	if rt.stateTracker == nil {
		t.Error("stateTracker should be initialized")
	}

	if rt.actionDispatcher == nil {
		t.Error("actionDispatcher should be initialized")
	}

	if rt.keyMap == nil {
		t.Error("keyMap should be initialized")
	}

	if rt.contextManager == nil {
		t.Error("contextManager should be initialized")
	}

	if rt.dirtyTracker == nil {
		t.Error("dirtyTracker should be initialized")
	}

	if rt.running {
		t.Error("should not be running initially")
	}

	if rt.windowWidth != 0 || rt.windowHeight != 0 {
		t.Error("window size should be 0 initially")
	}
}

func TestRuntime_SetRoot(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	node := &mockLayoutNode{id: "test-node"}
	rt.SetRoot(node)

	rt.mu.RLock()
	rootID := rt.root.ID()
	rt.mu.RUnlock()

	if rootID != "test-node" {
		t.Errorf("root should be set to test-node, got %s", rootID)
	}
}

func TestRuntime_GetRoot(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Initially nil
	if root := rt.GetRoot(); root != nil {
		t.Error("root should be nil initially")
	}

	// Set and get
	node := &mockLayoutNode{id: "test-node"}
	rt.SetRoot(node)

	root := rt.GetRoot()
	if root == nil || root.ID() != "test-node" {
		t.Error("root should return the set node")
	}
}

func TestRuntime_IsRunning(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Initially not running
	if rt.IsRunning() {
		t.Error("should not be running initially")
	}

	// After Start, should be running
	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !rt.IsRunning() {
		t.Error("should be running after Start")
	}

	// After Stop, should not be running
	if err := rt.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if rt.IsRunning() {
		t.Error("should not be running after Stop")
	}
}

func TestRuntime_Stop_NotRunning(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Stop when not running should return nil
	err := rt.Stop()
	if err != nil {
		t.Errorf("Stop should return nil when not running, got %v", err)
	}

	if rt.IsRunning() {
		t.Error("should still not be running")
	}
}

func TestRuntime_Start(t *testing.T) {
	pf := newMockPlatform()
	rt := NewRuntime(pf)

	err := rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	rt.mu.RLock()
	running := rt.running
	buffer := rt.buffer
	windowW := rt.windowWidth
	windowH := rt.windowHeight
	root := rt.root
	rt.mu.RUnlock()

	if !running {
		t.Error("should be running after Start")
	}

	if buffer == nil {
		t.Error("buffer should be created")
	}

	if windowW != 80 {
		t.Errorf("windowWidth should be 80, got %d", windowW)
	}

	if windowH != 24 {
		t.Errorf("windowHeight should be 24, got %d", windowH)
	}

	if !pf.initCalled {
		t.Error("platform.Init should be called")
	}

	if root != nil {
		// Root should be set if it was set before Start
		// But here root is nil, so we skip this check
	}
}

func TestRuntime_Start_AlreadyRunning(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	if err := rt.Start(); err != nil {
		t.Fatalf("first Start failed: %v", err)
	}

	// Start again when already running should return nil
	if err := rt.Start(); err != nil {
		t.Errorf("Start when already running should return nil, got %v", err)
	}
}

func TestRuntime_GetFocusManager(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	fm := rt.GetFocusManager()
	if fm == nil {
		t.Error("GetFocusManager should not return nil")
	}

	if fm != rt.focusManager {
		t.Error("GetFocusManager should return the internal focusManager")
	}
}

func TestRuntime_GetActionDispatcher(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	ad := rt.GetActionDispatcher()
	if ad == nil {
		t.Error("GetActionDispatcher should not return nil")
	}

	if ad != rt.actionDispatcher {
		t.Error("GetActionDispatcher should return the internal actionDispatcher")
	}
}

func TestRuntime_GetState(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	st := rt.GetState()
	if st == nil {
		t.Error("GetState should not return nil")
	}
}

func TestRuntime_GetState_Running(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// GetState should return a snapshot
	st := rt.GetState()
	if st == nil {
		t.Error("GetState should return a snapshot")
	}
}

func TestRuntime_Context(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	ctx := rt.Context()
	if ctx == nil {
		t.Error("Context() should not return nil")
	}

	// Should be cancelable when shutdown
	rt.Shutdown(0)

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be canceled after Shutdown")
	}
}

func TestRuntime_WithContext(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	newCtx, cancel := context.WithCancel(context.Background())
	rt.WithContext(newCtx)

	// New context manager should be created
	// The context itself is wrapped by ContextManager
	// Let's just verify it doesn't panic and is usable
	ctx := rt.Context()
	if ctx == nil {
		t.Error("Context should not be nil after WithContext")
	}

	// Cancel the context we created
	cancel()

	// Verify rt.Context is still usable (returns manager's context)
	_ = rt.Context()
}

func TestRuntime_Shutdown(t *testing.T) {
	pf := newMockPlatform()
	rt := NewRuntime(pf)

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err := rt.Shutdown(100 * time.Millisecond)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	if rt.IsRunning() {
		t.Error("should not be running after Shutdown")
	}

	if !pf.closed {
		t.Error("platform should be closed after Shutdown")
	}

	if !pf.screenCleared {
		t.Error("screen should be cleared after Shutdown")
	}
}

func TestRuntime_Shutdown_NotRunning(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Shutdown when not running
	err := rt.Shutdown(100 * time.Millisecond)
	if err != nil {
		t.Errorf("Shutdown when not running should return nil, got %v", err)
	}

	if rt.IsRunning() {
		t.Error("should not be running")
	}
}

func TestRuntime_Go(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	done := make(chan struct{})
	rt.Go(func(ctx context.Context) error {
		// Verify context is passed through
		if ctx == nil {
			t.Error("context should be passed to goroutine")
		}
		close(done)
		return nil
	})

	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Error("goroutine should complete")
	}
}

func TestRuntime_ContextValue(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	key := ContextKey("test-key")
	value := "test-value"

	rt.SetContextValue(key, value)

	retrieved := rt.ContextValue(key)
	if retrieved != value {
		t.Errorf("expected %v, got %v", value, retrieved)
	}

	// Missing key should return nil
	missing := rt.ContextValue(ContextKey("missing"))
	if missing != nil {
		t.Error("missing key should return nil")
	}
}

func TestRuntime_IsCanceled(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Initially not canceled
	if rt.IsCanceled() {
		t.Error("should not be canceled initially")
	}

	// Shutdown, which cancels context
	rt.Shutdown(0)

	if !rt.IsCanceled() {
		t.Error("should be canceled after Shutdown")
	}
}

func TestRuntime_Invalidate(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Should not panic
	rt.Invalidate()
}

func TestRuntime_InvalidateNode(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Should not panic
	rt.InvalidateNode("test-id")
}

func TestRuntime_FocusMethods(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// These delegate to focusManager - just verify they don't panic
	id1, ok1 := rt.FocusNext()
	id2, ok2 := rt.FocusPrev()

	if ok1 || id1 != "" {
		t.Error("FocusNext should return empty string and false with no focusables")
	}

	if ok2 || id2 != "" {
		t.Error("FocusPrev should return empty string and false with no focusables")
	}

	// FocusSpecific
	if rt.FocusSpecific("test") {
		t.Error("FocusSpecific should return false with non-existent component")
	}

	// GetFocused
	focused, has := rt.GetFocused()
	if has || focused != "" {
		t.Error("GetFocused should return false and empty string with no focusables")
	}
}

func TestRuntime_PushFocusScope(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	scope := focus.NewScope("test", "test")
	rt.PushFocusScope(scope)

	// Should not panic
	_ = rt.PopFocusScope()
}

func TestRuntime_RegisterActionTarget(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	target := &mockActionTarget{id: "test-target"}
	rt.RegisterActionTarget(target)

	// Should not panic
	rt.UnregisterActionTarget("test-target")
}

func TestRuntime_ProcessInput_Empty(t *testing.T) {
	pf := newMockPlatform()
	rt := NewRuntime(pf)

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// No input available
	if err := rt.ProcessInput(); err != nil {
		t.Errorf("ProcessInput should succeed with no input, got %v", err)
	}
}

func TestRuntime_ProcessInput_WithInput(t *testing.T) {
	pf := newMockPlatform()
	rt := NewRuntime(pf)

	// Add a keyboard input
	pf.SendInput(&platform.RawInput{
		Type: platform.InputKeyPress,
		Key:  'a',
	})

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Process the input
	if err := rt.ProcessInput(); err != nil {
		t.Errorf("ProcessInput failed: %v", err)
	}
}

func TestRuntime_HandleWindowSize(t *testing.T) {
	rt := NewRuntime(newMockPlatform())

	// Handle window resize
	rt.HandleWindowSize(100, 50)

	rt.mu.RLock()
	w, h := rt.windowWidth, rt.windowHeight
	buffer := rt.buffer
	rt.mu.RUnlock()

	if w != 100 {
		t.Errorf("windowWidth should be 100, got %d", w)
	}

	if h != 50 {
		t.Errorf("windowHeight should be 50, got %d", h)
	}

	if buffer == nil {
		t.Error("buffer should be recreated")
	}

	if buffer.Width != 100 || buffer.Height != 50 {
		t.Errorf("buffer size should be 100x50, got %dx%d", buffer.Width, buffer.Height)
	}
}

func TestRuntime_WriteToScreen_EmptyBuffer(t *testing.T) {
	pf := newMockPlatform()
	rt := NewRuntime(pf)

	// Start to initialize buffer
	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Render should not panic
	if err := rt.Render(); err != nil {
		t.Errorf("Render failed: %v", err)
	}

	// Buffer should be initialized
	rt.mu.RLock()
	buffer := rt.buffer
	rt.mu.RUnlock()

	if buffer == nil {
		t.Error("buffer should exist after Start")
	}
}

// =============================================================================
// mockActionTarget implements action.Target
// =============================================================================

type mockActionTarget struct {
	id string
}

func (m *mockActionTarget) ID() string {
	return m.id
}

func (m *mockActionTarget) HandleAction(act *action.Action) bool {
	return true
}

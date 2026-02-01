// mock/sandbox_test.go - 模拟沙箱测试
package mock

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

func TestNewMockSandbox(t *testing.T) {
	sb := New(80, 24)

	if sb == nil {
		t.Fatal("New() returned nil")
	}

	if sb.Type() != sandbox.TypeMock {
		t.Errorf("New() Type() = %v, want %v", sb.Type(), sandbox.TypeMock)
	}

	if sb.State() != sandbox.StateStopped {
		t.Errorf("New() State() = %v, want %v", sb.State(), sandbox.StateStopped)
	}

	width, height := sb.Size()
	if width != 80 || height != 24 {
		t.Errorf("New() Size() = %dx%d, want 80x24", width, height)
	}
}

func TestMockSandboxLifecycle(t *testing.T) {
	sb := New(80, 24)

	// Initialize
	if err := sb.Initialize(nil); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if sb.State() != sandbox.StateInitialized {
		t.Errorf("Initialize() State() = %v, want %v", sb.State(), sandbox.StateInitialized)
	}

	// Run
	if err := sb.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if sb.State() != sandbox.StateRunning {
		t.Errorf("Run() State() = %v, want %v", sb.State(), sandbox.StateRunning)
	}

	// Pause
	if err := sb.Pause(); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}

	if sb.State() != sandbox.StatePaused {
		t.Errorf("Pause() State() = %v, want %v", sb.State(), sandbox.StatePaused)
	}

	// Resume
	if err := sb.Resume(); err != nil {
		t.Fatalf("Resume() error = %v", err)
	}

	if sb.State() != sandbox.StateRunning {
		t.Errorf("Resume() State() = %v, want %v", sb.State(), sandbox.StateRunning)
	}

	// Close
	if err := sb.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if sb.State() != sandbox.StateStopped {
		t.Errorf("Close() State() = %v, want %v", sb.State(), sandbox.StateStopped)
	}
}

func TestMockSandboxInjectKey(t *testing.T) {
	sb := New(80, 24)
	sb.Initialize(nil)
	sb.Run()

	err := sb.InjectKey('a')
	if err != nil {
		t.Fatalf("InjectKey() error = %v", err)
	}

	stats := sb.QueueStats()
	if stats.Length != 1 {
		t.Errorf("InjectKey() queue length = %v, want 1", stats.Length)
	}
}

func TestMockSandboxInjectSpecialKey(t *testing.T) {
	sb := New(80, 24)
	sb.Initialize(nil)
	sb.Run()

	err := sb.InjectSpecialKey(platform.KeyEnter)
	if err != nil {
		t.Fatalf("InjectSpecialKey() error = %v", err)
	}

	stats := sb.QueueStats()
	if stats.Length != 1 {
		t.Errorf("InjectSpecialKey() queue length = %v, want 1", stats.Length)
	}
}

func TestMockSandboxInjectString(t *testing.T) {
	sb := New(80, 24)
	sb.Initialize(nil)
	sb.Run()

	text := "hello"
	err := sb.InjectString(text)
	if err != nil {
		t.Fatalf("InjectString() error = %v", err)
	}

	stats := sb.QueueStats()
	if stats.Length != len(text) {
		t.Errorf("InjectString() queue length = %v, want %v", stats.Length, len(text))
	}
}

func TestMockSandboxInjectMouse(t *testing.T) {
	sb := New(80, 24)
	sb.Initialize(nil)
	sb.Run()

	err := sb.InjectMouse(10, 20, platform.MouseLeft, platform.MousePress)
	if err != nil {
		t.Fatalf("InjectMouse() error = %v", err)
	}

	stats := sb.QueueStats()
	if stats.Length != 1 {
		t.Errorf("InjectMouse() queue length = %v, want 1", stats.Length)
	}
}

func TestMockSandboxResize(t *testing.T) {
	sb := New(80, 24)

	sb.Resize(100, 30)

	width, height := sb.Size()
	if width != 100 || height != 30 {
		t.Errorf("Resize() Size() = %dx%d, want 100x30", width, height)
	}
}

func TestMockSandboxRenderString(t *testing.T) {
	sb := New(10, 5)

	rendered := sb.RenderString()

	// 空缓冲区应该返回空字符串或正确数量的空格
	lines := 0
	for _, c := range rendered {
		if c == '\n' {
			lines++
		}
	}

	if lines > 5 {
		t.Errorf("RenderString() has too many lines: %d", lines)
	}
}

func TestMockSandboxAssertRender(t *testing.T) {
	sb := New(10, 5)

	// 空缓冲区不应该包含任何文本
	err := sb.AssertRender("nonexistent")
	if err == nil {
		t.Error("AssertRender() should fail for nonexistent text")
	}

	// 不应该包含空文本
	err = sb.AssertNotRender("nonexistent")
	if err != nil {
		t.Errorf("AssertNotRender() should pass for nonexistent text, got error: %v", err)
	}
}

func TestMockSandboxSnapshot(t *testing.T) {
	sb := New(80, 24)
	sb.Initialize(nil)
	sb.Run()

	snap, err := sb.Snapshot(sandbox.SnapshotMinimal, "test-tag")
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if snap == nil {
		t.Fatal("Snapshot() returned nil")
	}

	if snap.Metadata.Level != sandbox.SnapshotMinimal {
		t.Errorf("Snapshot() Level = %v, want %v", snap.Metadata.Level, sandbox.SnapshotMinimal)
	}

	if len(snap.Metadata.Tags) != 1 {
		t.Errorf("Snapshot() Tags length = %v, want 1", len(snap.Metadata.Tags))
	}
}

func TestMockSandboxListSnapshots(t *testing.T) {
	sb := New(80, 24)
	sb.Initialize(nil)
	sb.Run()

	sb.Snapshot(sandbox.SnapshotMinimal, "test1")
	sb.Snapshot(sandbox.SnapshotStandard, "test2")

	snapshots := sb.ListSnapshots()

	if len(snapshots) != 2 {
		t.Errorf("ListSnapshots() length = %v, want 2", len(snapshots))
	}
}

func TestTestHelperBasic(t *testing.T) {
	sb := New(80, 24)
	helper := NewTestHelper(sb)

	if helper == nil {
		t.Fatal("NewTestHelper() returned nil")
	}

	if helper.HasErrors() {
		t.Error("NewTestHelper() should have no errors initially")
	}
}

func TestTestHelperType(t *testing.T) {
	sb := New(80, 24)
	helper := NewTestHelper(sb)

	result := helper.Type("test").Result()

	if len(result.Errors) != 0 {
		t.Errorf("Type() produced errors: %v", result.Errors)
	}
}

func TestTestHelperPress(t *testing.T) {
	sb := New(80, 24)
	helper := NewTestHelper(sb)

	result := helper.Press(platform.KeyEnter).Result()

	if len(result.Errors) != 0 {
		t.Errorf("Press() produced errors: %v", result.Errors)
	}
}

func TestTestHelperChain(t *testing.T) {
	sb := New(80, 24)
	helper := NewTestHelper(sb)

	result := helper.
		Type("hello").
		Tab().
		Enter().
		Process().
		Result()

	if !result.OK() {
		t.Errorf("Chain produced errors: %v", result.Errors)
	}
}

func TestTestResultOK(t *testing.T) {
	result := TestResult{Errors: []error{}}

	if !result.OK() {
		t.Error("TestResult.OK() = false, want true")
	}
}

func TestTestResultError(t *testing.T) {
	err := sandbox.ErrAssertionFailed
	result := TestResult{Errors: []error{err}}

	if result.OK() {
		t.Error("TestResult.OK() = true, want false")
	}

	if result.Error() != err {
		t.Errorf("TestResult.Error() = %v, want %v", result.Error(), err)
	}
}

func TestMockSandboxIsMock(t *testing.T) {
	sb := New(80, 24)

	if !sb.IsMock() {
		t.Error("IsMock() = false, want true")
	}
}

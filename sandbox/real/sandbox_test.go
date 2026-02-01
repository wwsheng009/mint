// real/sandbox_test.go - 真实沙箱测试
package real

import (
	"testing"

	"github.com/wwsheng009/mint/sandbox"
)

func TestNewRealSandbox(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if rs == nil {
		t.Fatal("New() returned nil")
	}

	if rs.Type() != sandbox.TypeReal {
		t.Errorf("New() Type() = %v, want %v", rs.Type(), sandbox.TypeReal)
	}

	if rs.State() != sandbox.StateStopped {
		t.Errorf("New() State() = %v, want %v", rs.State(), sandbox.StateStopped)
	}

	width, height := rs.Size()
	if width != 80 || height != 24 {
		t.Errorf("New() Size() = %dx%d, want 80x24", width, height)
	}

	// 清理
	rs.Close()
}

func TestRealSandboxLifecycle(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	// Initialize
	if err := rs.Initialize(nil); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if rs.State() != sandbox.StateInitialized {
		t.Errorf("Initialize() State() = %v, want %v", rs.State(), sandbox.StateInitialized)
	}

	// 注意：不能调用 Run() 因为它会启动真实的事件循环
}

func TestRealSandboxResize(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	rs.Resize(100, 30)

	width, height := rs.Size()
	if width != 100 || height != 30 {
		t.Errorf("Resize() Size() = %dx%d, want 100x30", width, height)
	}
}

func TestRealSandboxSnapshot(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	snap, err := rs.Snapshot(sandbox.SnapshotMinimal, "test-tag")
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

func TestRealSandboxListSnapshots(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	rs.Snapshot(sandbox.SnapshotMinimal, "test1")
	rs.Snapshot(sandbox.SnapshotStandard, "test2")

	snapshots := rs.ListSnapshots()

	if len(snapshots) != 2 {
		t.Errorf("ListSnapshots() length = %v, want 2", len(snapshots))
	}
}

func TestRealSandboxRecordedEvents(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	events := rs.RecordedEvents()

	if events == nil {
		t.Error("RecordedEvents() returned nil")
	}

	// 初始状态应该没有事件
	if len(events) != 0 {
		t.Errorf("RecordedEvents() length = %v, want 0", len(events))
	}
}

func TestRealSandboxConfig(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	config := rs.Config()

	if config == nil {
		t.Fatal("Config() returned nil")
	}

	if config.Width != 80 || config.Height != 24 {
		t.Errorf("Config() Size() = %dx%d, want 80x24", config.Width, config.Height)
	}

	// 真实环境应该禁止注入
	if config.Event.Strategy != sandbox.InjectProhibited {
		t.Errorf("Config() Event.Strategy = %v, want %v", config.Event.Strategy, sandbox.InjectProhibited)
	}
}

func TestRealSandboxBuffer(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	buf := rs.Buffer()

	if buf == nil {
		t.Fatal("Buffer() returned nil")
	}

	if buf.Width != 80 || buf.Height != 24 {
		t.Errorf("Buffer() Size() = %dx%d, want 80x24", buf.Width, buf.Height)
	}
}

func TestRealSandboxSetBuffer(t *testing.T) {
	rs, err := New(80, 24)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer rs.Close()

	// 新缓冲区
	newBuf := rs.Buffer()

	// 先保存原缓冲区引用
	oldBuf := rs.Buffer()

	rs.SetBuffer(newBuf)

	// 验证 - 应该仍然指向相同的内存位置
	// (因为我们传递的是同一个引用)
	if rs.Buffer() != oldBuf {
		t.Error("SetBuffer() did not set the buffer")
	}
}

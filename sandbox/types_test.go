// sandbox/types_test.go - 核心类型测试
package sandbox

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/paint"
)

func TestSandboxTypeString(t *testing.T) {
	tests := []struct {
		name     string
		sandbox  SandboxType
		expected string
	}{
		{"Real", TypeReal, "real"},
		{"Mock", TypeMock, "mock"},
		{"Replay", TypeReplay, "replay"},
		{"Unknown", SandboxType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sandbox.String(); got != tt.expected {
				t.Errorf("SandboxType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		name     string
		state    State
		expected string
	}{
		{"Stopped", StateStopped, "stopped"},
		{"Initialized", StateInitialized, "initialized"},
		{"Running", StateRunning, "running"},
		{"Paused", StatePaused, "paused"},
		{"Error", StateError, "error"},
		{"Unknown", State(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInjectionStrategyString(t *testing.T) {
	tests := []struct {
		name     string
		strategy InjectionStrategy
		expected string
	}{
		{"Prohibited", InjectProhibited, "prohibited"},
		{"Allowed", InjectAllowed, "allowed"},
		{"Recorded", InjectRecorded, "recorded"},
		{"Unknown", InjectionStrategy(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.strategy.String(); got != tt.expected {
				t.Errorf("InjectionStrategy.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvictPolicyString(t *testing.T) {
	tests := []struct {
		name     string
		policy   EvictPolicy
		expected string
	}{
		{"Oldest", EvictOldest, "oldest"},
		{"Priority", EvictByPriority, "priority"},
		{"Persist", EvictPersist, "persist"},
		{"Unknown", EvictPolicy(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.policy.String(); got != tt.expected {
				t.Errorf("EvictPolicy.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSnapshotLevelString(t *testing.T) {
	tests := []struct {
		name     string
		level    SnapshotLevel
		expected string
	}{
		{"Minimal", SnapshotMinimal, "minimal"},
		{"Standard", SnapshotStandard, "standard"},
		{"Full", SnapshotFull, "full"},
		{"Unknown", SnapshotLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("SnapshotLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHookKey(t *testing.T) {
	key1 := HookKey{State: StateRunning, Phase: PhaseBefore}
	key2 := HookKey{State: StateRunning, Phase: PhaseBefore}
	key3 := HookKey{State: StateRunning, Phase: PhaseAfter}

	if key1 != key2 {
		t.Error("Identical HookKeys should be equal")
	}

	if key1 == key3 {
		t.Error("HookKeys with different phases should not be equal")
	}
}

func TestAssertionError(t *testing.T) {
	err := &AssertionError{
		Message:  "render does not contain expected text",
		Expected: "Hello",
		Actual:   "World",
	}

	expected := "render does not contain expected text: expected Hello, got World"
	if got := err.Error(); got != expected {
		t.Errorf("AssertionError.Error() = %v, want %v", got, expected)
	}
}

func TestAssertionErrorNilValues(t *testing.T) {
	err := &AssertionError{
		Message: "test error",
	}

	expected := "test error"
	if got := err.Error(); got != expected {
		t.Errorf("AssertionError.Error() = %v, want %v", got, expected)
	}
}

func TestBufferWrapper(t *testing.T) {
	buf := paint.NewBuffer(10, 5)
	bw := NewBufferWrapper(buf, 3)

	if bw.Buffer != buf {
		t.Error("BufferWrapper.Buffer should be the provided buffer")
	}

	if bw.maxHistory != 3 {
		t.Errorf("BufferWrapper.maxHistory = %v, want %v", bw.maxHistory, 3)
	}

	if len(bw.history) != 0 {
		t.Errorf("BufferWrapper.history should be empty, got length %v", len(bw.history))
	}
}

func TestBufferWrapperSaveSnapshot(t *testing.T) {
	buf := paint.NewBuffer(10, 5)
	bw := NewBufferWrapper(buf, 3)

	bw.SaveSnapshot()

	if len(bw.history) != 1 {
		t.Errorf("BufferWrapper.history length = %v, want %v", len(bw.history), 1)
	}

	snap := bw.history[0]
	if snap.Width != 10 || snap.Height != 5 {
		t.Errorf("Snapshot dimensions = %vx%v, want 10x5", snap.Width, snap.Height)
	}
}

func TestBufferWrapperMaxHistory(t *testing.T) {
	buf := paint.NewBuffer(10, 5)
	bw := NewBufferWrapper(buf, 2)

	// 保存3个快照，应该只保留最后2个
	bw.SaveSnapshot()
	bw.SaveSnapshot()
	bw.SaveSnapshot()

	if len(bw.history) != 2 {
		t.Errorf("BufferWrapper.history length = %v, want %v", len(bw.history), 2)
	}
}

func TestBufferWrapperClearHistory(t *testing.T) {
	buf := paint.NewBuffer(10, 5)
	bw := NewBufferWrapper(buf, 3)

	bw.SaveSnapshot()
	bw.SaveSnapshot()

	bw.ClearHistory()

	if len(bw.history) != 0 {
		t.Errorf("BufferWrapper.history should be empty after ClearHistory, got length %v", len(bw.history))
	}
}

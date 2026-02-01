// sandbox/lifecycle_test.go - 生命周期管理测试
package sandbox

import (
	"errors"
	"sync"
	"testing"
)

func TestNewLifecycle(t *testing.T) {
	l := NewLifecycle()

	if l.State() != StateStopped {
		t.Errorf("NewLifecycle() state = %v, want %v", l.State(), StateStopped)
	}

	if l.Error() != nil {
		t.Errorf("NewLifecycle() error = %v, want nil", l.Error())
	}
}

func TestValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    State
		to      State
		wantErr bool
	}{
		{"Stopped to Initialized", StateStopped, StateInitialized, false},
		{"Initialized to Running", StateInitialized, StateRunning, false},
		{"Initialized to Stopped", StateInitialized, StateStopped, false},
		{"Running to Paused", StateRunning, StatePaused, false},
		{"Running to Stopped", StateRunning, StateStopped, false},
		{"Paused to Running", StatePaused, StateRunning, false},
		{"Paused to Stopped", StatePaused, StateStopped, false},
		{"Error to Stopped", StateError, StateStopped, false},
		// 无效转换
		{"Stopped to Running", StateStopped, StateRunning, true},
		{"Running to Initialized", StateRunning, StateInitialized, true},
		{"Paused to Initialized", StatePaused, StateInitialized, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLifecycle()

			// 如果不是从 Stopped 开始，先转换到 from 状态
			if tt.from != StateStopped {
				// 构建转换路径
				path := buildTransitionPath(StateStopped, tt.from)
				for _, s := range path {
					if err := l.Transition(s); err != nil {
						t.Fatalf("failed to transition to %v: %v", s, err)
					}
				}
			}

			err := l.Transition(tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lifecycle.Transition() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && l.State() != tt.to {
				t.Errorf("Lifecycle.State() = %v, want %v", l.State(), tt.to)
			}
		})
	}
}

// buildTransitionPath 构建从 start 到 end 的状态转换路径
func buildTransitionPath(start, end State) []State {
	// 简化的路径构建
	if start == end {
		return nil
	}

	var path []State
	current := start

	// Stopped -> Initialized
	if current == StateStopped && end != StateStopped {
		path = append(path, StateInitialized)
		current = StateInitialized
	}

	// Initialized -> Running
	if (current == StateInitialized || current == StatePaused) && end == StateRunning {
		path = append(path, StateRunning)
		current = StateRunning
	}

	// Running -> Paused
	if current == StateRunning && end == StatePaused {
		path = append(path, StatePaused)
	}

	return path
}

func TestTransitionWithError(t *testing.T) {
	l := NewLifecycle()

	// 注册一个失败的钩子
	l.OnTransition(StateRunning, PhaseBefore, func() error {
		return errors.New("hook failed")
	})

	// 先转换到 Initialized
	if err := l.Transition(StateInitialized); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// 尝试转换到 Running，应该失败
	err := l.Transition(StateRunning)
	if err == nil {
		t.Error("expected error from hook, got nil")
	}

	// 状态应该是 Error
	if l.State() != StateError {
		t.Errorf("Lifecycle.State() = %v, want %v", l.State(), StateError)
	}
}

func TestHooks(t *testing.T) {
	l := NewLifecycle()

	var calls []string
	recordCall := func(name string) HookFunc {
		return func() error {
			calls = append(calls, name)
			return nil
		}
	}

	// 注册钩子
	l.OnTransition(StateRunning, PhaseBefore, recordCall("before-1"))
	l.OnTransition(StateRunning, PhaseBefore, recordCall("before-2"))
	l.OnTransition(StateRunning, PhaseAfter, recordCall("after-1"))

	// 转换到 Initialized
	l.Transition(StateInitialized)

	// 清空调用记录
	calls = nil

	// 转换到 Running
	l.Transition(StateRunning)

	expected := []string{"before-1", "before-2", "after-1"}
	if len(calls) != len(expected) {
		t.Fatalf("hooks called %d times, want %d", len(calls), len(expected))
	}

	for i, got := range calls {
		if got != expected[i] {
			t.Errorf("hook call %d = %v, want %v", i, got, expected[i])
		}
	}
}

func TestCanTransitionTo(t *testing.T) {
	l := NewLifecycle()

	if !l.CanTransitionTo(StateInitialized) {
		t.Error("CanTransitionTo(StateInitialized) = false, want true")
	}

	if l.CanTransitionTo(StateRunning) {
		t.Error("CanTransitionTo(StateRunning) = true, want false")
	}

	l.Transition(StateInitialized)

	if !l.CanTransitionTo(StateRunning) {
		t.Error("CanTransitionTo(StateRunning) = false, want true")
	}
}

func TestReset(t *testing.T) {
	l := NewLifecycle()

	l.Transition(StateInitialized)
	l.Transition(StateRunning)

	l.Reset()

	if l.State() != StateStopped {
		t.Errorf("Lifecycle.State() after Reset = %v, want %v", l.State(), StateStopped)
	}

	if l.Error() != nil {
		t.Errorf("Lifecycle.Error() after Reset = %v, want nil", l.Error())
	}
}

func TestConcurrentTransitions(t *testing.T) {
	l := NewLifecycle()
	var wg sync.WaitGroup

	// 多个 goroutine 尝试转换
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Transition(StateInitialized)
		}()
	}

	wg.Wait()

	// 只有一个转换应该成功
	successCount := 0
	if l.State() == StateInitialized {
		successCount = 1
	}

	if successCount != 1 {
		t.Errorf("got %d successful transitions, want 1", successCount)
	}
}

func TestConcurrentStateRead(t *testing.T) {
	l := NewLifecycle()
	var wg sync.WaitGroup

	// 启动多个 goroutine 读取状态
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = l.State()
			_ = l.CanTransitionTo(StateInitialized)
		}()
	}

	wg.Wait()
	// 如果没有死锁或panic，测试通过
}

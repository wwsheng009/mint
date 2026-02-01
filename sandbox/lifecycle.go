// sandbox/lifecycle.go - 生命周期状态机
package sandbox

import (
	"fmt"
	"sync"
)

// validTransitions 合法状态转换表
var validTransitions = map[State][]State{
	StateStopped:     {StateInitialized},
	StateInitialized: {StateRunning, StateStopped},
	StateRunning:     {StatePaused, StateStopped},
	StatePaused:      {StateRunning, StateStopped},
	StateError:       {StateStopped},
}

// Lifecycle 生命周期管理器
type Lifecycle struct {
	mu     sync.RWMutex
	state  State
	err    error
	hooks  map[HookKey][]HookFunc
}

// HookFunc 钩子函数类型
type HookFunc func() error

// NewLifecycle 创建生命周期管理器
func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		state: StateStopped,
		hooks: make(map[HookKey][]HookFunc),
	}
}

// State 获取当前状态
func (l *Lifecycle) State() State {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}

// Error 获取错误状态
func (l *Lifecycle) Error() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.err
}

// Transition 执行状态转移
func (l *Lifecycle) Transition(to State) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	from := l.state

	// 验证状态转移是否合法
	if !l.isValidTransition(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}

	// 执行前置钩子
	if err := l.executeHooks(HookKey{to, PhaseBefore}); err != nil {
		l.state = StateError
		l.err = err
		return err
	}

	// 更新状态
	l.state = to

	// 执行后置钩子
	if err := l.executeHooks(HookKey{to, PhaseAfter}); err != nil {
		l.state = StateError
		l.err = err
		return err
	}

	return nil
}

// isValidTransition 检查状态转移是否合法
func (l *Lifecycle) isValidTransition(from, to State) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// executeHooks 执行指定钩子
func (l *Lifecycle) executeHooks(key HookKey) error {
	hooks, ok := l.hooks[key]
	if !ok {
		return nil
	}
	for _, hook := range hooks {
		if err := hook(); err != nil {
			return err
		}
	}
	return nil
}

// OnTransition 注册状态转移钩子
func (l *Lifecycle) OnTransition(state State, phase Phase, fn HookFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := HookKey{state, phase}
	l.hooks[key] = append(l.hooks[key], fn)
}

// Reset 重置生命周期
func (l *Lifecycle) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.state = StateStopped
	l.err = nil
}

// CanTransitionTo 检查是否可以转移到目标状态
func (l *Lifecycle) CanTransitionTo(to State) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.isValidTransition(l.state, to)
}

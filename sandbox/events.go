// sandbox/events.go - 事件注入系统
package sandbox

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/platform"
)

// EventInjector 事件注入器
type EventInjector struct {
	mu       sync.RWMutex
	strategy InjectionStrategy
	handler  EventHandler
	recorder *EventRecorder
}

// NewEventInjector 创建事件注入器
func NewEventInjector(strategy InjectionStrategy) *EventInjector {
	return &EventInjector{
		strategy: strategy,
	}
}

// SetHandler 设置事件处理器
func (ei *EventInjector) SetHandler(handler EventHandler) {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	ei.handler = handler
}

// SetRecorder 设置事件录制器
func (ei *EventInjector) SetRecorder(recorder *EventRecorder) {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	ei.recorder = recorder
}

// Strategy 获取当前策略
func (ei *EventInjector) Strategy() InjectionStrategy {
	ei.mu.RLock()
	defer ei.mu.RUnlock()
	return ei.strategy
}

// SetStrategy 动态切换策略
func (ei *EventInjector) SetStrategy(strategy InjectionStrategy) {
	ei.mu.Lock()
	defer ei.mu.Unlock()
	ei.strategy = strategy
}

// Inject 注入事件 (根据策略)
func (ei *EventInjector) Inject(event platform.RawInput) error {
	ei.mu.RLock()
	strategy := ei.strategy
	handler := ei.handler
	recorder := ei.recorder
	ei.mu.RUnlock()

	switch strategy {
	case InjectProhibited:
		return ei.injectProhibited(event, recorder)

	case InjectAllowed:
		return ei.injectAllowed(event, handler, recorder)

	case InjectRecorded:
		return ei.injectRecorded(event, recorder)

	default:
		return ErrInvalidStrategy
	}
}

func (ei *EventInjector) injectProhibited(event platform.RawInput, recorder *EventRecorder) error {
	// 真实环境：仅记录，不注入
	if recorder != nil {
		recorder.Record(event)
	}
	return ErrInjectionNotAllowed
}

func (ei *EventInjector) injectAllowed(event platform.RawInput, handler EventHandler, recorder *EventRecorder) error {
	// 测试环境：记录并注入
	if recorder != nil {
		recorder.Record(event)
	}
	if handler != nil {
		return handler(event)
	}
	return nil
}

func (ei *EventInjector) injectRecorded(event platform.RawInput, recorder *EventRecorder) error {
	// 录制模式：只记录不注入
	if recorder != nil {
		return recorder.Record(event)
	}
	return nil
}

// EventRecorder 事件录制器
type EventRecorder struct {
	mu     sync.Mutex
	events []platform.RawInput
	maxLen int
}

// NewEventRecorder 创建事件录制器
func NewEventRecorder(maxLen int) *EventRecorder {
	if maxLen <= 0 {
		maxLen = 10000
	}
	return &EventRecorder{
		events: make([]platform.RawInput, 0, maxLen),
		maxLen: maxLen,
	}
}

// Record 录制事件
func (r *EventRecorder) Record(event platform.RawInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.events) >= r.maxLen {
		// 淘汰最旧的
		r.events = r.events[1:]
	}
	r.events = append(r.events, event)
	return nil
}

// Events 获取所有录制的事件
func (r *EventRecorder) Events() []platform.RawInput {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]platform.RawInput, len(r.events))
	copy(result, r.events)
	return result
}

// Clear 清空录制
func (r *EventRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = r.events[:0]
}

// Len 返回事件数量
func (r *EventRecorder) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

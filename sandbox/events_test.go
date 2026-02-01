// sandbox/events_test.go - 事件注入系统测试
package sandbox

import (
	"sync/atomic"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
)

func TestNewEventInjector(t *testing.T) {
	ei := NewEventInjector(InjectAllowed)

	if ei.Strategy() != InjectAllowed {
		t.Errorf("NewEventInjector() strategy = %v, want %v", ei.Strategy(), InjectAllowed)
	}
}

func TestSetStrategy(t *testing.T) {
	ei := NewEventInjector(InjectProhibited)

	ei.SetStrategy(InjectAllowed)

	if ei.Strategy() != InjectAllowed {
		t.Errorf("SetStrategy() strategy = %v, want %v", ei.Strategy(), InjectAllowed)
	}
}

func TestSetHandler(t *testing.T) {
	ei := NewEventInjector(InjectAllowed)
	called := false

	handler := func(event platform.RawInput) error {
		called = true
		return nil
	}

	ei.SetHandler(handler)

	// 注入事件
	event := platform.RawInput{Type: platform.InputKeyPress}
	ei.Inject(event)

	if !called {
		t.Error("SetHandler() handler was not called")
	}
}

func TestInjectProhibited(t *testing.T) {
	ei := NewEventInjector(InjectProhibited)
	recorder := NewEventRecorder(100)
	ei.SetRecorder(recorder)

	event := platform.RawInput{Type: platform.InputKeyPress}
	err := ei.Inject(event)

	if err != ErrInjectionNotAllowed {
		t.Errorf("Inject() error = %v, want %v", err, ErrInjectionNotAllowed)
	}

	if recorder.Len() != 1 {
		t.Errorf("Inject() recorded %d events, want 1", recorder.Len())
	}
}

func TestInjectAllowed(t *testing.T) {
	ei := NewEventInjector(InjectAllowed)
	recorder := NewEventRecorder(100)
	ei.SetRecorder(recorder)

	var called bool
	handler := func(event platform.RawInput) error {
		called = true
		return nil
	}
	ei.SetHandler(handler)

	event := platform.RawInput{Type: platform.InputKeyPress}
	err := ei.Inject(event)

	if err != nil {
		t.Errorf("Inject() error = %v, want nil", err)
	}

	if !called {
		t.Error("Inject() handler was not called")
	}

	if recorder.Len() != 1 {
		t.Errorf("Inject() recorded %d events, want 1", recorder.Len())
	}
}

func TestInjectRecorded(t *testing.T) {
	ei := NewEventInjector(InjectRecorded)
	recorder := NewEventRecorder(100)
	ei.SetRecorder(recorder)

	var called bool
	handler := func(event platform.RawInput) error {
		called = true
		return nil
	}
	ei.SetHandler(handler)

	event := platform.RawInput{Type: platform.InputKeyPress}
	err := ei.Inject(event)

	if err != nil {
		t.Errorf("Inject() error = %v, want nil", err)
	}

	if called {
		t.Error("Inject() handler should not be called in Recorded mode")
	}

	if recorder.Len() != 1 {
		t.Errorf("Inject() recorded %d events, want 1", recorder.Len())
	}
}

func TestNewEventRecorder(t *testing.T) {
	r := NewEventRecorder(100)

	if r.maxLen != 100 {
		t.Errorf("NewEventRecorder() maxLen = %v, want 100", r.maxLen)
	}

	if r.Len() != 0 {
		t.Errorf("NewEventRecorder() Len() = %v, want 0", r.Len())
	}
}

func TestEventRecorderRecord(t *testing.T) {
	r := NewEventRecorder(10)

	for i := 0; i < 5; i++ {
		event := platform.RawInput{Type: platform.InputKeyPress}
		r.Record(event)
	}

	if r.Len() != 5 {
		t.Errorf("Record() Len() = %v, want 5", r.Len())
	}
}

func TestEventRecorderMaxLen(t *testing.T) {
	r := NewEventRecorder(3)

	// 记录超过最大长度的事件
	for i := 0; i < 10; i++ {
		event := platform.RawInput{Type: platform.InputKeyPress}
		r.Record(event)
	}

	// 应该只保留最后3个
	if r.Len() != 3 {
		t.Errorf("Record() Len() = %v, want 3", r.Len())
	}

	events := r.Events()
	if len(events) != 3 {
		t.Errorf("Events() length = %v, want 3", len(events))
	}
}

func TestEventRecorderEvents(t *testing.T) {
	r := NewEventRecorder(10)

	events := []platform.RawInput{
		{Type: platform.InputKeyPress, Key: 'a'},
		{Type: platform.InputKeyPress, Key: 'b'},
		{Type: platform.InputKeyPress, Key: 'c'},
	}

	for _, e := range events {
		r.Record(e)
	}

	retrieved := r.Events()

	if len(retrieved) != len(events) {
		t.Fatalf("Events() length = %v, want %v", len(retrieved), len(events))
	}

	for i, e := range retrieved {
		if e.Key != events[i].Key {
			t.Errorf("Events()[%d] Key = %v, want %v", i, e.Key, events[i].Key)
		}
	}
}

func TestEventRecorderClear(t *testing.T) {
	r := NewEventRecorder(10)

	for i := 0; i < 5; i++ {
		event := platform.RawInput{Type: platform.InputKeyPress}
		r.Record(event)
	}

	r.Clear()

	if r.Len() != 0 {
		t.Errorf("Clear() Len() = %v, want 0", r.Len())
	}
}

func TestEventRecorderConcurrent(t *testing.T) {
	r := NewEventRecorder(1000)
	var ops int32

	// 并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				event := platform.RawInput{Type: platform.InputKeyPress}
				r.Record(event)
				atomic.AddInt32(&ops, 1)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if ops != 1000 {
		t.Errorf("Concurrent Record() performed %d operations, want 1000", ops)
	}

	// 并发读取
	for i := 0; i < 10; i++ {
		go func() {
			_ = r.Events()
			_ = r.Len()
		}()
	}
}

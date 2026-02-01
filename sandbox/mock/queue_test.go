// mock/queue_test.go - 有界事件队列测试
package mock

import (
	"sync"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

func TestNewBoundedQueue(t *testing.T) {
	config := QueueConfig{MaxSize: 100}
	q := NewBoundedQueue(config)

	if q.config.MaxSize != 100 {
		t.Errorf("NewBoundedQueue() MaxSize = %v, want 100", q.config.MaxSize)
	}

	if !q.IsEmpty() {
		t.Error("NewBoundedQueue() should be empty")
	}
}

func TestNewBoundedQueueDefaultConfig(t *testing.T) {
	config := DefaultQueueConfig()
	q := NewBoundedQueue(config)

	if q.config.MaxSize != 10000 {
		t.Errorf("DefaultQueueConfig() MaxSize = %v, want 10000", q.config.MaxSize)
	}

	if q.config.MaxMemory != 100*1024*1024 {
		t.Errorf("DefaultQueueConfig() MaxMemory = %v, want 100MB", q.config.MaxMemory)
	}
}

func TestQueuePushPop(t *testing.T) {
	q := NewBoundedQueue(DefaultQueueConfig())

	event := platform.RawInput{
		Type: platform.InputKeyPress,
		Key:  'a',
	}

	err := q.Push(event)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if q.Len() != 1 {
		t.Errorf("Push() Len() = %v, want 1", q.Len())
	}

	popped, err := q.Pop()
	if err != nil {
		t.Fatalf("Pop() error = %v", err)
	}

	if popped.Key != 'a' {
		t.Errorf("Pop() Key = %v, want 'a'", popped.Key)
	}

	if q.Len() != 0 {
		t.Errorf("Pop() Len() = %v, want 0", q.Len())
	}
}

func TestQueuePeek(t *testing.T) {
	q := NewBoundedQueue(DefaultQueueConfig())

	event := platform.RawInput{
		Type: platform.InputKeyPress,
		Key:  'x',
	}

	q.Push(event)

	peeked, err := q.Peek()
	if err != nil {
		t.Fatalf("Peek() error = %v", err)
	}

	if peeked.Key != 'x' {
		t.Errorf("Peek() Key = %v, want 'x'", peeked.Key)
	}

	// Peek 不应移除事件
	if q.Len() != 1 {
		t.Errorf("Peek() Len() = %v, want 1", q.Len())
	}
}

func TestQueueCapacityLimit(t *testing.T) {
	config := QueueConfig{MaxSize: 5}
	q := NewBoundedQueue(config)

	// 添加超过容量的元素
	for i := 0; i < 10; i++ {
		event := platform.RawInput{
			Type: platform.InputKeyPress,
			Key:  rune('a' + i),
		}
		q.Push(event)
	}

	// 应该只保留最后5个
	if q.Len() != 5 {
		t.Errorf("Capacity limit Len() = %v, want 5", q.Len())
	}

	// 验证是最后5个元素
	for i := 0; i < 5; i++ {
		event, _ := q.Pop()
		expected := rune('a' + 5 + i)
		if event.Key != expected {
			t.Errorf("After eviction Key = %v, want %v", event.Key, expected)
		}
	}
}

func TestQueueMemoryLimit(t *testing.T) {
	config := QueueConfig{
		MaxSize:   1000,
		MaxMemory: 200, // 非常小的限制
	}
	q := NewBoundedQueue(config)

	// 添加大事件
	for i := 0; i < 10; i++ {
		event := platform.RawInput{
			Type: platform.InputPaste,
			Data: make([]byte, 50), // 50 bytes data
		}
		q.Push(event)
	}

	// 由于内存限制，队列应该自动淘汰
	stats := q.Stats()
	if stats.MemoryUsed > stats.MemoryLimit {
		// 允许稍微超过（当前事件的大小）
		if stats.MemoryUsed > stats.MemoryLimit+100 {
			t.Errorf("Memory limit exceeded: %d > %d", stats.MemoryUsed, stats.MemoryLimit)
		}
	}
}

func TestQueueClear(t *testing.T) {
	q := NewBoundedQueue(DefaultQueueConfig())

	for i := 0; i < 10; i++ {
		event := platform.RawInput{Type: platform.InputKeyPress}
		q.Push(event)
	}

	q.Clear()

	if q.Len() != 0 {
		t.Errorf("Clear() Len() = %v, want 0", q.Len())
	}

	if !q.IsEmpty() {
		t.Error("Clear() IsEmpty() = false, want true")
	}
}

func TestQueueStats(t *testing.T) {
	config := QueueConfig{MaxSize: 100, MaxMemory: 1000}
	q := NewBoundedQueue(config)

	event := platform.RawInput{Type: platform.InputKeyPress, Key: 'a'}
	q.Push(event)

	stats := q.Stats()

	if stats.Length != 1 {
		t.Errorf("Stats() Length = %v, want 1", stats.Length)
	}

	if stats.MemoryLimit != 1000 {
		t.Errorf("Stats() MemoryLimit = %v, want 1000", stats.MemoryLimit)
	}

	if stats.MemoryUsed <= 0 {
		t.Error("Stats() MemoryUsed should be > 0")
	}
}

func TestQueueEvictCount(t *testing.T) {
	config := QueueConfig{MaxSize: 3}
	q := NewBoundedQueue(config)

	for i := 0; i < 10; i++ {
		event := platform.RawInput{Type: platform.InputKeyPress}
		q.Push(event)
	}

	stats := q.Stats()
	if stats.EvictCount != 7 {
		t.Errorf("Stats() EvictCount = %v, want 7", stats.EvictCount)
	}
}

func TestQueuePopEmpty(t *testing.T) {
	q := NewBoundedQueue(DefaultQueueConfig())

	_, err := q.Pop()
	if err != sandbox.ErrQueueEmpty {
		t.Errorf("Pop() error = %v, want %v", err, sandbox.ErrQueueEmpty)
	}
}

func TestQueuePeekEmpty(t *testing.T) {
	q := NewBoundedQueue(DefaultQueueConfig())

	_, err := q.Peek()
	if err != sandbox.ErrQueueEmpty {
		t.Errorf("Peek() error = %v, want %v", err, sandbox.ErrQueueEmpty)
	}
}

func TestQueueConcurrent(t *testing.T) {
	q := NewBoundedQueue(DefaultQueueConfig())
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				event := platform.RawInput{
					Type: platform.InputKeyPress,
					Key:  rune('a' + j),
				}
				q.Push(event)
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				q.Pop()
				_ = q.Len()
				_ = q.IsEmpty()
				_ = q.Stats()
			}
		}()
	}

	wg.Wait()

	// 只验证没有死锁或panic
	finalLen := q.Len()
	t.Logf("After concurrent operations, queue length: %d", finalLen)
}

func TestEstimateEventSize(t *testing.T) {
	event1 := platform.RawInput{Type: platform.InputKeyPress}
	size1 := estimateEventSize(event1)

	if size1 <= 0 {
		t.Errorf("estimateEventSize() = %v, want > 0", size1)
	}

	event2 := platform.RawInput{
		Type: platform.InputPaste,
		Data: make([]byte, 100),
	}
	size2 := estimateEventSize(event2)

	if size2 <= size1 {
		t.Errorf("estimateEventSize() with data should be larger, got %v <= %v", size2, size1)
	}
}

func TestMin(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5, 10) = 5")
	}

	if min(10, 5) != 5 {
		t.Error("min(10, 5) = 5")
	}

	if min(5, 5) != 5 {
		t.Error("min(5, 5) = 5")
	}
}

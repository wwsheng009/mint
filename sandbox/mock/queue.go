// mock/queue.go - 有界事件队列
package mock

import (
	"sync"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

// QueueConfig 队列配置
type QueueConfig struct {
	MaxSize     int               // 最大队列长度 (默认 10000)
	MaxMemory   int64             // 最大内存占用 (字节，默认 100MB)
	EvictPolicy sandbox.EvictPolicy
}

// DefaultQueueConfig 默认配置
func DefaultQueueConfig() QueueConfig {
	return QueueConfig{
		MaxSize:     10000,
		MaxMemory:   100 * 1024 * 1024, // 100MB
		EvictPolicy: sandbox.EvictOldest,
	}
}

// BoundedQueue 有界事件队列
type BoundedQueue struct {
	mu         sync.RWMutex
	config     QueueConfig
	events     []platform.RawInput
	memory     int64
	evictCount int64 // 已淘汰事件数
}

// NewBoundedQueue 创建有界队列
func NewBoundedQueue(config QueueConfig) *BoundedQueue {
	if config.MaxSize <= 0 {
		config.MaxSize = 10000
	}
	return &BoundedQueue{
		config: config,
		events: make([]platform.RawInput, 0, min(config.MaxSize, 1000)),
	}
}

// Push 添加事件
func (q *BoundedQueue) Push(event platform.RawInput) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	eventSize := estimateEventSize(event)

	// 检查内存限制
	for q.config.MaxMemory > 0 && q.memory+eventSize > q.config.MaxMemory && len(q.events) > 0 {
		if err := q.evictOne(); err != nil {
			return err
		}
	}

	// 检查容量限制
	for len(q.events) >= q.config.MaxSize {
		if err := q.evictOne(); err != nil {
			return err
		}
	}

	q.events = append(q.events, event)
	q.memory += eventSize

	return nil
}

// Pop 取出最旧的事件
func (q *BoundedQueue) Pop() (platform.RawInput, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.events) == 0 {
		return platform.RawInput{}, sandbox.ErrQueueEmpty
	}

	event := q.events[0]
	q.events = q.events[1:]
	q.memory -= estimateEventSize(event)

	return event, nil
}

// Peek 查看最旧的事件 (不移除)
func (q *BoundedQueue) Peek() (platform.RawInput, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.events) == 0 {
		return platform.RawInput{}, sandbox.ErrQueueEmpty
	}

	return q.events[0], nil
}

// Len 返回队列长度
func (q *BoundedQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.events)
}

// IsEmpty 检查队列是否为空
func (q *BoundedQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.events) == 0
}

// Clear 清空队列
func (q *BoundedQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.events = q.events[:0]
	q.memory = 0
}

// evictOne 淘汰一个事件
func (q *BoundedQueue) evictOne() error {
	if len(q.events) == 0 {
		return nil
	}

	event := q.events[0]
	q.events = q.events[1:]
	q.memory -= estimateEventSize(event)
	q.evictCount++

	return nil
}

// QueueStats 队列统计
type QueueStats struct {
	Length      int   // 当前队列长度
	MaxQueueSize int  // 队列最大长度
	MemoryUsed  int64 // 已用内存
	MemoryLimit int64 // 内存限制
	EvictCount  int64 // 淘汰事件数
}

// Stats 获取队列统计
func (q *BoundedQueue) Stats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return QueueStats{
		Length:       len(q.events),
		MaxQueueSize: q.config.MaxSize,
		MemoryUsed:   q.memory,
		MemoryLimit:  q.config.MaxMemory,
		EvictCount:   q.evictCount,
	}
}

// estimateEventSize 估算事件内存占用
func estimateEventSize(event platform.RawInput) int64 {
	// 基础结构大小
	size := int64(64) // RawInput 结构体大小估算

	// Data 字段
	if event.Data != nil {
		size += int64(len(event.Data))
	}

	return size
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package scheduler

import (
	"sync"
	"time"

	runtimeevent "github.com/wwsheng009/mint/runtime/event"
)

// MouseThrottler 节流鼠标事件以减少处理频率
//
// MouseThrottler 使用两种策略来减少鼠标事件的处理：
// 1. 时间间隔：两次事件之间的最小时间间隔
// 2. 像素阈值：鼠标移动的最小像素距离
//
// 这样可以避免过多的 MouseMove 事件导致性能问题。
type MouseThrottler struct {
	mu sync.Mutex

	// 配置
	minInterval    time.Duration // 最小时间间隔
	pixelThreshold int           // 像素阈值

	// 状态
	lastX      int
	lastY      int
	lastTime   time.Time
	lastEvent  *runtimeevent.MouseEvent
	pending    bool
}

// NewMouseThrottler 创建新的鼠标节流器
//
// interval: 最小时间间隔（例如 16ms for 60Hz）
// pixelThreshold: 像素阈值（例如 2 像素）
func NewMouseThrottler(interval time.Duration, pixelThreshold int) *MouseThrottler {
	return &MouseThrottler{
		minInterval:    interval,
		pixelThreshold: pixelThreshold,
		lastTime:       time.Time{},
	}
}

// NewMouseThrottler60Hz 创建一个 60Hz 的鼠标节流器
//
// 60Hz 意味着每秒最多处理 60 个鼠标事件，间隔约为 16.67ms
func NewMouseThrottler60Hz() *MouseThrottler {
	return NewMouseThrottler(16*time.Millisecond, 2)
}

// Throttle 节流鼠标事件
//
// 如果事件应该被处理，返回 true。
// 如果事件应该被丢弃，返回 false。
func (mt *MouseThrottler) Throttle(event *runtimeevent.MouseEvent) bool {
	if event == nil {
		return false
	}

	mt.mu.Lock()
	defer mt.mu.Unlock()

	now := time.Now()

	// 检查时间间隔
	if !mt.lastTime.IsZero() && now.Sub(mt.lastTime) < mt.minInterval {
		// 时间间隔太短，保存待处理事件
		mt.lastEvent = event
		mt.pending = true
		return false
	}

	// 检查像素阈值（仅对 MouseMove 有效）
	if event.Type == "move" {
		if !mt.shouldProcessMove(event.X, event.Y) {
			return false
		}
	}

	// 更新状态
	mt.lastX = event.X
	mt.lastY = event.Y
	mt.lastTime = now
	mt.lastEvent = nil
	mt.pending = false

	return true
}

// shouldProcessMove 检查是否应该处理鼠标移动事件
func (mt *MouseThrottler) shouldProcessMove(x, y int) bool {
	// 如果是第一次或者移动距离超过阈值，则处理
	dx := x - mt.lastX
	dy := y - mt.lastY

	// 计算移动距离（平方）
	distanceSq := dx*dx + dy*dy
	thresholdSq := mt.pixelThreshold * mt.pixelThreshold

	return distanceSq >= thresholdSq
}

// GetPending 获取待处理的事件
//
// 如果有待处理的事件（因为时间间隔被延迟），返回该事件。
// 否则返回 nil。
func (mt *MouseThrottler) GetPending() *runtimeevent.MouseEvent {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if !mt.pending {
		return nil
	}

	// 检查是否已经过了最小间隔
	if time.Since(mt.lastTime) >= mt.minInterval {
		event := mt.lastEvent
		mt.lastEvent = nil
		mt.pending = false
		return event
	}

	return nil
}

// Reset 重置节流器状态
//
// 当焦点切换或组件隐藏时，应该重置节流器。
func (mt *MouseThrottler) Reset() {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.lastX = 0
	mt.lastY = 0
	mt.lastTime = time.Time{}
	mt.lastEvent = nil
	mt.pending = false
}

// SetMinInterval 设置最小时间间隔
func (mt *MouseThrottler) SetMinInterval(interval time.Duration) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.minInterval = interval
}

// SetPixelThreshold 设置像素阈值
func (mt *MouseThrottler) SetPixelThreshold(threshold int) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.pixelThreshold = threshold
}

// GetStats 获取节流器统计信息
func (mt *MouseThrottler) GetStats() map[string]interface{} {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	stats := map[string]interface{}{
		"min_interval":     mt.minInterval.String(),
		"pixel_threshold":  mt.pixelThreshold,
		"last_x":          mt.lastX,
		"last_y":          mt.lastY,
		"has_pending":     mt.pending,
		"time_since_last": time.Since(mt.lastTime).String(),
	}

	return stats
}

// ============================================================================
// 全局鼠标节流器
// ============================================================================

var (
	globalMouseThrottler *MouseThrottler
	globalThrottlerOnce  sync.Once
)

// GetGlobalMouseThrottler 获取全局鼠标节流器
//
// 全局节流器是一个 60Hz 的节流器，用于整个应用。
func GetGlobalMouseThrottler() *MouseThrottler {
	globalThrottlerOnce.Do(func() {
		globalMouseThrottler = NewMouseThrottler60Hz()
	})

	return globalMouseThrottler
}

// ThrottleMouseEvent 使用全局节流器节流鼠标事件
//
// 这是一个便捷函数，用于快速节流鼠标事件。
func ThrottleMouseEvent(event *runtimeevent.MouseEvent) bool {
	throttler := GetGlobalMouseThrottler()
	return throttler.Throttle(event)
}

// ProcessPendingMouseEvents 处理待处理的鼠标事件
//
// 应该在主循环中定期调用此函数，以确保被延迟的事件得到处理。
func ProcessPendingMouseEvents(handler func(*runtimeevent.MouseEvent)) {
	throttler := GetGlobalMouseThrottler()

	for {
		event := throttler.GetPending()
		if event == nil {
			break
		}

		handler(event)

		// 让出 CPU，避免阻塞
		time.Sleep(time.Millisecond)
	}
}

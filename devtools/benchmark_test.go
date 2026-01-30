// Package devtools benchmarks for DevTools.
package devtools_test

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/devtools"
)

// BenchmarkDevTools_Disabled 测试禁用 DevTools 的性能开销
func BenchmarkDevTools_Disabled(b *testing.B) {
	dt := devtools.New()
	// 不启用 DevTools

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt.BeginFrame()
		dt.RecordEvent("keypress", "node", "bubble", nil)
		dt.EndFrame()
	}
}

// BenchmarkDevTools_Enabled 测试启用 DevTools 的性能开销
func BenchmarkDevTools_Enabled(b *testing.B) {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt.BeginFrame()
		dt.RecordEvent("keypress", "node", "bubble", nil)
		dt.EndFrame()
	}
}

// BenchmarkDevTools_WithLayout 测试包含布局收集的性能
func BenchmarkDevTools_WithLayout(b *testing.B) {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt.BeginFrame()
		dt.CollectLayout(nil) // nil for now, uses adapter
		dt.EndFrame()
	}
}

// BenchmarkEventBus_Emit 测试 EventBus 发送性能
func BenchmarkEventBus_Emit(b *testing.B) {
	bus := devtools.NewEventBus(4096)
	bus.Enable()
	defer bus.Close()

	ev := devtools.DebugEvent{Type: devtools.EventLayout}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Emit(ev)
		}
	})
}

// BenchmarkEventBus_Subscribe 测试 EventBus 订阅性能
func BenchmarkEventBus_Subscribe(b *testing.B) {
	bus := devtools.NewEventBus(4096)
	bus.Enable()
	defer bus.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan devtools.DebugEvent, 100)
		unsub := bus.Subscribe(ch)
		unsub()
	}
}

// BenchmarkFrameTimeline_AddFrame 测试 FrameTimeline 添加帧性能
func BenchmarkFrameTimeline_AddFrame(b *testing.B) {
	ft := devtools.NewFrameTimeline()
	ft.Enable()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := ft.BeginFrame(devtools.FrameID(i))
		ft.EndFrame()
		_ = entry
	}
}

// BenchmarkFrameTimeline_GetAllFrames 测试获取所有帧性能
func BenchmarkFrameTimeline_GetAllFrames(b *testing.B) {
	ft := devtools.NewFrameTimeline()
	ft.Enable()

	// Pre-fill with frames
	for i := 0; i < 100; i++ {
		ft.BeginFrame(devtools.FrameID(i))
		ft.EndFrame()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ft.GetAllFrames()
	}
}

// BenchmarkCausalGraph_New 测试 CausalGraph 创建性能（对象池 vs 直接分配）
func BenchmarkCausalGraph_New(b *testing.B) {
	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cg := devtools.NewCausalGraph(devtools.FrameID(i))
			cg.AddEvent("test", "node", "bubble")
			// 不释放，让 GC 处理
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cg := devtools.NewCausalGraph(devtools.FrameID(i))
			cg.AddEvent("test", "node", "bubble")
			cg.Release()
		}
	})
}

// BenchmarkCausalGraph_AddEvent 测试添加事件性能
func BenchmarkCausalGraph_AddEvent(b *testing.B) {
	cg := devtools.NewCausalGraph(devtools.FrameID(1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.AddEvent("keypress", "input1", "bubble")
	}
}

// BenchmarkLogger_Log 测试日志记录性能
func BenchmarkLogger_Log(b *testing.B) {
	logger := devtools.NewLogger(1024)
	logger.Enable()
	logger.SetLevel(devtools.LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message: %d", i)
	}
}

// BenchmarkLogger_Disabled 测试禁用日志的性能
func BenchmarkLogger_Disabled(b *testing.B) {
	logger := devtools.NewLogger(1024)
	// 不启用

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message: %d", i)
	}
}

// BenchmarkDevTools_Concurrent 测试并发访问性能
func BenchmarkDevTools_Concurrent(b *testing.B) {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dt.BeginFrame()
			dt.RecordEvent("test", "node", "bubble", nil)
			dt.EndFrame()
		}
	})
}

// BenchmarkDevTools_FullCycle 测试完整的 DevTools 周期
func BenchmarkDevTools_FullCycle(b *testing.B) {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	bus := dt.GetEventBus()
	ch := make(chan devtools.DebugEvent, 100)
	unsub := bus.Subscribe(ch)
	defer unsub()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt.BeginFrame()
		dt.RecordEvent("keypress", "input1", "bubble", nil)
		bus.Emit(devtools.DebugEvent{Type: devtools.EventLayout})
		dt.EndFrame()
	}
}

// BenchmarkDevTools_MemoryAllocation 测试内存分配
func BenchmarkDevTools_MemoryAllocation(b *testing.B) {
	dt := devtools.New()
	dt.Enable()
	defer dt.Shutdown()

	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dt.BeginFrame()
		dt.RecordEvent("keypress", "node", "bubble", nil)
		dt.EndFrame()
	}
}

// BenchmarkComparison_WithVsWithoutDevTools 对比测试
func BenchmarkComparison_WithVsWithoutDevTools(b *testing.B) {
	b.Run("WithoutDevTools", func(b *testing.B) {
		dt := devtools.New()
		// 不启用

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dt.BeginFrame()
			dt.RecordEvent("test", "node", "bubble", nil)
			dt.EndFrame()
		}
	})

	b.Run("WithDevTools", func(b *testing.B) {
		dt := devtools.New()
		dt.Enable()
		defer dt.Shutdown()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dt.BeginFrame()
			dt.RecordEvent("test", "node", "bubble", nil)
			dt.EndFrame()
		}
	})
}

// BenchmarkEventBus_RingBuffer 测试不同大小的环形缓冲区
func BenchmarkEventBus_RingBuffer(b *testing.B) {
	sizes := []int{64, 256, 1024, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			bus := devtools.NewEventBus(size)
			bus.Enable()
			defer bus.Close()

			ev := devtools.DebugEvent{Type: devtools.EventLayout}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bus.Emit(ev)
			}
		})
	}
}

// BenchmarkFrameTimeline_Capacity 测试不同容量的 FrameTimeline
func BenchmarkFrameTimeline_Capacity(b *testing.B) {
	capacities := []int{32, 64, 128, 256}

	for _, cap := range capacities {
		b.Run(fmt.Sprintf("Capacity_%d", cap), func(b *testing.B) {
			ft := devtools.NewFrameTimelineWithCapacity(cap)
			ft.Enable()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ft.BeginFrame(devtools.FrameID(i))
				ft.EndFrame()
			}
		})
	}
}

// BenchmarkEventBus_Stats 测试获取统计信息性能
func BenchmarkEventBus_Stats(b *testing.B) {
	bus := devtools.NewEventBus(4096)
	bus.Enable()
	defer bus.Close()

	// Pre-populate with events
	for i := 0; i < 1000; i++ {
		bus.Emit(devtools.DebugEvent{Type: devtools.EventLayout})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bus.GetStats()
	}
}

// 04_queue_stats/demo_test.go
// 队列统计演示测试

package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox/mock"
	"github.com/wwsheng009/mint/ui"
)

// TestQueueBasicStats 演示基本的队列统计
func TestQueueBasicStats(t *testing.T) {
	t.Log("=== 基本队列统计演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 获取初始统计
	stats := sb.QueueStats()
	t.Log("\n--- 初始队列状态 ---")
	t.Logf("队列长度: %d / %d", stats.Length, stats.MemoryLimit)
	t.Logf("内存使用: %d / %d 字节", stats.MemoryUsed, stats.MemoryLimit)
	t.Logf("淘汰事件数: %d", stats.EvictCount)
}

// TestQueueAfterEvents 演示事件注入后的统计
func TestQueueAfterEvents(t *testing.T) {
	t.Log("=== 事件注入后的队列统计 ===")

	testApp, err := ui.RunTestWithSandbox(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 获取初始统计
	initialStats := sb.QueueStats()
	t.Logf("\n--- 初始状态 ---")
	t.Logf("队列长度: %d", initialStats.Length)

	// 注入一些事件
	t.Log("\n--- 注入事件 ---")
	eventCount := 10
	for i := 0; i < eventCount; i++ {
		sb.InjectKey('a')
	}

	// 等待处理
	time.Sleep(100 * time.Millisecond)

	// 获取新的统计
	newStats := sb.QueueStats()
	t.Logf("\n--- 注入 %d 个事件后 ---", eventCount)
	t.Logf("队列长度: %d (变化: %+d)", newStats.Length, newStats.Length-initialStats.Length)
	t.Logf("内存使用: %d 字节", newStats.MemoryUsed)
}

// TestQueueMemoryMonitoring 演示内存监控
func TestQueueMemoryMonitoring(t *testing.T) {
	t.Log("=== 队列内存监控演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 监控内存使用
	t.Log("\n--- 内存监控 ---")

	initialStats := sb.QueueStats()
	t.Logf("初始内存: %d 字节", initialStats.MemoryUsed)

	// 注入大量事件
	largeCount := 100
	for i := 0; i < largeCount; i++ {
		sb.InjectSpecialKey(platform.KeyTab)
	}

	time.Sleep(50 * time.Millisecond)

	afterStats := sb.QueueStats()
	t.Logf("注入 %d 个事件后:", largeCount)
	t.Logf("内存使用: %d 字节", afterStats.MemoryUsed)
	t.Logf("内存增长: %+d 字节", afterStats.MemoryUsed-initialStats.MemoryUsed)

	// 计算平均每事件内存
	if afterStats.Length > initialStats.Length {
		avgMemory := (afterStats.MemoryUsed - initialStats.MemoryUsed) / int64(afterStats.Length-initialStats.Length)
		t.Logf("平均每事件: %d 字节", avgMemory)
	}
}

// TestQueueEviction 演示队列淘汰机制
func TestQueueEviction(t *testing.T) {
	t.Log("=== 队列淘汰机制演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 获取初始统计
	initialStats := sb.QueueStats()
	initialEvict := initialStats.EvictCount

	t.Log("\n--- 注入大量事件触发淘汰 ---")
	t.Logf("队列容量: %d", initialStats.MaxQueueSize)

	// 注入足够多的事件以触发淘汰
	// 注意：实际是否触发取决于队列配置
	for i := 0; i < 1000; i++ {
		sb.InjectKey('x')
		if i%100 == 0 {
			stats := sb.QueueStats()
			t.Logf("已注入 %d 事件, 队列: %d, 淘汰: %d",
				i, stats.Length, stats.EvictCount)
		}
	}

	time.Sleep(100 * time.Millisecond)

	finalStats := sb.QueueStats()
	t.Logf("\n--- 最终状态 ---")
	t.Logf("队列长度: %d / %d", finalStats.Length, finalStats.MaxQueueSize)
	t.Logf("总淘汰事件: %d", finalStats.EvictCount)
	t.Logf("本次测试淘汰: %d", finalStats.EvictCount-initialEvict)
}

// TestQueueDuringInteraction 演示交互过程中的队列监控
func TestQueueDuringInteraction(t *testing.T) {
	t.Log("=== 交互过程中的队列监控 ===")

	testApp, err := ui.RunTestWithSandbox(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 模拟用户交互并监控队列
	actions := []struct {
		name string
		action func()
	}{
		{"输入文本", func() {
			sb.InjectString("hello")
		}},
		{"按 Tab", func() {
			sb.InjectSpecialKey(platform.KeyTab)
		}},
		{"按 Enter", func() {
			sb.InjectSpecialKey(platform.KeyEnter)
		}},
		{"组合键", func() {
			sb.InjectKeyWithMod('c', platform.ModCtrl)
		}},
	}

	t.Log("\n--- 执行操作并监控 ---")
	for i, act := range actions {
		beforeStats := sb.QueueStats()

		act.action()
		time.Sleep(50 * time.Millisecond)

		afterStats := sb.QueueStats()

		t.Logf("\n操作 %d: %s", i+1, act.name)
		t.Logf("  队列: %d → %d", beforeStats.Length, afterStats.Length)
		t.Logf("  内存: %d → %d 字节", beforeStats.MemoryUsed, afterStats.MemoryUsed)
	}
}

// TestQueueComparison 演示不同操作对队列的影响
func TestQueueComparison(t *testing.T) {
	t.Log("=== 不同操作的队列影响对比 ===")

	operations := []struct {
		name     string
		operate  func(*testing.T, *mock.MockSandbox)
	}{
		{
			name: "单次按键",
			operate: func(t *testing.T, sb *mock.MockSandbox) {
				sb.InjectKey('a')
			},
		},
		{
			name: "字符串输入",
			operate: func(t *testing.T, sb *mock.MockSandbox) {
				sb.InjectString("test")
			},
		},
		{
			name: "多次操作",
			operate: func(t *testing.T, sb *mock.MockSandbox) {
				for i := 0; i < 5; i++ {
					sb.InjectKey('x')
				}
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			testApp, err := ui.RunTestWithSandbox(StatsApp,
				ui.WithWidth(40),
				ui.WithHeight(18),
			)
			if err != nil {
				t.Fatalf("RunTestWithSandbox failed: %v", err)
			}
			defer testApp.Close()

			sb := testApp.GetSandbox()
			time.Sleep(50 * time.Millisecond)

			before := sb.QueueStats()
			op.operate(t, sb)
			time.Sleep(50 * time.Millisecond)
			after := sb.QueueStats()

			t.Logf("%s: 队列 %+d, 内存 %+d 字节",
				op.name,
				after.Length-before.Length,
				after.MemoryUsed-before.MemoryUsed)
		})
	}
}

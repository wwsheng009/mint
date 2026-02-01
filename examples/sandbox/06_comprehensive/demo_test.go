// 06_comprehensive/demo_test.go
// 综合演示测试 - 组合使用多个高级功能

package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestComprehensiveWorkflow 演示完整的工作流
func TestComprehensiveWorkflow(t *testing.T) {
	t.Log("=== 综合工作流演示 ===")

	// 1. 创建测试应用
	testApp, err := ui.RunTestWithSandbox(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	helper := sb.Helper()
	time.Sleep(100 * time.Millisecond)

	// 2. 设置事件录制
	t.Log("\n--- 1. 启动事件录制 ---")
	recorder := sandbox.NewEventRecorder(1000)
	sb.SetRecorder(recorder)
	t.Log("✅ 录制器已设置")

	// 3. 获取初始队列状态
	initialStats := sb.QueueStats()
	t.Logf("\n--- 2. 初始队列状态 ---")
	t.Logf("队列: %d/%d, 内存: %d bytes",
		initialStats.Length, initialStats.MaxQueueSize, initialStats.MemoryUsed)

	// 4. 使用 TestHelper 执行操作序列
	t.Log("\n--- 3. 执行操作序列 ---")

	// Step 1: 输入名字
	helper.Type("Alice").Process()
	time.Sleep(50 * time.Millisecond)

	// 进入下一步
	helper.Tab().Press(platform.KeyEnter).Process()
	time.Sleep(50 * time.Millisecond)

	// Step 2: 点击 + 按钮几次
	for i := 0; i < 3; i++ {
		helper.Tab().Press(platform.KeyEnter).Process()
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("✅ 操作序列完成")

	// 5. 创建快照
	t.Log("\n--- 4. 创建快照 ---")
	snapshot, err := sb.Snapshot(sandbox.SnapshotStandard, "workflow-step2")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	t.Logf("✅ 快照创建成功: ID=%s", snapshot.Metadata.ID)

	// 6. 获取录制的事件
	recordedEvents := recorder.Events()
	t.Logf("\n--- 5. 录制统计 ---")
	t.Logf("录制了 %d 个事件", len(recordedEvents))

	// 7. 获取当前队列状态
	finalStats := sb.QueueStats()
	t.Logf("\n--- 6. 最终队列状态 ---")
	t.Logf("队列: %d/%d, 内存: %d bytes",
		finalStats.Length, finalStats.MaxQueueSize, finalStats.MemoryUsed)
	t.Logf("内存变化: %+d bytes", finalStats.MemoryUsed-initialStats.MemoryUsed)

	// 8. 验证渲染结果
	rendered := testApp.GetRenderString()
	t.Logf("\n--- 7. 验证结果 ---")
	if contains(rendered, "Alice") && contains(rendered, "Count: 3") {
		t.Log("✅ 工作流执行成功")
	}
}

// TestRecordingSnapshotReplay 演示录制、快照和回放的组合
func TestRecordingSnapshotReplay(t *testing.T) {
	t.Log("=== 录制 + 快照 + 回放 组合演示 ===")

	// === 阶段 1: 录制操作 ===
	t.Log("\n【阶段 1: 录制操作】")

	recorder := sandbox.NewEventRecorder(1000)
	testApp1, err := ui.RunTestWithSandbox(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}

	sb1 := testApp1.GetSandbox()
	sb1.SetRecorder(recorder)
	time.Sleep(50 * time.Millisecond)

	// 执行操作
	sb1.Helper().
		Type("Bob").
		Tab().Press(platform.KeyEnter).  // Next
		Tab().Press(platform.KeyEnter).  // +
		Tab().Press(platform.KeyEnter).  // +
		Process()

	time.Sleep(100 * time.Millisecond)

	// 获取录制
	events := recorder.Events()
	t.Logf("录制了 %d 个事件", len(events))

	// 创建快照
	snapshot, _ := sb1.Snapshot(sandbox.SnapshotStandard, "after-bob")
	t.Logf("快照 ID: %s", snapshot.Metadata.ID)

	rendered1 := testApp1.GetRenderString()
	t.Logf("应用 1 状态:\n%s", rendered1)

	testApp1.Close()

	// === 阶段 2: 回放到新实例 ===
	t.Log("\n【阶段 2: 回放到新实例】")

	testApp2, err := ui.RunTestWithSandbox(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp2.Close()

	sb2 := testApp2.GetSandbox()

	// 回放事件
	for _, ev := range events {
		sb2.InjectRaw(ev)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)
	rendered2 := testApp2.GetRenderString()
	t.Logf("\n回放后状态:\n%s", rendered2)

	// === 阶段 3: 恢复快照 ===
	t.Log("\n【阶段 3: 恢复快照】")

	// 再修改状态
	sb2.Helper().Tab().Press(platform.KeyEnter).Process()  // 再点一次 +
	time.Sleep(50 * time.Millisecond)

	modified := testApp2.GetRenderString()
	t.Logf("\n修改后状态:\n%s", modified)

	// 恢复快照
	sb2.Restore(snapshot)
	time.Sleep(100 * time.Millisecond)

	restored := testApp2.GetRenderString()
	t.Logf("\n恢复快照后状态:\n%s", restored)

	// 验证
	if contains(restored, "Bob") {
		t.Log("✅ 组合功能测试成功")
	}
}

// TestMultiStepSnapshotStrategy 演示多步骤快照策略
func TestMultiStepSnapshotStrategy(t *testing.T) {
	t.Log("=== 多步骤快照策略演示 ===")

	testApp, err := ui.RunTestWithSandbox(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 在不同步骤创建快照
	snapshots := make([]*sandbox.Snapshot, 0)
	stepNames := []string{"initial", "named", "counter-adjusted", "final"}

	// Step 1: 初始快照
	snap, _ := sb.Snapshot(sandbox.SnapshotMinimal, stepNames[0])
	snapshots = append(snapshots, snap)
	t.Logf("快照 1: %s", stepNames[0])

	// Step 2: 输入名字
	sb.Helper().Type("Charlie").Process()
	time.Sleep(50 * time.Millisecond)
	snap, _ = sb.Snapshot(sandbox.SnapshotStandard, stepNames[1])
	snapshots = append(snapshots, snap)
	t.Logf("快照 2: %s", stepNames[1])

	// Step 3: 进入下一步并调整计数器
	sb.Helper().Tab().Press(platform.KeyEnter).Process()
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < 2; i++ {
		sb.Helper().Tab().Press(platform.KeyEnter).Process()
		time.Sleep(50 * time.Millisecond)
	}
	snap, _ = sb.Snapshot(sandbox.SnapshotFull, stepNames[2])
	snapshots = append(snapshots, snap)
	t.Logf("快照 3: %s", stepNames[2])

	// 显示快照列表
	allSnapshots := sb.ListSnapshots()
	t.Logf("\n共有 %d 个快照:", len(allSnapshots))
	for i, sn := range allSnapshots {
		t.Logf("  %d. %s (大小: %d 字节)", i+1, sn.Tags, sn.Size)
	}

	// 恢复到中间状态
	t.Log("\n恢复到步骤 2 状态...")
	sb.Restore(snapshots[1])
	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	if contains(rendered, "Charlie") {
		t.Log("✅ 多步骤快照策略成功")
	}
}

// TestPerformanceMonitoring 演示性能监控
func TestPerformanceMonitoring(t *testing.T) {
	t.Log("=== 性能监控演示 ===")

	testApp, err := ui.RunTestWithSandbox(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 记录初始状态
	initialStats := sb.QueueStats()

	// 执行大量操作
	t.Log("\n执行操作序列...")
	start := time.Now()

	for i := 0; i < 50; i++ {
		sb.InjectKey('a')
		if i%10 == 0 {
			currentStats := sb.QueueStats()
			t.Logf("已注入 %d 事件, 队列: %d", i, currentStats.Length)
		}
	}

	elapsed := time.Since(start)

	// 最终统计
	finalStats := sb.QueueStats()

	t.Logf("\n=== 性能统计 ===")
	t.Logf("操作数量: 50")
	t.Logf("执行时间: %v", elapsed)
	t.Logf("队列变化: %d → %d", initialStats.Length, finalStats.Length)
	t.Logf("内存使用: %d → %d (变化: %+d)",
		initialStats.MemoryUsed, finalStats.MemoryUsed,
		finalStats.MemoryUsed-initialStats.MemoryUsed)

	if finalStats.EvictCount > 0 {
		t.Logf("淘汰事件: %d", finalStats.EvictCount)
	}

	// 性能评估
	avgTime := elapsed / 50
	t.Logf("\n平均每事件: %v", avgTime)

	if avgTime < 1*time.Millisecond {
		t.Log("✅ 性能良好")
	}
}

// TestErrorRecoveryWithSnapshot 演示使用快照进行错误恢复
func TestErrorRecoveryWithSnapshot(t *testing.T) {
	t.Log("=== 快照错误恢复演示 ===")

	testApp, err := ui.RunTestWithSandbox(ComprehensiveApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 建立已知良好状态
	t.Log("\n--- 建立基准状态 ---")
	sb.Helper().
		Type("David").
		Tab().Press(platform.KeyEnter).  // Next
		Tab().Press(platform.KeyEnter).  // +
		Tab().Press(platform.KeyEnter).  // +
		Process()

	time.Sleep(100 * time.Millisecond)

	// 保存基准快照
	baseSnapshot, _ := sb.Snapshot(sandbox.SnapshotStandard, "baseline")
	t.Logf("基准快照: %s", baseSnapshot.Metadata.ID)

	baseRender := testApp.GetRenderString()
	t.Logf("基准状态:\n%s", baseRender)

	// 模拟错误状态 - 混乱操作
	t.Log("\n--- 模拟错误操作 ---")
	for i := 0; i < 5; i++ {
		sb.Helper().
			Tab().
			Press(platform.KeyEnter).
			Process()
		time.Sleep(50 * time.Millisecond)
	}

	errorRender := testApp.GetRenderString()
	t.Logf("错误状态:\n%s", errorRender)

	// 恢复到基准状态
	t.Log("\n--- 恢复到基准状态 ---")
	sb.Restore(baseSnapshot)
	time.Sleep(100 * time.Millisecond)

	recoveredRender := testApp.GetRenderString()
	t.Logf("恢复后状态:\n%s", recoveredRender)

	// 验证恢复
	if contains(recoveredRender, "David") && contains(recoveredRender, "Count: 2") {
		t.Log("✅ 错误恢复成功")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

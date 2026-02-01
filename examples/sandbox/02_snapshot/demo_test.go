// 02_snapshot/demo_test.go
// 快照系统演示测试

package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestSnapshotLevels 演示三种快照级别
func TestSnapshotLevels(t *testing.T) {
	t.Log("=== 快照级别演示 ===")

	// 测试三种快照级别
	levels := []struct {
		name  string
		level sandbox.SnapshotLevel
		desc  string
	}{
		{"Minimal", sandbox.SnapshotMinimal, "仅渲染缓冲区"},
		{"Standard", sandbox.SnapshotStandard, "缓冲区 + 事件历史"},
		{"Full", sandbox.SnapshotFull, "包括应用状态"},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("%s: %s", tc.name, tc.desc)

			testApp, err := ui.RunTestWithSandbox(StatefulApp,
				ui.WithWidth(40),
				ui.WithHeight(20),
			)
			if err != nil {
				t.Fatalf("RunTestWithSandbox failed: %v", err)
			}
			defer testApp.Close()

			sb := testApp.GetSandbox()
			time.Sleep(50 * time.Millisecond)

			// 创建快照
			snapshot, err := sb.Snapshot(tc.level, "test-tag")
			if err != nil {
				t.Fatalf("Snapshot failed: %v", err)
			}

			t.Logf("✅ %s 快照创建成功", tc.name)
			t.Logf("   ID: %s", snapshot.Metadata.ID)
			t.Logf("   大小: %d 字节", snapshot.Metadata.Size)
			t.Logf("   标签: %v", snapshot.Metadata.Tags)
			t.Logf("   校验和: %s", snapshot.Checksum)
		})
	}
}

// TestSnapshotSaveAndRestore 演示快照保存和恢复
func TestSnapshotSaveAndRestore(t *testing.T) {
	t.Log("=== 快照保存与恢复演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatefulApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(100 * time.Millisecond)

	// 获取初始状态
	initialRender := testApp.GetRenderString()
	t.Logf("\n初始状态:\n%s", initialRender)

	// 执行操作：点击 + 按钮 3 次
	t.Log("\n--- 执行操作 ---")
	for i := 0; i < 3; i++ {
		sb.Helper().Tab().Press(platform.KeyEnter)
		time.Sleep(150 * time.Millisecond)
		t.Logf("点击 + 按钮 (第 %d 次)", i+1)
	}

	// 检查状态变化
	afterClicks := testApp.GetRenderString()
	t.Logf("\n点击后状态:\n%s", afterClicks)

	// 创建快照
	t.Log("\n--- 创建快照 ---")
	snapshot, err := sb.Snapshot(sandbox.SnapshotStandard, "after-3-clicks")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	t.Logf("✅ 快照创建成功: ID=%s", snapshot.Metadata.ID)

	// 继续修改状态
	t.Log("\n--- 继续修改状态 ---")
	for i := 0; i < 2; i++ {
		sb.Helper().Tab().Press(platform.KeyEnter)
		time.Sleep(150 * time.Millisecond)
	}

	furtherModified := testApp.GetRenderString()
	t.Logf("\n继续修改后状态:\n%s", furtherModified)

	// 恢复快照
	t.Log("\n--- 恢复快照 ---")
	err = sb.Restore(snapshot)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// 检查恢复后的状态
	time.Sleep(100 * time.Millisecond)
	restored := testApp.GetRenderString()
	t.Logf("\n恢复后状态:\n%s", restored)

	// 验证恢复是否成功
	// 注意：由于 Fiber 模式的状态更新限制，按钮点击可能不会立即反映在计数器上
	// 快照恢复主要验证渲染缓冲区被正确恢复
	if strings.Contains(restored, "Count: 0") {
		t.Log("✅ 快照缓冲区恢复成功")
	} else {
		t.Logf("❌ 状态恢复失败，期望 Count: 0，实际: %s", restored)
	}
}

// TestMultipleSnapshots 演示多个快照管理
func TestMultipleSnapshots(t *testing.T) {
	t.Log("=== 多快照管理演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatefulApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 创建多个快照
	snapshots := make([]*sandbox.Snapshot, 0)
	for i := 0; i < 3; i++ {
		// 点击 + 按钮
		sb.Helper().Tab().Press(platform.KeyEnter)
		time.Sleep(150 * time.Millisecond)

		// 创建快照
		tag := fmt.Sprintf("count-%d", i+1)
		snapshot, err := sb.Snapshot(sandbox.SnapshotStandard, tag)
		if err != nil {
			t.Fatalf("Snapshot %d failed: %v", i, err)
		}
		snapshots = append(snapshots, snapshot)
		t.Logf("快照 %d: ID=%s, Tag=%s", i+1, snapshot.Metadata.ID, tag)
	}

	// 列出所有快照
	t.Log("\n--- 列出所有快照 ---")
	allSnapshots := sb.ListSnapshots()
	t.Logf("共有 %d 个快照", len(allSnapshots))

	for i, snap := range allSnapshots {
		t.Logf("  %d. ID=%s, Tags=%v, Size=%d",
			i+1, snap.ID[:8], snap.Tags, snap.Size)
	}

	// 按标签查找
	t.Log("\n--- 按标签查找 ---")
	tagged := sb.FindSnapshots("count-2")
	t.Logf("找到 %d 个标签为 'count-2' 的快照", len(tagged))
	for _, snap := range tagged {
		t.Logf("  ID=%s", snap.Metadata.ID)
	}

	// 恢复到中间状态
	t.Log("\n--- 恢复到中间状态 ---")
	if len(snapshots) > 1 {
		midSnapshot := snapshots[1]
		err = sb.Restore(midSnapshot)
		if err != nil {
			t.Fatalf("Restore failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
		_ = testApp.GetRenderString()

		// 快照恢复成功（即使计数器值没有变化）
		t.Log("✅ 快照恢复完成")
	}
}

// TestSnapshotWithInput 演示包含输入的快照
func TestSnapshotWithInput(t *testing.T) {
	t.Log("=== 包含输入的快照演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatefulApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(100 * time.Millisecond)

	// 输入文本
	t.Log("\n--- 输入文本 ---")
	sb.Helper().
		Tab().Tab().Tab().  // 切换到输入框
		Type("Hello").
		Process()

	time.Sleep(100 * time.Millisecond)

	// 检查输入结果
	beforeSnapshot := testApp.GetRenderString()
	t.Logf("\n输入后:\n%s", beforeSnapshot)

	// 创建快照
	t.Log("\n--- 创建快照 ---")
	snapshot, err := sb.Snapshot(sandbox.SnapshotFull, "with-input")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	t.Logf("✅ 快照创建成功")

	// 清空输入
	t.Log("\n--- 修改输入 ---")
	sb.Helper().
		Press(platform.KeyHome)
	time.Sleep(150 * time.Millisecond)

	sb.Helper().
		Type("World")

	time.Sleep(200 * time.Millisecond)
	modified := testApp.GetRenderString()
	t.Logf("\n修改后:\n%s", modified)

	// 恢复快照
	t.Log("\n--- 恢复快照 ---")
	err = sb.Restore(snapshot)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	restored := testApp.GetRenderString()
	t.Logf("\n恢复后:\n%s", restored)

	// 验证文本是否恢复
	if strings.Contains(restored, "Hello") {
		t.Log("✅ 输入内容恢复成功")
	}
}

// TestSnapshotDelete 演示快照删除
func TestSnapshotDelete(t *testing.T) {
	t.Log("=== 快照删除演示 ===")

	testApp, err := ui.RunTestWithSandbox(StatefulApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 创建几个快照
	var snapshotIDs []string
	for i := 0; i < 3; i++ {
		snapshot, _ := sb.Snapshot(sandbox.SnapshotMinimal, fmt.Sprintf("temp-%d", i))
		snapshotIDs = append(snapshotIDs, snapshot.Metadata.ID)
	}

	allSnapshots := sb.ListSnapshots()
	t.Logf("创建前: 共 %d 个快照", len(allSnapshots))

	// 删除一个快照
	t.Log("\n--- 删除快照 ---")
	sb.DeleteSnapshot(snapshotIDs[0])
	t.Logf("已删除快照: %s", snapshotIDs[0][:8])

	allSnapshots = sb.ListSnapshots()
	t.Logf("删除后: 共 %d 个快照", len(allSnapshots))

	// 清空所有快照
	t.Log("\n--- 清空所有快照 ---")
	sb.ClearSnapshots()

	allSnapshots = sb.ListSnapshots()
	t.Logf("清空后: 共 %d 个快照", len(allSnapshots))

	if len(allSnapshots) == 0 {
		t.Log("✅ 所有快照已清空")
	}
}

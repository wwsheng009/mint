package main

// ============================================================================
// 旧版测试 - 已弃用
// 这些测试使用 ui.TestRun (旧版 API)，依赖 Sandbox Helper 功能
// 新版 ui.RunTest 不提供 Sandbox() 方法
// 请使用 sandbox_mock_test.go 中的新版测试
// ============================================================================

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// SnapshotExamples_OLD 演示快照功能的各种用法 (已弃用)
func SnapshotExamples_OLD(t *testing.T) {
	t.Run("BasicSnapshot", testBasicSnapshot_OLD)
	t.Run("SnapshotLevels", testSnapshotLevels_OLD)
	t.Run("RestoreSnapshot", testRestoreSnapshot_OLD)
	t.Run("SnapshotTags", testSnapshotTags_OLD)
	t.Run("SnapshotWithActions", testSnapshotWithActions_OLD)
	t.Run("SnapshotReusability", testSnapshotReusability_OLD)
}

// testBasicSnapshot_OLD 基础快照功能 (已弃用)
func testBasicSnapshot_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 1. 创建初始快照
	snap, err := sb.Snapshot(sandbox.SnapshotStandard, "initial")
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	t.Logf("Created snapshot: ID=%s, Time=%v, Level=%s",
		snap.Metadata.ID,
		snap.Metadata.Timestamp,
		snap.Metadata.Level)

	// 2. 验证快照包含正确数据
	if snap.Buffer == nil {
		t.Error("Snapshot buffer is nil")
	}

	if len(snap.Events) != 0 {
		t.Logf("Snapshot contains %d events", len(snap.Events))
	}
}

// testSnapshotLevels_OLD 测试不同快照级别 (已弃用)
func testSnapshotLevels_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	helper := testApp.Helper()

	// 先做一些操作
	helper.Tab().Tab().Enter().Process()

	// 测试 Minimal 级别
	snapMinimal, err := sb.Snapshot(sandbox.SnapshotMinimal, "minimal-test")
	if err != nil {
		t.Errorf("Minimal snapshot failed: %v", err)
	}
	t.Logf("Minimal snapshot: ID=%s", snapMinimal.Metadata.ID)
	t.Logf("  - Buffer: %v", snapMinimal.Buffer != nil)
	t.Logf("  - Events: %d", len(snapMinimal.Events))
	t.Logf("  - State: %d items", len(snapMinimal.State))

	// 测试 Standard 级别
	snapStandard, err := sb.Snapshot(sandbox.SnapshotStandard, "standard-test")
	if err != nil {
		t.Errorf("Standard snapshot failed: %v", err)
	}
	t.Logf("Standard snapshot: ID=%s", snapStandard.Metadata.ID)
	t.Logf("  - Buffer: %v", snapStandard.Buffer != nil)
	t.Logf("  - Events: %d", len(snapStandard.Events))
	t.Logf("  - State: %d items", len(snapStandard.State))

	// 测试 Full 级别
	snapFull, err := sb.Snapshot(sandbox.SnapshotFull, "full-test")
	if err != nil {
		t.Errorf("Full snapshot failed: %v", err)
	}
	t.Logf("Full snapshot: ID=%s", snapFull.Metadata.ID)
	t.Logf("  - Buffer: %v", snapFull.Buffer != nil)
	t.Logf("  - Events: %d", len(snapFull.Events))
	t.Logf("  - State: %d items", len(snapFull.State))
}

// testRestoreSnapshot_OLD 快照恢复测试 (已弃用)
func testRestoreSnapshot_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	helper := testApp.Helper()

	// 1. 执行一些操作
	helper.Tab().Tab().Enter().Process()
	helper.AssertRender("Count: 1").Result()

	// 2. 创建快照
	snap, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-first-click")

	// 3. 继续更多操作
	helper.Tab().Tab().Enter().Process()
	helper.AssertRender("Count: 2").Result()

	// 4. 恢复到快照
	err = sb.Restore(snap)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// 5. 验证已恢复到快照状态
	err = sb.AssertRender("Count: 1")
	if err != nil {
		t.Errorf("After restore, state mismatch: %v", err)
	}
}

// testSnapshotTags_OLD 测试快照标签 (已弃用)
func testSnapshotTags_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 创建带多个标签的快照
	snap, err := sb.Snapshot(
		sandbox.SnapshotStandard,
		"initial",
		"clean-state",
		"before-tests",
	)
	if err != nil {
		t.Fatal(err)
	}

	// 验证标签
	t.Logf("Snapshot tags: %v", snap.Metadata.Tags)
	if len(snap.Metadata.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(snap.Metadata.Tags))
	}

	expectedTags := []string{"initial", "clean-state", "before-tests"}
	for _, tag := range expectedTags {
		found := false
		for _, snapTag := range snap.Metadata.Tags {
			if snapTag == tag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag '%s' not found", tag)
		}
	}
}

// testSnapshotWithActions_OLD 使用快照保存和恢复操作 (已弃用)
func testSnapshotWithActions_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	helper := testApp.Helper()

	// 场景：测试多个按钮操作

	// 保存初始状态
	initial, _ := sb.Snapshot(sandbox.SnapshotStandard, "initial")

	// 操作 1：点击 "+"
	helper.Tab().Tab().Enter().Process()
	snap1, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-+")

	// 操作 2：再点击 "+"
	helper.Tab().Tab().Enter().Process()
	snap2, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-++")

	// 操作 3：点击 "-"
	helper.Tab().Tab().Tab().Enter().Process()
	snap3, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-+-")

	// 测试：恢复到不同快照
	restores := []struct {
		name      string
		snap      *sandbox.Snapshot
		expected  string
	}{
		{"Initial", initial, "Count: 0"},
		{"After first +", snap1, "Count: 1"},
		{"After second +", snap2, "Count: 2"},
		{"After -", snap3, "Count: 1"},
	}

	for _, test := range restores {
		t.Run(test.name, func(t *testing.T) {
			// 恢复
			err := sb.Restore(test.snap)
			if err != nil {
				t.Fatalf("Restore failed: %v", err)
			}

			// 验证
			err = sb.AssertRender(test.expected)
			if err != nil {
				t.Errorf("After restore to %s: %v", test.name, err)
			}
		})
	}
}

// testSnapshotReusability_OLD 快照可重用性测试 (已弃用)
func testSnapshotReusability_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	helper := testApp.Helper()

	// 创建快照
	helper.Tab().Tab().Enter().Process()
	snap, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-first-increment")

	// 多次恢复到同一个快照
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("Restore %d", i), func(t *testing.T) {
			// 先做一些操作
			helper.Tab().Tab().Enter().Process()

			// 恢复
			err := sb.Restore(snap)
			if err != nil {
				t.Fatalf("Restore %d failed: %v", i, err)
			}

			// 验证状态一致
			err = sb.AssertRender("Count: 1")
			if err != nil {
				t.Errorf("Restore %d: state mismatch: %v", i, err)
			}
		})
	}
}

// testSnapshotList_OLD 列出快照测试 (已弃用)
func testSnapshotList_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 创建多个快照
	_, _ = sb.Snapshot(sandbox.SnapshotStandard, "snap1", "tag1")
	_, _ = sb.Snapshot(sandbox.SnapshotStandard, "snap2", "tag2")
	_, _ = sb.Snapshot(sandbox.SnapshotStandard, "snap3", "tag1")

	// 列出所有快照
	snapshots := sb.ListSnapshots()

	if len(snapshots) != 3 {
		t.Errorf("Expected 3 snapshots, got %d", len(snapshots))
	}

	// 打印快照信息
	for i, meta := range snapshots {
		t.Logf("Snapshot %d:", i)
		t.Logf("  ID: %s", meta.ID)
		t.Logf("  Time: %v", meta.Timestamp)
		t.Logf("  Level: %s", meta.Level)
		t.Logf("  Tags: %v", meta.Tags)
		t.Logf("  Size: %d bytes", meta.Size)
	}
}

// testSnapshotEviction_OLD 快照淘汰测试 (已弃用)
func testSnapshotEviction_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 注意：沙箱默认最大快照数为 100
	// 这里我们创建少量快照来验证淘汰逻辑

	// 创建快照并记录 ID
	snapshotIDs := []string{}
	for i := 0; i < 10; i++ {
		snap, _ := sb.Snapshot(sandbox.SnapshotStandard, fmt.Sprintf("snap-%d", i))
		snapshotIDs = append(snapshotIDs, snap.Metadata.ID)
	}

	// 列出当前快照
	snapshots := sb.ListSnapshots()

	// 验证快照数量（应该是最新的 10 个）
	if len(snapshots) != 10 {
		t.Errorf("Expected 10 snapshots, got %d", len(snapshots))
	}

	// 验证最新快照是最后创建的
	lastSnap := snapshots[len(snapshots)-1]
	if lastSnap.ID != snapshotIDs[len(snapshotIDs)-1] {
		t.Errorf("Last snapshot ID mismatch")
	}
}

// testSnapshotPerformance_OLD 快照性能测试 (已弃用)
func testSnapshotPerformance_OLD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	helper := testApp.Helper()

	// 性能测试：创建 50 个快照
	start := time.Now()

	for i := 0; i < 50; i++ {
		helper.Tab().Tab().Enter().Process()
		_, err := sb.Snapshot(sandbox.SnapshotStandard, fmt.Sprintf("perf-snap-%d", i))
		if err != nil {
			t.Errorf("Snapshot %d failed: %v", i, err)
			break
		}
	}

	elapsed := time.Since(start)
	t.Logf("Created 50 snapshots in: %v", elapsed)
	t.Logf("Average per snapshot: %v", elapsed/time.Duration(50))

	// 测试恢复性能
	start = time.Now()

	snapshots := sb.ListSnapshots()
	for i, snapMeta := range snapshots {
		// 获取快照（这里简化了，实际需要从管理器获取）
		_ = snapMeta
		if i >= 10 { // 只测试 10 次恢复
			break
		}
	}

	elapsed = time.Since(start)
	t.Logf("Listed and accessed snapshots in: %v", elapsed)
}

// testSnapshotIntegrity_OLD 快照完整性测试 (已弃用)
func testSnapshotIntegrity_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 创建快照
	snap, err := sb.Snapshot(sandbox.SnapshotFull, "integrity-test")
	if err != nil {
		t.Fatal(err)
	}

	// 验证校验和存在
	if snap.Checksum == "" {
		t.Error("Snapshot checksum is empty")
	}

	// 验证元数据完整
	if snap.Metadata.ID == "" {
		t.Error("Snapshot ID is empty")
	}

	if snap.Metadata.Timestamp.IsZero() {
		t.Error("Snapshot timestamp is zero")
	}

	if snap.Metadata.Size <= 0 {
		t.Error("Snapshot size is invalid")
	}

	t.Logf("Snapshot integrity verified")
	t.Logf("  Checksum: %s", snap.Checksum)
	t.Logf("  Size: %d bytes", snap.Metadata.Size)
}

// testSnapshotWithComplexState_OLD 测试包含复杂状态的快照 (已弃用)
func testSnapshotWithComplexState_OLD(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	helper := testApp.Helper()

	// 创建复杂状态：修改名字并多次点击按钮
	helper.Tab().Tab().Tab().Type("Alice").Process()
	for i := 0; i < 5; i++ {
		helper.Tab().Tab().Tab().Enter().Process()
	}

	// 验证状态
	helper.AssertRender("Hello, Alice").Result()
	helper.AssertRender("Count: 5").Result()

	// 创建快照
	snap, _ := sb.Snapshot(sandbox.SnapshotFull, "complex-state")

	// 做更多操作改变状态
	helper.Tab().Tab().Tab().Type("Bob").Process()
	helper.Tab().Tab().Tab().Tab().Enter().Process()
	helper.AssertRender("Hello, Bob").Result()
	helper.AssertRender("Count: 6").Result()

	// 恢复到复杂状态
	sb.Restore(snap)

	// 验证恢复的复杂状态
	err = sb.AssertRender("Hello, Alice")
	if err != nil {
		t.Errorf("Name not restored: %v", err)
	}

	err = sb.AssertRender("Count: 5")
	if err != nil {
		t.Errorf("Count not restored: %v", err)
	}
}

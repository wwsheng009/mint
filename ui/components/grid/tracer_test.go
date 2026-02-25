package grid

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestGrid_TracingIntegration 测试 Grid 组件的追踪系统集成
func TestGrid_TracingIntegration(t *testing.T) {
	// 启用追踪器
	layout.EnableTracer()
	defer layout.DisableTracer()
	defer layout.ClearTrace()

	// 创建一个带 Flex 列的 Grid
	g := New().
		SetColumns(Flex{Factor: 1}, Flex{Factor: 2}).
		SetRows(Auto{}, Auto{}).
		SetGap(1, 1).
		SetChildrenAuto([]rtui.VNode{})
	g.SetKey("test-grid")  // 单独调用，避免类型推断问题

	// 创建 Instance 并执行测量
	inst := g.CreateInstance()
	gridInst, ok := inst.(*Instance)
	if !ok {
		t.Fatalf("Failed to cast to *Instance")
	}
	constraints := layout.NewConstraints(0, 100, 0, 50)
	size := gridInst.Measure(constraints)

	// 验证测量结果
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Invalid size: %v", size)
	}

	// 获取追踪条目
	entries := layout.GetTraceEntries()
	if len(entries) == 0 {
		t.Fatal("No trace entries were recorded. Tracing integration may not be working.")
	}

	// 验证入口追踪
	foundEntrance := false
	for _, entry := range entries {
		if entry.Reason == "Grid.Measure entrance" {
			foundEntrance = true
			if entry.Input.MaxWidth != 100 || entry.Input.MaxHeight != 50 {
				t.Errorf("Input constraints: expected {0..100} × {0..50}, got %v", entry.Input)
			}
			break
		}
	}
	if !foundEntrance {
		t.Error("Grid.Measure entrance trace not found")
	}

	// 验证配置追踪
	foundConfig := false
	for _, entry := range entries {
		if strings.Contains(entry.Reason, "Config:") {
			foundConfig = true
			if !strings.Contains(entry.Reason, "cols=2") || !strings.Contains(entry.Reason, "rows=2") {
				t.Errorf("Config trace missing expected info: %s", entry.Reason)
			}
			break
		}
	}
	if !foundConfig {
		t.Error("Grid config trace not found")
	}

	// 验证列追踪
	foundColumns := false
	for _, entry := range entries {
		if strings.Contains(entry.Path, "/col-0") && strings.Contains(entry.Reason, "Column 0") {
			foundColumns = true
			if !strings.Contains(entry.Reason, "Flex") {
				t.Errorf("Column 0 trace missing Flex description: %s", entry.Reason)
			}
			break
		}
	}
	if !foundColumns {
		t.Error("Column traces not found")
	}

	// 验证行追踪
	foundRows := false
	for _, entry := range entries {
		if strings.Contains(entry.Path, "/row-0") && strings.Contains(entry.Reason, "Row 0") {
			foundRows = true
			if !strings.Contains(entry.Reason, "Auto") {
				t.Errorf("Row 0 trace missing Auto description: %s", entry.Reason)
			}
			break
		}
	}
	if !foundRows {
		t.Error("Row traces not found")
	}

	// 验证出口追踪
	foundExit := false
	for _, entry := range entries {
		if strings.Contains(entry.Reason, "Grid.Measure complete") {
			foundExit = true
			if entry.Dimension.Width != size.Width || entry.Dimension.Height != size.Height {
				t.Errorf("Exit dimension: expected %v, got %v", size, entry.Dimension)
			}
			break
		}
	}
	if !foundExit {
		t.Error("Grid.Measure complete trace not found")
	}

	t.Logf("Tracing integration test passed with %d entries", len(entries))

	// 输出追踪日志（调试用）
	t.Log("Trace output:")
	t.Log(layout.DumpTrace())
}

// TestGrid_TracingDisabled 测试禁用追踪时的行为
func TestGrid_TracingDisabled(t *testing.T) {
	// 确保追踪器已禁用
	layout.DisableTracer()
	layout.ClearTrace()

	// 创建并测量 Grid
	g := New().SetColumns(Flex{Factor: 1}).SetRows(Auto{})
	inst := g.CreateInstance()
	gridInst, ok := inst.(*Instance)
	if !ok {
		t.Fatalf("Failed to cast to *Instance")
	}
	constraints := layout.NewConstraints(0, 50, 0, 30)
	size := gridInst.Measure(constraints)

	// 验证测量仍然工作
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("Measurement failed: %v", size)
	}

	// 验证没有追踪条目
	entries := layout.GetTraceEntries()
	if len(entries) != 0 {
		t.Errorf("Expected no trace entries when disabled, got %d", len(entries))
	}
}

// TestGrid_TracingWithCellBorders 测试包含格子边框的追踪
func TestGrid_TracingWithCellBorders(t *testing.T) {
	layout.EnableTracer()
	defer layout.DisableTracer()
	defer layout.ClearTrace()

	g := New().
		SetColumns(Fixed(20), Fixed(20)).
		SetRows(Auto{}).
		ShowCellBorders().
		SetGap(1, 1)
	g.SetKey("borders-grid")  // 单独调用，避免类型推断问题

	inst := g.CreateInstance()
	gridInst, ok := inst.(*Instance)
	if !ok {
		t.Fatalf("Failed to cast to *Instance")
	}
	constraints := layout.NewConstraints(0, 100, 0, 50)
	_ = gridInst.Measure(constraints)

	entries := layout.GetTraceEntries()

	// 验证配置追踪包含边框信息
	foundBorderInfo := false
	for _, entry := range entries {
		if strings.Contains(entry.Reason, "Config:") && strings.Contains(entry.Reason, "showCellBorders=true") {
			foundBorderInfo = true
			break
		}
	}
	if !foundBorderInfo {
		t.Error("Config trace missing cell borders info")
	}

	t.Logf("Tracing with cell borders: %d entries", len(entries))
}

// BenchmarkGrid_TracingWith 测试启用追踪时的性能影响
func BenchmarkGrid_TracingWith(b *testing.B) {
	layout.EnableTracer()

	g := New().
		SetColumns(Fixed(10), Flex{Factor: 1}, Flex{Factor: 2}).
		SetRows(Auto{}, Auto{}).
		SetGap(1, 1)

	inst := g.CreateInstance()
	gridInst := inst.(*Instance)
	constraints := layout.NewConstraints(0, 100, 0, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layout.ClearTrace()
		gridInst.Measure(constraints)
	}
}

// BenchmarkGrid_TracingWithout 测试禁用追踪时的性能
func BenchmarkGrid_TracingWithout(b *testing.B) {
	layout.DisableTracer()

	g := New().
		SetColumns(Fixed(10), Flex{Factor: 1}, Flex{Factor: 2}).
		SetRows(Auto{}, Auto{}).
		SetGap(1, 1)

	inst := g.CreateInstance()
	gridInst := inst.(*Instance)
	constraints := layout.NewConstraints(0, 100, 0, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gridInst.Measure(constraints)
	}
}

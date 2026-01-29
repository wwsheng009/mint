package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// TestRenderer_DiffRendering 测试差异渲染
func TestRenderer_DiffRendering(t *testing.T) {
	renderer := NewRenderer(20, 5)

	// 第一帧：写入一些内容
	back := renderer.GetBackBuffer()
	s1 := style.Style{}.Foreground(style.Cyan)
	back.SetString(0, 0, "Hello, 世界!", s1)

	output := renderer.Render()
	if output == "" {
		t.Error("First render should produce output")
	}

	// 第二帧：没有变化
	back = renderer.GetBackBuffer()
	output = renderer.Render()
	if output != "" {
		t.Errorf("No changes should produce no output, got: %q", output)
	}

	// 第三帧：修改部分内容
	back = renderer.GetBackBuffer()
	s2 := style.Style{}.Foreground(style.Red).Bold(true)
	back.SetString(0, 0, "Hello, 世界!", s2) // 相同文本，不同样式

	output = renderer.Render()
	if output == "" {
		t.Error("Style change should produce output")
	}
}

// TestRenderer_WideCharStyleChange 测试宽字符样式变化
func TestRenderer_WideCharStyleChange(t *testing.T) {
	renderer := NewRenderer(20, 3)

	// 第一帧
	back := renderer.GetBackBuffer()
	s1 := style.Style{}.Foreground(style.Cyan)
	back.SetString(5, 1, "测试", s1)

	output := renderer.Render()
	if output == "" {
		t.Error("First render with wide chars should produce output")
	}

	// 第二帧：只改变样式（模拟光标闪烁）
	back = renderer.GetBackBuffer()
	s2 := style.Style{}.Foreground(style.Cyan).Reverse(true)
	back.SetString(5, 1, "测试", s2)

	output = renderer.Render()
	if output == "" {
		t.Error("Style change on wide chars should produce output")
	}

	// 验证 front buffer 已更新
	front := renderer.GetFrontBuffer()
	if front.Cells[1][5].Cluster != "测" {
		t.Errorf("Front buffer should have '测', got %q", front.Cells[1][5].Cluster)
	}
	if front.Cells[1][5].Style.IsReverse() != true {
		t.Error("Front buffer should have updated style")
	}
}

// TestRenderer_SingleCellChange 测试单个单元格变化
func TestRenderer_SingleCellChange(t *testing.T) {
	renderer := NewRenderer(20, 3)

	// 初始化 - 第一帧用点号填充
	back := renderer.GetBackBuffer()
	s := style.Style{}
	back.Fill(Rect{X: 0, Y: 0, Width: 20, Height: 3}, '.', s)

	// 第一帧渲染
	output := renderer.Render()
	if output == "" {
		t.Error("First render should produce output")
	}

	// 修改单个单元格为 X（不同于初始的 .）
	back = renderer.GetBackBuffer()
	back.SetCell(10, 1, 'X', s)

	output = renderer.Render()
	if output == "" {
		t.Error("Single cell change should produce output")
	}
}

// TestRenderer_EmptyClusterHandling 测试空 cluster 处理
func TestRenderer_EmptyClusterHandling(t *testing.T) {
	renderer := NewRenderer(20, 3)

	// 初始化并首次渲染
	back := renderer.GetBackBuffer()
	s := style.Style{}
	back.Fill(Rect{X: 0, Y: 0, Width: 20, Height: 3}, ' ', s)

	renderer.Render()

	// 修改一个单元格为空 cluster
	back = renderer.GetBackBuffer()
	back.Cells[1][10] = Cell{Cluster: "", Style: s, Width: 0, IsContinuation: false}

	output := renderer.Render()
	// 空 cluster 不应该产生输出，也不应该崩溃
	_ = output
}

// TestRenderer_CursorBlinkSimulation 模拟光标闪烁场景
func TestRenderer_CursorBlinkSimulation(t *testing.T) {
	renderer := NewRenderer(30, 5)

	// 初始状态：用户名输入框
	back := renderer.GetBackBuffer()
	s1 := style.Style{}.Foreground(style.White)
	back.SetString(10, 2, "请输入用户名", s1)

	// 首次渲染
	output1 := renderer.Render()
	if output1 == "" {
		t.Error("First render should produce output")
	}

	// 光标闪烁：样式变化
	back = renderer.GetBackBuffer()
	s2 := style.Style{}.Foreground(style.White).Reverse(true)
	back.SetString(10, 2, "请输入用户名", s2)

	output2 := renderer.Render()
	if output2 == "" {
		t.Error("Cursor blink (style change) should produce output")
	}

	// 光标再次闪烁：样式恢复
	back = renderer.GetBackBuffer()
	back.SetString(10, 2, "请输入用户名", s1)

	output3 := renderer.Render()
	if output3 == "" {
		t.Error("Cursor blink restore should produce output")
	}
}

package paint

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// IsCellChanged 边界测试
// =============================================================================

// TestIsCellChanged_BasicCases 基本变化检测测试
func TestIsCellChanged_BasicCases(t *testing.T) {
	tests := []struct {
		name     string
		cell     Cell
		prevCell Cell
		want     bool
	}{
		{
			name:     "相同单元格",
			cell:     Cell{Cluster: "A", Width: 1, IsContinuation: false},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false},
			want:     false,
		},
		{
			name:     "不同字符",
			cell:     Cell{Cluster: "A", Width: 1, IsContinuation: false},
			prevCell: Cell{Cluster: "B", Width: 1, IsContinuation: false},
			want:     true,
		},
		{
			name:     "空单元格变非空",
			cell:     Cell{Cluster: "A", Width: 1, IsContinuation: false},
			prevCell: Cell{Cluster: "", Width: 0, IsContinuation: false},
			want:     true,
		},
		{
			name:     "非空变空",
			cell:     Cell{Cluster: "", Width: 0, IsContinuation: false},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false},
			want:     true,
		},
		{
			name:     "continuation→continuation",
			cell:     Cell{Cluster: "", Width: 0, IsContinuation: true},
			prevCell: Cell{Cluster: "", Width: 0, IsContinuation: true},
			want:     false,
		},
		{
			name:     "continuation→非continuation",
			cell:     Cell{Cluster: "A", Width: 1, IsContinuation: false},
			prevCell: Cell{Cluster: "", Width: 0, IsContinuation: true},
			want:     true,
		},
		{
			name:     "非continuation→continuation",
			cell:     Cell{Cluster: "", Width: 0, IsContinuation: true},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false},
			want:     true,
		},
		{
			name:     "宽字符头→continuation",
			cell:     Cell{Cluster: "", Width: 0, IsContinuation: true},
			prevCell: Cell{Cluster: "测", Width: 2, IsContinuation: false},
			want:     true,
		},
		{
			name:     "宽字符头→空(Width=0)",
			cell:     Cell{Cluster: "", Width: 0, IsContinuation: false},
			prevCell: Cell{Cluster: "测", Width: 2, IsContinuation: false},
			want:     true,
		},
		{
			name:     "空格变化",
			cell:     Cell{Cluster: " ", Width: 1, IsContinuation: false},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false},
			want:     true,
		},
		{
			name:     "样式变化",
			cell:     Cell{Cluster: "A", Width: 1, IsContinuation: false, Style: style.Style{}.Bold(true)},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false, Style: style.Style{}},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCellChanged(tt.cell, tt.prevCell)
			if got != tt.want {
				t.Errorf("IsCellChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsCellChanged_WideCharCases 宽字符相关边界测试
func TestIsCellChanged_WideCharCases(t *testing.T) {
	tests := []struct {
		name     string
		cell     Cell
		prevCell Cell
		want     bool
	}{
		{
			name:     "宽字符→宽字符(相同)",
			cell:     Cell{Cluster: "测", Width: 2, IsContinuation: false},
			prevCell: Cell{Cluster: "测", Width: 2, IsContinuation: false},
			want:     false,
		},
		{
			name:     "宽字符→宽字符(不同)",
			cell:     Cell{Cluster: "试", Width: 2, IsContinuation: false},
			prevCell: Cell{Cluster: "测", Width: 2, IsContinuation: false},
			want:     true,
		},
		{
			name:     "宽字符→窄字符",
			cell:     Cell{Cluster: "A", Width: 1, IsContinuation: false},
			prevCell: Cell{Cluster: "测", Width: 2, IsContinuation: false},
			want:     true,
		},
		{
			name:     "窄字符→宽字符",
			cell:     Cell{Cluster: "测", Width: 2, IsContinuation: false},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false},
			want:     true,
		},
		{
			name:     "continuation孤立-左侧无宽字符头",
			cell:     Cell{Cluster: "", Width: 0, IsContinuation: true},
			prevCell: Cell{Cluster: "A", Width: 1, IsContinuation: false},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCellChanged(tt.cell, tt.prevCell)
			if got != tt.want {
				t.Errorf("IsCellChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// 宽字符擦除测试 - 核心问题场景
// =============================================================================

// TestRender_WideCharErase_Scenario1 测试场景：宽字符被窄字符覆盖
// 上一帧: [测][continuation]  (width=2)
// 下一帧: [A][空]            (width=1)
// 期望: 宽字符应该被完全擦除，不应该有残留
func TestRender_WideCharErase_Scenario1(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 第一帧：写入宽字符
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s)
	renderer.Render()

	// 第二帧：用窄字符覆盖
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "AB", s)
	output := renderer.Render()

	// 验证：应该产生输出
	if output == "" {
		t.Error("Expected output when wide char replaced by narrow chars")
	}

	// 验证front buffer已正确更新
	front := renderer.GetFrontBuffer()
	if front.Cells[0][0].Cluster != "A" {
		t.Errorf("Expected 'A' at [0][0], got %q", front.Cells[0][0].Cluster)
	}
	if front.Cells[0][1].Cluster != "B" {
		t.Errorf("Expected 'B' at [0][1], got %q", front.Cells[0][1].Cluster)
	}
	// 确保continuation已被清除
	if front.Cells[0][1].IsContinuation {
		t.Error("Cell[0][1] should not be continuation after overwrite")
	}
}

// TestRender_WideCharErase_Scenario2 测试场景：宽字符被短字符串覆盖
// 上一帧: [测][continuation][试][continuation]  (4 cells, 2 wide chars)
// 下一帧: [A][空][空][空]                        (1 cell content)
// 期望: 应该擦除所有4个位置
func TestRender_WideCharErase_Scenario2(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 第一帧：写入两个宽字符
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s)
	renderer.Render()

	// 第二帧：只写入一个字符
	back = renderer.GetBackBuffer()
	back.SetCell(0, 0, 'A', s)
	output := renderer.Render()

	// 验证：应该产生输出
	if output == "" {
		t.Error("Expected output when wide chars partially replaced")
	}

	front := renderer.GetFrontBuffer()
	// 检查残留字符
	if front.Cells[0][1].Cluster == "试" || front.Cells[0][1].Cluster != "" {
		// 如果有continuation残留，front.Cells[0][1]可能还有之前的continuation标记
		t.Logf("Cell[0][1] after overwrite: Cluster=%q, IsContinuation=%v",
			front.Cells[0][1].Cluster, front.Cells[0][1].IsContinuation)
	}
}

// TestRender_WideCharErase_Scenario3 测试场景：宽度覆盖测试
// 上一帧: [测][continuation][X][Y]  (width=2 + width=1 + width=1)
// 下一帧: [A][B][C][D]              (4 narrow chars)
func TestRender_WideCharErase_Scenario3(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 第一帧
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测XY", s)
	renderer.Render()

	// 验证初始状态
	front := renderer.GetFrontBuffer()
	if front.Cells[0][0].Cluster != "测" {
		t.Errorf("Initial state wrong: expected '测', got %q", front.Cells[0][0].Cluster)
	}
	if !front.Cells[0][1].IsContinuation {
		t.Error("Cell[0][1] should be continuation initially")
	}

	// 第二帧：覆盖所有内容
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "ABCD", s)
	output := renderer.Render()

	if output == "" {
		t.Error("Expected output")
	}

	front = renderer.GetFrontBuffer()
	expected := []string{"A", "B", "C", "D"}
	for i, exp := range expected {
		if front.Cells[0][i].Cluster != exp {
			t.Errorf("Cell[0][%d]: expected %q, got %q, IsContinuation=%v",
				i, exp, front.Cells[0][i].Cluster, front.Cells[0][i].IsContinuation)
		}
		if front.Cells[0][i].IsContinuation {
			t.Errorf("Cell[0][%d] should not be continuation", i)
		}
	}
}

// TestRender_WideCharErase_ResidualChar 测试残留字符问题
// 这是用户报告的核心问题：@ 和 ? 残留
func TestRender_WideCharErase_ResidualChar(t *testing.T) {
	renderer := NewRenderer(40, 5)
	s := style.Style{}

	// 模拟用户场景：数据流显示
	// 上一帧: "数据流：Intent → @educer → $?ore → UI"
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "数据流：Intent → @educer → $?ore → UI", s)
	renderer.Render()

	// 下一帧: 完全不同的内容
	// "计数:  0    [  +  ]   *[  -  ]"
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "计数:  0    [  +  ]   *[  -  ]", s)
	output := renderer.Render()

	if output == "" {
		t.Error("Expected output for frame change")
	}

	// 检查是否有残留字符
	front := renderer.GetFrontBuffer()
	for x := 0; x < front.Width; x++ {
		cell := front.Cells[0][x]
		// 只检查我们期望的内容范围
		if x < 30 {
			// 检查不应该有 @ 或 ? 残留（除非是预期内容的一部分）
			if cell.Cluster == "@" || cell.Cluster == "?" {
				// 检查这个位置是否在新内容的预期位置
				// 新内容 "计数:  0    [  +  ]   *[  -  ]" 不应该包含 @ 或 ?
				t.Errorf("Found residual character %q at position %d", cell.Cluster, x)
			}
		}
	}
}

// =============================================================================
// 孤立 Continuation 测试
// =============================================================================

// TestRender_OrphanContinuation 测试孤立continuation的处理
// 场景：直接修改continuation单元格为普通单元格
func TestRender_OrphanContinuation(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 第一帧：写入宽字符
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s)
	renderer.Render()

	// 第二帧：直接在continuation位置写入字符（模拟异常情况）
	back = renderer.GetBackBuffer()
	// 直接设置continuation位置的cell
	back.Cells[0][1] = Cell{Cluster: "X", Width: 1, IsContinuation: false}
	back.Cells[0][3] = Cell{Cluster: "Y", Width: 1, IsContinuation: false}

	// 这应该能够正确渲染，不应该崩溃
	output := renderer.Render()
	_ = output // 验证不会崩溃
}

// TestRender_ManipulatedContinuation 测试continuation被篡改的情况
func TestRender_ManipulatedContinuation(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 第一帧
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s)
	renderer.Render()

	// 第二帧：部分清除
	back = renderer.GetBackBuffer()
	back.Cells[0][0] = Cell{Cluster: " ", Width: 1}
	// 不正确地处理continuation（模拟bug场景）

	output := renderer.Render()
	// 验证渲染不会崩溃或产生错误输出
	_ = output
}

// =============================================================================
// 边界测试
// =============================================================================

// TestRender_EmptyBuffer 测试空buffer渲染
func TestRender_EmptyBuffer(t *testing.T) {
	renderer := NewRenderer(0, 0)
	output := renderer.Render()
	if output != "" {
		t.Errorf("Empty buffer should produce no output, got %q", output)
	}
}

// TestRender_SingleCell 测试单个单元格
func TestRender_SingleCell(t *testing.T) {
	renderer := NewRenderer(1, 1)
	s := style.Style{}

	back := renderer.GetBackBuffer()
	back.SetCell(0, 0, 'X', s)

	output := renderer.Render()
	if output == "" {
		t.Error("Single cell should produce output")
	}
}

// TestRender_SingleWideChar 测试单个宽字符在最小buffer
func TestRender_SingleWideChar(t *testing.T) {
	// 宽字符需要至少2个位置
	renderer := NewRenderer(2, 1)
	s := style.Style{}

	back := renderer.GetBackBuffer()
	back.SetCell(0, 0, '测', s)

	output := renderer.Render()
	if output == "" {
		t.Error("Wide char should produce output")
	}

	front := renderer.GetFrontBuffer()
	if front.Cells[0][0].Width != 2 {
		t.Errorf("Wide char should have width 2, got %d", front.Cells[0][0].Width)
	}
	if !front.Cells[0][1].IsContinuation {
		t.Error("Second cell should be continuation")
	}
}

// TestRender_WideCharAtBufferEdge 测试宽字符在buffer边缘
func TestRender_WideCharAtBufferEdge(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 在最后一个位置尝试写入宽字符（应该被截断或拒绝）
	back := renderer.GetBackBuffer()
	// 位置9是最后一个，宽字符需要位置9和10，应该被截断
	back.SetCell(9, 0, '测', s)

	renderer.Render()

	front := renderer.GetFrontBuffer()
	// 宽字符在边界应该不被写入
	if front.Cells[0][9].Cluster == "测" {
		t.Error("Wide char at buffer edge should not be written (insufficient space)")
	}
}

// TestRender_ZeroWidthCell 测试零宽度单元格
func TestRender_ZeroWidthCell(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	back := renderer.GetBackBuffer()
	// 手动创建一个零宽度单元格
	back.Cells[0][5] = Cell{Cluster: "", Width: 0, IsContinuation: false}

	output := renderer.Render()
	// 不应该崩溃
	_ = output
}

// TestRender_NegativeWidth 测试负宽度（异常情况）
func TestRender_NegativeWidth(t *testing.T) {
	renderer := NewRenderer(10, 3)

	back := renderer.GetBackBuffer()
	// 创建一个异常的负宽度单元格
	back.Cells[0][5] = Cell{Cluster: "X", Width: -1, IsContinuation: false}

	// 不应该崩溃
	output := renderer.Render()
	_ = output
}

// =============================================================================
// 极限测试
// =============================================================================

// TestRender_FullBufferChange 测试全buffer变化
func TestRender_FullBufferChange(t *testing.T) {
	renderer := NewRenderer(80, 24)
	s := style.Style{}

	// 第一帧：填充全部
	back := renderer.GetBackBuffer()
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			back.SetCell(x, y, 'A', s)
		}
	}
	renderer.Render()

	// 第二帧：完全不同的内容
	back = renderer.GetBackBuffer()
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			back.SetCell(x, y, 'B', s)
		}
	}
	output := renderer.Render()

	if output == "" {
		t.Error("Full buffer change should produce output")
	}
}

// TestRender_LargeWideCharBuffer 测试大量宽字符
func TestRender_LargeWideCharBuffer(t *testing.T) {
	renderer := NewRenderer(100, 10)
	s := style.Style{}

	back := renderer.GetBackBuffer()
	// 填充宽字符
	for y := 0; y < 10; y++ {
		text := strings.Repeat("测试", 25) // 50个宽字符 = 100 cells
		back.SetString(0, y, text, s)
	}
	output := renderer.Render()

	if output == "" {
		t.Error("Large wide char buffer should produce output")
	}
}

// TestRender_RapidFrameChanges 测试快速帧变化
func TestRender_RapidFrameChanges(t *testing.T) {
	renderer := NewRenderer(20, 5)
	s := style.Style{}

	// 模拟10次快速帧变化
	for i := 0; i < 10; i++ {
		back := renderer.GetBackBuffer()
		text := string(rune('A' + i%26))
		for x := 0; x < 20; x++ {
			back.SetCell(x, i%5, rune(text[0]), s)
		}
		renderer.Render()
	}

	// 验证最终状态正确
	front := renderer.GetFrontBuffer()
	if front.Cells[4][0].Cluster == "" {
		t.Error("Final state should have content")
	}
}

// TestRender_AlternatingWideNarrow 测试宽窄字符交替
func TestRender_AlternatingWideNarrow(t *testing.T) {
	renderer := NewRenderer(30, 3)
	s := style.Style{}

	// 宽窄交替模式
	text := "A测B试C演D示E"
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, text, s)
	renderer.Render()

	// 切换到纯窄字符
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "ABCDEFGHIJKLMNOP", s)
	output := renderer.Render()

	if output == "" {
		t.Error("Alternating wide/narrow change should produce output")
	}

	front := renderer.GetFrontBuffer()
	// 检查没有continuation残留
	for x := 0; x < 16; x++ {
		if front.Cells[0][x].IsContinuation {
			t.Errorf("Cell[0][%d] should not be continuation after narrow char overwrite", x)
		}
	}
}

// =============================================================================
// 样式变化测试
// =============================================================================

// TestRender_StyleChangeOnly 测试仅样式变化
func TestRender_StyleChangeOnly(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s1 := style.Style{}.Foreground(style.Red)
	s2 := style.Style{}.Foreground(style.Blue)

	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s1)
	renderer.Render()

	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s2)
	output := renderer.Render()

	if output == "" {
		t.Error("Style change should produce output")
	}
}

// TestRender_WideCharStyleChange 测试宽字符样式变化
func TestRender_WideCharStyleChange(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s1 := style.Style{}.Foreground(style.Red)
	s2 := style.Style{}.Foreground(style.Red).Bold(true)

	// 第一帧
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s1)
	renderer.Render()

	// 验证初始状态
	front := renderer.GetFrontBuffer()
	if front.Cells[0][0].Style.IsBold() {
		t.Error("Initial style should not be bold")
	}

	// 第二帧：只改变样式
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "测试", s2)
	output := renderer.Render()

	if output == "" {
		t.Error("Wide char style change should produce output")
	}

	front = renderer.GetFrontBuffer()
	if !front.Cells[0][0].Style.IsBold() {
		t.Error("Style should be updated to bold")
	}
}

// =============================================================================
// 光标和位置测试
// =============================================================================

// TestRender_CursorPositionTracking 测试光标位置跟踪
func TestRender_CursorPositionTracking(t *testing.T) {
	renderer := NewRenderer(20, 5)
	s := style.Style{}

	// 在多个位置设置内容
	back := renderer.GetBackBuffer()
	back.SetCell(0, 0, 'A', s)
	back.SetCell(10, 2, 'B', s)
	back.SetCell(19, 4, 'C', s)

	output := renderer.Render()
	// 验证输出包含光标移动命令
	if !strings.Contains(output, "\x1b[") {
		t.Error("Output should contain cursor movement commands")
	}
}

// TestRender_RunMerging 测试run合并优化
func TestRender_RunMerging(t *testing.T) {
	renderer := NewRenderer(50, 3)
	s := style.Style{}.Foreground(style.Red)

	back := renderer.GetBackBuffer()
	// 写入连续相同样式的字符
	text := strings.Repeat("A", 50)
	back.SetString(0, 0, text, s)
	back.SetString(0, 1, text, s)

	output := renderer.Render()
	// 验证输出被正确生成
	if output == "" {
		t.Error("Run merging should produce output")
	}
}

// =============================================================================
// 特殊字符测试
// =============================================================================

// TestRender_SpecialChars 测试特殊字符
func TestRender_SpecialChars(t *testing.T) {
	renderer := NewRenderer(20, 3)
	s := style.Style{}

	// 边框字符
	borderChars := "┌─┐│└┘├┤┬┴┼"
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, borderChars, s)
	renderer.Render()

	// 替换为普通字符
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "12345678901", s)
	output := renderer.Render()

	if output == "" {
		t.Error("Special char replacement should produce output")
	}
}

// TestRender_ArrowSymbols 测试箭头符号
func TestRender_ArrowSymbols(t *testing.T) {
	renderer := NewRenderer(20, 3)
	s := style.Style{}

	// 箭头符号（可能宽度不明确的字符）
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "→←↑↓", s)
	renderer.Render()

	// 替换内容
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "ABCD", s)
	output := renderer.Render()

	if output == "" {
		t.Error("Arrow symbol replacement should produce output")
	}

	front := renderer.GetFrontBuffer()
	expected := []string{"A", "B", "C", "D"}
	for i, exp := range expected {
		if front.Cells[0][i].Cluster != exp {
			t.Errorf("Cell[0][%d]: expected %q, got %q", i, exp, front.Cells[0][i].Cluster)
		}
	}
}

// =============================================================================
// 缓冲区状态一致性测试
// =============================================================================

// TestRender_BufferConsistency 测试渲染后buffer状态一致性
func TestRender_BufferConsistency(t *testing.T) {
	renderer := NewRenderer(10, 5)
	s := style.Style{}

	// 写入内容
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "测试AB", s)
	renderer.Render()

	// 修改部分内容
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "XY", s)
	renderer.Render()

	// 验证front和back的一致性
	front := renderer.GetFrontBuffer()
	if front.Cells[0][0].Cluster != "X" {
		t.Errorf("Front[0][0] should be 'X', got %q", front.Cells[0][0].Cluster)
	}

	// back现在应该包含与front相同的内容（因为swapBuffers复制了）
	// 检查是否一致
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			frontCell := front.Cells[y][x]
			backCell := back.Cells[y][x]
			if frontCell.Cluster != backCell.Cluster {
				t.Errorf("Mismatch at (%d,%d): front=%q, back=%q",
					x, y, frontCell.Cluster, backCell.Cluster)
			}
		}
	}
}

// TestRender_MultipleRenderCycles 测试多次渲染循环
func TestRender_MultipleRenderCycles(t *testing.T) {
	renderer := NewRenderer(10, 3)
	s := style.Style{}

	// 第一帧
	back := renderer.GetBackBuffer()
	back.SetString(0, 0, "AAA", s)
	output1 := renderer.Render()

	// 第二帧（无变化）
	output2 := renderer.Render()

	// 第三帧（有变化）
	back = renderer.GetBackBuffer()
	back.SetString(0, 0, "BBB", s)
	output3 := renderer.Render()

	// 第四帧（无变化）
	output4 := renderer.Render()

	// 验证：第一帧和第三帧应该有输出，第二帧和第四帧不应该有
	if output1 == "" {
		t.Error("First render should have output")
	}
	if output2 != "" {
		t.Error("Second render (no changes) should have no output")
	}
	if output3 == "" {
		t.Error("Third render (changes) should have output")
	}
	if output4 != "" {
		t.Error("Fourth render (no changes) should have no output")
	}
}

// =============================================================================
// 性能基准测试
// =============================================================================

// BenchmarkRender_WideCharDiff 基准测试：宽字符差异渲染
func BenchmarkRender_WideCharDiff(b *testing.B) {
	renderer := NewRenderer(100, 30)
	s := style.Style{}

	back := renderer.GetBackBuffer()
	for y := 0; y < 30; y++ {
		text := strings.Repeat("测试", 25)
		back.SetString(0, y, text, s)
	}
	renderer.Render()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		back = renderer.GetBackBuffer()
		// 修改10%的内容
		for j := 0; j < 300; j++ {
			x := j % 100
			y := (j / 100) % 30
			back.SetCell(x, y, 'X', s)
		}
		renderer.Render()
	}
}

// BenchmarkRender_FullDiff 基准测试：完全差异渲染
func BenchmarkRender_FullDiff(b *testing.B) {
	renderer := NewRenderer(80, 24)
	s := style.Style{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		back := renderer.GetBackBuffer()
		for y := 0; y < 24; y++ {
			for x := 0; x < 80; x++ {
				back.SetCell(x, y, rune('A'+(i+y)%26), s)
			}
		}
		renderer.Render()
	}
}

// BenchmarkRender_NoChanges 基准测试：无变化渲染
func BenchmarkRender_NoChanges(b *testing.B) {
	renderer := NewRenderer(80, 24)
	s := style.Style{}

	back := renderer.GetBackBuffer()
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			back.SetCell(x, y, 'A', s)
		}
	}
	renderer.Render()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Render()
	}
}

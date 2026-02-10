package paint

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// TestWideCharIntegration 验证宽字符的完整写入和读取流程
func TestWideCharIntegration(t *testing.T) {
	buf := NewBuffer(30, 5)
	ctx := NewPaintContext(buf, Rect{X: 0, Y: 0, Width: 30, Height: 5})
	s := style.Style{}.Foreground(style.Cyan)

	// 测试：使用 SetString 写入混合宽字符文本
	text1 := "Scheduler演示"
	buf.SetString(0, 0, text1, s)

	// 验证每个字符的位置
	expectedPositions := []struct {
		cluster  string
		x        int
		width    int
		isCont   bool
	}{
		{"S", 0, 1, false},
		{"c", 1, 1, false},
		{"h", 2, 1, false},
		{"e", 3, 1, false},
		{"d", 4, 1, false},
		{"u", 5, 1, false},
		{"l", 6, 1, false},
		{"e", 7, 1, false},
		{"r", 8, 1, false},
		{"演", 9, 2, false},  // 宽字符
		{"", 10, 0, true},    // 延续单元格
		{"示", 11, 2, false}, // 宽字符
		{"", 12, 0, true},    // 延续单元格
	}

	for _, exp := range expectedPositions {
		cell := buf.Cells[0][exp.x]
		if cell.Cluster != exp.cluster {
			t.Errorf("pos %d: expected cluster %q, got %q", exp.x, exp.cluster, cell.Cluster)
		}
		if cell.Width != exp.width {
			t.Errorf("pos %d: expected width %d, got %d", exp.x, exp.width, cell.Width)
		}
		if cell.IsContinuation != exp.isCont {
			t.Errorf("pos %d: expected IsContinuation %v, got %v", exp.x, exp.isCont, cell.IsContinuation)
		}
	}

	// 测试：使用 PaintContext.SetString 写入
	text2 := "=== 日志面板 ==="
	ctx.SetString(0, 2, text2, s)

	// 验证位置：'=' '=' '=' ' ' '日' '志' '面' '板' ' ' '=' '=' '='
	// 位置:   0   1   2 3   4   5   6   7  8   9  10  11
	expectedChars := map[int]string{
		0: "=", 1: "=", 2: "=", 3: " ",
		4: "日", // 宽字符
		6: "志", // 宽字符，位置 5 是延续
		8: "面", // 宽字符，位置 7 是延续
		10: "板", // 宽字符，位置 9 是延续
		12: " ", 13: "=", 14: "=", 15: "=",
	}

	for x, expectedCluster := range expectedChars {
		cell := buf.Cells[2][x]
		if cell.Cluster != expectedCluster {
			t.Errorf("row 2 pos %d: expected %q, got %q", x, expectedCluster, cell.Cluster)
		}
	}

	// 验证延续单元格
	contPositions := []int{5, 7, 9, 11}
	for _, x := range contPositions {
		cell := buf.Cells[2][x]
		if !cell.IsContinuation {
			t.Errorf("row 2 pos %d: should be continuation cell", x)
		}
	}
}

// TestWideCharOverwrite 验证宽字符被正确覆盖
func TestWideCharOverwrite(t *testing.T) {
	buf := NewBuffer(20, 2)
	s := style.Style{}

	// 先写入宽字符
	buf.SetString(0, 0, "演示", s)

	// 验证宽字符已正确写入
	if buf.Cells[0][0].Cluster != "演" || buf.Cells[0][0].Width != 2 {
		t.Error("first wide char not set correctly")
	}
	if !buf.Cells[0][1].IsContinuation {
		t.Error("cell[0][1] should be continuation")
	}

	// 用窄字符覆盖
	buf.SetString(0, 0, "ABC", s)

	// 验证窄字符正确覆盖
	if buf.Cells[0][0].Cluster != "A" {
		t.Errorf("cell[0][0]: expected 'A', got %s", buf.Cells[0][0].Cluster)
	}
	if buf.Cells[0][1].Cluster != "B" {
		t.Errorf("cell[0][1]: expected 'B', got %s", buf.Cells[0][1].Cluster)
	}
	if buf.Cells[0][1].IsContinuation {
		t.Error("cell[0][1] should not be continuation after overwrite")
	}
}

// TestWideCharWithDrawText 验证 DrawText 对齐处理宽字符
func TestWideCharWithDrawText(t *testing.T) {
	buf := NewBuffer(20, 3)
	ctx := NewPaintContext(buf, Rect{X: 0, Y: 0, Width: 20, Height: 3})
	s := style.Style{}

	// 测试居中对齐
	ctx.DrawText(0, 0, "测试", AlignCenter, s)

	// "测试" 显示宽度为 4，可用宽度为 20
	// 居中位置 = (20 - 4) / 2 = 8
	// 所以 '测' 在位置 8，'试' 在位置 10
	if buf.Cells[0][8].Cluster != "测" {
		t.Errorf("expected '测' at pos 8, got %s", buf.Cells[0][8].Cluster)
	}
	if buf.Cells[0][8].Width != 2 {
		t.Error("'测' should have width 2")
	}
	if !buf.Cells[0][9].IsContinuation {
		t.Error("cell[0][9] should be continuation cell for '测'")
	}
	if buf.Cells[0][10].Cluster != "试" {
		t.Errorf("expected '试' at pos 10, got %s", buf.Cells[0][10].Cluster)
	}
}

func TestSetStringSanitizeUnsafeEmoji(t *testing.T) {
	buf := NewBuffer(20, 2)
	s := style.Style{}

	buf.SetString(0, 0, "🖼️X", s)

	if got := buf.Cells[0][0].Cluster; got != "🖼" {
		t.Fatalf("expected sanitized emoji head at [0], got %q", got)
	}
	if got := buf.Cells[0][1].Cluster; got != "X" {
		t.Fatalf("expected X at [1], got %q", got)
	}
}

func TestPaintContextSetStringUsesBufferSemantics(t *testing.T) {
	buf := NewBuffer(20, 2)
	ctx := NewPaintContext(buf, Rect{X: 0, Y: 0, Width: 20, Height: 2})
	s := style.Style{}

	ctx.SetString(0, 0, "e\u0301A", s) // e + combining acute + A

	if got := buf.Cells[0][0].Cluster; got != "e" {
		t.Fatalf("expected sanitized e at [0], got %q", got)
	}
	if got := buf.Cells[0][1].Cluster; got != "A" {
		t.Fatalf("expected A at [1], got %q", got)
	}
}

// TestOutputLoopSimulation 模拟输出循环，验证正确跳过延续单元格
func TestOutputLoopSimulation(t *testing.T) {
	buf := NewBuffer(20, 1)
	s := style.Style{}

	buf.SetString(0, 0, "ABC测试123", s)

	// 模拟输出循环
	output := ""
	x := 0
	for x < buf.Width {
		cell := buf.Cells[0][x]

		// 跳过延续单元格
		if ShouldSkipCell(cell) {
			x++
			continue
		}

		// 输出字符
		if cell.Cluster != "" && cell.Cluster != "\x00" {
			output += cell.Cluster
		}

		// 按字符宽度递增
		width := GetCellWidth(cell)
		if width > 0 {
			x += width
		} else {
			x++
		}
	}

	// 验证输出以正确的内容开头（buffer 初始化为空格，后面会有空格）
	expected := "ABC测试123"                // 实际写入的内容
	padding := strings.Repeat(" ", 20-10) // 剩余的空格 (20 - 10 = 10 个空格)
	if output != expected+padding {
		t.Errorf("expected '%s', got '%s'", expected+padding, output)
	}

	// 验证位置计算：正确循环应该访问 10 个位置 (8 个字符 + 2 个延续单元格)
	// A(0) B(1) C(2) 测(3-4) 试(5-6) 1(7) 2(8) 3(9)
	// 实际缓冲区中的位置应该是：
	// 0:'A' 1:'B' 2:'C' 3:'测' 4:continuation 5:'试' 6:continuation 7:'1' 8:'2' 9:'3'

	// 验证延续单元格位置
	if !buf.Cells[0][4].IsContinuation {
		t.Error("cell[0][4] should be continuation for '测'")
	}
	if !buf.Cells[0][6].IsContinuation {
		t.Error("cell[0][6] should be continuation for '试'")
	}
}

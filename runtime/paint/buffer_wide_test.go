package paint

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestSetCellWideChar(t *testing.T) {
	buf := NewBuffer(10, 2)
	s := style.Style{}.Foreground(style.Red)

	// 写入宽字符 "演" (中文，宽度为2)
	buf.SetCell(0, 0, '演', s)

	// 验证主单元格
	cell := buf.Cells[0][0]
	if cell.Cluster != "演" {
		t.Errorf("expected cluster '演', got %s", cell.Cluster)
	}
	if cell.Width != 2 {
		t.Errorf("expected width 2, got %d", cell.Width)
	}
	if cell.IsContinuation {
		t.Error("main cell should not be marked as continuation")
	}

	// 验证延续单元格
	contCell := buf.Cells[0][1]
	if contCell.Cluster != "" {
		t.Errorf("continuation cell should have no cluster, got %s", contCell.Cluster)
	}
	if !contCell.IsContinuation {
		t.Error("cell[0][1] should be marked as continuation")
	}
}

func TestSetStringWideChar(t *testing.T) {
	buf := NewBuffer(20, 2)
	s := style.Style{}.Foreground(style.Cyan)

	// 写入包含宽字符的字符串 "Scheduler演示"
	text := "Scheduler演示"
	buf.SetString(0, 0, text, s)

	// 验证每个字符
	expectedRunes := []rune(text)
	col := 0
	for _, r := range expectedRunes {
		if col >= buf.Width {
			t.Fatalf("column %d out of bounds", col)
		}
		cell := buf.Cells[0][col]
		if cell.Cluster != string(r) {
			t.Errorf("cell[0][%d]: expected char %c, got %s", col, r, cell.Cluster)
		}
		// 验证非延续单元格
		if cell.IsContinuation {
			t.Errorf("cell[0][%d]: should not be continuation", col)
		}
		width := runeWidth(r)
		if cell.Width != width {
			t.Errorf("cell[0][%d]: expected width %d, got %d", col, width, cell.Width)
		}
		// 如果是宽字符，验证下一个单元格是延续
		if width == 2 && col+1 < buf.Width {
			contCell := buf.Cells[0][col+1]
			if !contCell.IsContinuation {
				t.Errorf("cell[0][%d]: next cell should be continuation", col+1)
			}
		}
		col += width
	}
}

func TestSetStringOverwriteWideChar(t *testing.T) {
	buf := NewBuffer(20, 2)
	s := style.Style{}

	// 先写入宽字符
	buf.SetString(0, 0, "演示", s)

	// 验证 "演" (宽度2) 和 "示" (宽度2)
	if (buf.Cells[0][0].Cluster != "演") || buf.Cells[0][0].Width != 2 {
		t.Error("first wide char not set correctly")
	}
	if !buf.Cells[0][1].IsContinuation {
		t.Error("cell[0][1] should be continuation")
	}
	if (buf.Cells[0][2].Cluster != "示") || buf.Cells[0][2].Width != 2 {
		t.Error("second wide char not set correctly")
	}
	if !buf.Cells[0][3].IsContinuation {
		t.Error("cell[0][3] should be continuation")
	}

	// 用窄字符覆盖
	buf.SetString(0, 0, "ABC", s)

	// 验证窄字符正确覆盖了宽字符
	if (buf.Cells[0][0].Cluster != "A") || buf.Cells[0][0].Width != 1 {
		t.Errorf("cell[0][0]: expected 'A' with width 1, got %s with width %d",
			buf.Cells[0][0].Cluster, buf.Cells[0][0].Width)
	}
	if (buf.Cells[0][1].Cluster != "B") {
		t.Errorf("cell[0][1]: expected 'B', got %s", buf.Cells[0][1].Cluster)
	}
	if buf.Cells[0][1].IsContinuation {
		t.Error("cell[0][1] should not be continuation after overwrite")
	}
	if (buf.Cells[0][2].Cluster != "C") {
		t.Errorf("cell[0][2]: expected 'C', got %s", buf.Cells[0][2].Cluster)
	}
}

func TestIsCellChanged(t *testing.T) {
	s1 := style.Style{}.Foreground(style.Red)
	s2 := style.Style{}.Foreground(style.Blue)

	tests := []struct {
		name     string
		cell     Cell
		prevCell Cell
		want     bool
	}{
		{
			name:     "both empty",
			cell:     Cell{},
			prevCell: Cell{},
			want:     false,
		},
		{
			name:     "char changed",
			cell:     Cell{Cluster: "A"},
			prevCell: Cell{Cluster: "B"},
			want:     true,
		},
		{
			name:     "style changed",
			cell:     Cell{Cluster: "A", Style: s1},
			prevCell: Cell{Cluster: "A", Style: s2},
			want:     true,
		},
		{
			name:     "cell is continuation - should skip",
			cell:     Cell{Cluster: "A", IsContinuation: true},
			prevCell: Cell{},
			want:     false,
		},
		{
			name:     "prev is continuation - should skip",
			cell:     Cell{Cluster: "A"},
			prevCell: Cell{Cluster: "B", IsContinuation: true},
			want:     false,
		},
		{
			name:     "both continuation - should skip",
			cell:     Cell{IsContinuation: true},
			prevCell: Cell{IsContinuation: true},
			want:     false,
		},
		{
			name:     "same char and style",
			cell:     Cell{Cluster: "A", Style: s1},
			prevCell: Cell{Cluster: "A", Style: s1},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCellChanged(tt.cell, tt.prevCell); got != tt.want {
				t.Errorf("IsCellChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCellWidth(t *testing.T) {
	tests := []struct {
		name string
		cell Cell
		want int
	}{
		{
			name: "normal char",
			cell: Cell{Cluster: "A", Width: 1},
			want: 1,
		},
		{
			name: "wide char",
			cell: Cell{Cluster: "演", Width: 2},
			want: 2,
		},
		{
			name: "continuation cell",
			cell: Cell{IsContinuation: true, Width: 0},
			want: 0,
		},
		{
			name: "empty cell",
			cell: Cell{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCellWidth(tt.cell); got != tt.want {
				t.Errorf("GetCellWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldSkipCell(t *testing.T) {
	tests := []struct {
		name string
		cell Cell
		want bool
	}{
		{
			name: "normal cell",
			cell: Cell{Cluster: "A"},
			want: false,
		},
		{
			name: "wide char main cell",
			cell: Cell{Cluster: "演", Width: 2},
			want: false,
		},
		{
			name: "continuation cell",
			cell: Cell{IsContinuation: true},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldSkipCell(tt.cell); got != tt.want {
				t.Errorf("ShouldSkipCell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClearWideChar(t *testing.T) {
	buf := NewBuffer(10, 2)
	s := style.Style{}

	// 写入宽字符
	buf.SetCell(0, 0, '演', s)

	// 验证已设置
	if (buf.Cells[0][0].Cluster != "演") {
		t.Fatal("wide char not set")
	}
	if !buf.Cells[0][1].IsContinuation {
		t.Fatal("continuation not set")
	}

	// 清除宽字符
	buf.ClearWideChar(0, 0)

	// 验证主单元格已清除
	if (buf.Cells[0][0].Cluster != "" && buf.Cells[0][0].Cluster != "\x00") {
		t.Errorf("main cell not cleared, got %s", buf.Cells[0][0].Cluster)
	}

	// 验证延续单元格已清除
	if buf.Cells[0][1].IsContinuation {
		t.Error("continuation flag not cleared")
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		r    rune
		want int
	}{
		{'A', 1},
		{' ', 1},
		{'0', 1},
		{'演', 2},       // 中文
		{'示', 2},       // 中文
		{'あ', 2},       // 日文
		{'가', 2},       // 韩文
		{'😀', 2},       // Emoji
		{'🎉', 2},       // Emoji
		{'\u0300', 1},   // Combining accent
		{'\u200d', 1},   // Zero width joiner
	}

	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			if got := runeWidth(tt.r); got != tt.want {
				t.Errorf("runeWidth(%c) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

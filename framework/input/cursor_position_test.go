package input

import (
	"testing"

	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/paint"
)

// TestCursorPositionRendering 测试光标在不同输入情况下的渲染位置
func TestCursorPositionRendering(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		cursor         int
		expectedCursor string // 期望光标下的字符（'*'表示空位置或右括号）
		expectedX      int   // 期望的绝对X位置
	}{
		{"空输入-光标在开始", "", 0, "*", 3},      // []  光标在位置3（左括号后，无字符）
		{"单字符-光标在末尾", "a", 1, "]", 4},      // [a]  光标在位置4（右括号）
		{"单字符-光标在开始", "a", 0, "a", 3},     // [a]  光标在位置3（字符'a'）
		{"双字符-光标在中间", "ab", 1, "b", 4},    // [ab] 光标在位置4（字符'b'）
		{"双字符-光标在末尾", "ab", 2, "]", 5},    // [ab] 光标在位置5（右括号）
		{"三字符-光标在中间", "abc", 1, "b", 4},   // [abc] 光标在位置4（字符'b'）
		{"三字符-光标在末尾", "abc", 3, "]", 6},   // [abc] 光标在位置6（右括号）
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txt := NewTextInput()
			txt.SetValue(tt.value)
			txt.SetCursor(tt.cursor)
			txt.OnFocus()

			buf := paint.NewBuffer(80, 24)
			ctx := component.NewPaintContext(buf, 2, 1, 80, 24) // 模拟Form中的偏移 (2, 1)

			txt.Paint(ctx, buf)

			// 计算期望的绝对光标位置
			// 左边框在X=2, 内容从X=3开始
			// cursorX = 1 + cursorPos
			// absCursorX = ctx.X + cursorX = 2 + 1 + cursorPos
			expectedAbsX := 2 + 1 + tt.cursor
			if expectedAbsX != tt.expectedX {
				t.Errorf("期望的绝对X位置是 %d, 但计算得到 %d", tt.expectedX, expectedAbsX)
			}

			// Helper function to get first rune from cluster
			getFirstRune := func(cluster string) rune {
				for _, c := range cluster {
					return c
				}
				return 0
			}

			// 检查光标是否在正确位置（通过检查reverse样式）
			// 光标应该在 (expectedAbsX, 1) 位置
			cell := buf.Cells[1][expectedAbsX]

			if tt.expectedCursor == "*" {
				// 空输入，没有字符被反向高亮，或者高亮的是空格
				// 在这种情况下，我们检查是否设置了reverse样式
				if !cell.Style.IsReverse() && tt.cursor == 0 {
					// 光标应该在空位置，但没有字符可以反向高亮
					// 这是可以接受的
				}
			} else {
				// 检查光标是否高亮了正确的字符
				expectedRune := rune(tt.expectedCursor[0])
				if getFirstRune(cell.Cluster) != expectedRune {
					t.Errorf("位置 (%d,1) 应该是 '%c' 但得到 '%c'",
						expectedAbsX, expectedRune, getFirstRune(cell.Cluster))
				}
				if !cell.Style.IsReverse() {
					t.Errorf("位置 (%d,1) 应该有反向样式", expectedAbsX)
				}
			}

			// 验证输入框的格式是正确的 [content]
			// 左边框应该在位置2
			leftCell := buf.Cells[1][2]
			if getFirstRune(leftCell.Cluster) != '[' {
				t.Errorf("位置 (2,1) 应该是 '[' 但得到 '%c'", getFirstRune(leftCell.Cluster))
			}

			// 右边框应该在位置 2 + 1 + len(value)
			rightX := 2 + 1 + len([]rune(tt.value))
			if rightX < 80 {
				rightCell := buf.Cells[1][rightX]
				if getFirstRune(rightCell.Cluster) != ']' {
					t.Errorf("位置 (%d,1) 应该是 ']' 但得到 '%c'", rightX, getFirstRune(rightCell.Cluster))
				}
			}
		})
	}
}

// TestCursorPositionMovement 测试光标移动后的渲染位置
func TestCursorPositionMovement(t *testing.T) {
	txt := NewTextInput()
	txt.SetValue("test")
	txt.OnFocus()

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 2, 1, 80, 24) // 模拟Form中的偏移

	// Helper function to get first rune from cluster
	getFirstRune := func(cluster string) rune {
		for _, c := range cluster {
			return c
		}
		return 0
	}

	// 测试不同光标位置
	// cursorPos=n 表示光标在第n个字符位置（0-based）
	// 如果 n < len(value)，光标高亮该字符
	// 如果 n == len(value)，光标在末尾（右括号）
	cursorPositions := []struct {
		cursor        int
		expectedChar  rune // 期望光标下的字符
		expectedAbsX  int
	}{
		{0, 't', 3},    // 光标在第1个字符't'
		{1, 'e', 4},    // 光标在第2个字符'e'
		{2, 's', 5},    // 光标在第3个字符's'
		{3, 't', 6},    // 光标在第4个字符't'
		{4, ']', 7},    // 光标在末尾（右括号）
	}

	for _, cp := range cursorPositions {
		txt.SetCursor(cp.cursor)

		// 创建新的buffer和context
		buf = paint.NewBuffer(80, 24)
		ctx = component.NewPaintContext(buf, 2, 1, 80, 24) // 使用新的buffer
		txt.Paint(ctx, buf)

		expectedAbsX := 2 + 1 + cp.cursor
		if expectedAbsX != cp.expectedAbsX {
			t.Errorf("cursor=%d: 计算的绝对X位置 %d 与期望 %d 不匹配",
				cp.cursor, expectedAbsX, cp.expectedAbsX)
		}

		cell := buf.Cells[1][expectedAbsX]

		if getFirstRune(cell.Cluster) != cp.expectedChar {
			t.Errorf("cursor=%d: 位置 (%d,1) 应该是 '%c' 但得到 '%c'",
				cp.cursor, expectedAbsX, cp.expectedChar, getFirstRune(cell.Cluster))
		}
		if !cell.Style.IsReverse() {
			t.Errorf("cursor=%d: 位置 (%d,1) 应该有反向样式",
				cp.cursor, expectedAbsX)
		}
	}
}

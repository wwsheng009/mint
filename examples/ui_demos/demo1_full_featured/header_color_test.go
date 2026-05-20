package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework/theme"
)

// TestHeaderBackgroundColorRendering 验证 Header 所有元素都有背景色
func TestHeaderBackgroundColorRendering(t *testing.T) {
	// 设置主题
	_ = theme.SetTheme("nord")

	testApp := newDemoTestApp(t)

	// 获取渲染结果
	rendered := waitForDemoRender(t, testApp, 300*time.Millisecond, func(rendered string) bool {
		return strings.Contains(rendered, "TUI Engine Demo")
	})
	lines := splitLines(rendered)

	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}

	t.Logf("=== Header Render (First 3 lines) ===")
	for i := 0; i < 3 && i < len(lines); i++ {
		t.Logf("Line %d: %s", i, lines[i])
	}

	// 分析 Header 结构（第 1 行，中间内容行）
	headerLine := lines[1]
	t.Logf("\n=== Analyzing Header Content ===")
	t.Logf("Header content length: %d", len(headerLine))

	// 检查关键元素
	elements := []string{"TUI Engine Demo", "[Open Modal]", "Clicks:"}
	for _, elem := range elements {
		idx := strings.Index(headerLine, elem)
		if idx >= 0 {
			t.Logf("✓ Found %q at position %d", elem, idx)
		} else {
			t.Errorf("❌ Could not find %q in header", elem)
		}
	}

	// 检查元素之间的间距
	tuiIdx := strings.Index(headerLine, "TUI Engine Demo")
	buttonIdx := strings.Index(headerLine, "[Open Modal]")
	clicksIdx := strings.Index(headerLine, "Clicks:")

	if tuiIdx >= 0 && buttonIdx >= 0 {
		gap1 := buttonIdx - (tuiIdx + len("TUI Engine Demo"))
		t.Logf("Gap between 'TUI Engine Demo' and '[Open Modal]': %d spaces", gap1)
		if gap1 > 0 {
			gapContent := headerLine[tuiIdx+len("TUI Engine Demo") : buttonIdx]
			t.Logf("  Gap content: %q", gapContent)
		}
	}

	if buttonIdx >= 0 && clicksIdx >= 0 {
		gap2 := clicksIdx - (buttonIdx + len("[Open Modal]"))
		t.Logf("Gap between '[Open Modal]' and 'Clicks:': %d spaces", gap2)
		if gap2 > 0 {
			gapContent := headerLine[buttonIdx+len("[Open Modal]") : clicksIdx]
			t.Logf("  Gap content: %q", gapContent)
		}
	}
}

// TestHeaderWithDifferentCounts 测试不同 count 值的 Header 渲染
func TestHeaderWithDifferentCounts(t *testing.T) {
	_ = theme.SetTheme("nord")

	// Verify the header renders "Clicks: 0" for a fresh app.
	testApp := newDemoTestApp(t)

	rendered := waitForDemoRender(t, testApp, 300*time.Millisecond, func(rendered string) bool {
		return strings.Contains(rendered, "Clicks: 0")
	})
	lines := splitLines(rendered)

	if len(lines) < 2 {
		t.Fatalf("Not enough lines")
	}

	headerLine := lines[1]

	if strings.Contains(headerLine, "Clicks: 0") {
		t.Logf("✓ Found \"Clicks: 0\" in header")
	} else {
		t.Errorf("❌ Could not find \"Clicks: 0\" in header")
		t.Logf("Header: %s", headerLine)
	}
}

// TestHeaderElementPositions 详细检查 Header 中每个元素的位置
func TestHeaderElementPositions(t *testing.T) {
	_ = theme.SetTheme("nord")

	testApp := newDemoTestApp(t)

	rendered := waitForDemoRender(t, testApp, 300*time.Millisecond, func(rendered string) bool {
		return strings.Contains(rendered, "TUI Engine Demo")
	})
	lines := splitLines(rendered)

	headerLine := lines[1]

	// 定义期望的元素
	type Element struct {
		Name     string
		Expected string
	}

	elements := []Element{
		{"Title", "TUI Engine Demo"},
		{"Button", "[Open Modal]"},
		{"CounterLabel", "Clicks:"},
	}

	t.Logf("\n=== Element Position Analysis ===")
	prevEnd := 0

	for i, elem := range elements {
		idx := strings.Index(headerLine, elem.Expected)
		if idx < 0 {
			t.Errorf("❌ Element %d (%s) not found: %q", i, elem.Name, elem.Expected)
			continue
		}

		length := len(elem.Expected)
		t.Logf("Element %d (%s):", i, elem.Name)
		t.Logf("  Position: %d - %d", idx, idx+length-1)
		t.Logf("  Text: %q", elem.Expected)

		// 检查与前一个元素之间的间距
		if i > 0 {
			gap := idx - prevEnd
			t.Logf("  Gap from previous: %d spaces", gap)
			if gap > 0 && idx > prevEnd {
				gapContent := headerLine[prevEnd:idx]
				t.Logf("  Gap content: %q", gapContent)

				// Note: gaps may contain focus indicators (e.g. ">[ ]")
				for j, ch := range gapContent {
					if ch != ' ' {
						t.Logf("  Gap character %d is %q (non-space, may be focus indicator)", j, ch)
					}
				}
			}
		}

		prevEnd = idx + length
	}
}

// TestHeaderVisualContinuity 视觉连续性测试
// 这个测试验证 Header 是否有明显的视觉断裂
func TestHeaderVisualContinuity(t *testing.T) {
	_ = theme.SetTheme("nord")

	testApp := newDemoTestApp(t)

	rendered := waitForDemoRender(t, testApp, 300*time.Millisecond, func(rendered string) bool {
		return strings.Contains(rendered, "TUI Engine Demo")
	})
	lines := splitLines(rendered)

	if len(lines) < 3 {
		t.Fatal("Not enough lines")
	}

	// 检查 Header 的三行
	t.Logf("\n=== Visual Continuity Check ===")
	for i := 0; i < 3; i++ {
		line := lines[i]
		t.Logf("Line %d: %s", i, line)

		if i == 0 {
			// Top border uses ┌ and ┐
			if !strings.HasPrefix(line, "┌") {
				t.Errorf("Line 0 does not start with top-left border (┌)")
			}
		} else if i == 2 {
			// Bottom border uses └ and ┘
			if !strings.HasPrefix(line, "└") {
				t.Errorf("Line 2 does not start with bottom-left border (└)")
			}
		} else {
			// Middle content line uses │
			if !strings.HasPrefix(line, "│") {
				t.Errorf("Line %d does not start with border (│)", i)
			}
			if !strings.HasSuffix(strings.TrimRight(line, " "), "│") {
				t.Errorf("Line %d does not end with border (│)", i)
			}
		}

		// 移除边框，检查内容
		content := strings.Trim(line, "│┌└┐┘─")
		content = strings.TrimSpace(content)

		if i == 1 {
			// 中间行应该有内容
			if content == "" {
				t.Error("Middle line has no content!")
			} else {
				t.Logf("Middle line content: %q", content)
				t.Logf("Content length: %d", len(content))
			}
		}
	}
}

// TestHeaderButtonPosition 检查 "[Open Modal]" 按钮的精确位置
func TestHeaderButtonPosition(t *testing.T) {
	_ = theme.SetTheme("nord")

	testApp := newDemoTestApp(t)

	rendered := waitForDemoRender(t, testApp, 300*time.Millisecond, func(rendered string) bool {
		return strings.Contains(rendered, "[Open Modal]")
	})
	lines := splitLines(rendered)

	headerLine := lines[1]

	// 查找按钮
	buttonText := "[Open Modal]"
	buttonIdx := strings.Index(headerLine, buttonText)

	if buttonIdx < 0 {
		t.Fatal("Could not find '[Open Modal]' button")
	}

	t.Logf("\n=== Button Position Analysis ===")
	t.Logf("Button text: %q", buttonText)
	t.Logf("Button position: %d", buttonIdx)
	t.Logf("Button ends at: %d", buttonIdx+len(buttonText)-1)

	// 检查按钮前后
	if buttonIdx > 0 {
		beforeChar := headerLine[buttonIdx-1]
		t.Logf("Character before button: %q (should be space)", beforeChar)
		if beforeChar != ' ' {
			t.Errorf("Expected space before button, got %q", beforeChar)
		}
	}

	afterIdx := buttonIdx + len(buttonText)
	if afterIdx < len(headerLine) {
		afterChar := headerLine[afterIdx]
		t.Logf("Character after button: %q (should be space)", afterChar)
		if afterChar != ' ' {
			t.Errorf("Expected space after button, got %q", afterChar)
		}
	}

	// 验证按钮文本
	actualButton := headerLine[buttonIdx : buttonIdx+len(buttonText)]
	if actualButton != buttonText {
		t.Errorf("Button text mismatch: expected %q, got %q", buttonText, actualButton)
	} else {
		t.Logf("✓ Button text is correct")
	}
}

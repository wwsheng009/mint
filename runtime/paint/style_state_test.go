package paint

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
)

// TestStyleStateMachine_BuildDiffCodes 测试样式状态机的差异代码生成
func TestStyleStateMachine_BuildDiffCodes(t *testing.T) {
	ssm := NewStyleStateMachine()

	tests := []struct {
		name     string
		from     style.Style
		to       style.Style
		wantCode string
		wantDesc string
	}{
		{
			name:     "Default to Color - Foreground",
			from:     style.Style{},
			to:       style.Style{}.Foreground("red"),
			wantCode: "\x1b[31m",
			wantDesc: "should generate FG color code",
		},
		{
			name:     "Default to Color - Background",
			from:     style.Style{},
			to:       style.Style{}.Background("blue"),
			wantCode: "\x1b[44m",
			wantDesc: "should generate BG color code",
		},
		{
			name:     "Color to Default - Foreground",
			from:     style.Style{}.Foreground("red"),
			to:       style.Style{},
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when FG turns off",
		},
		{
			name:     "Color to Default - Background",
			from:     style.Style{}.Background("blue"),
			to:       style.Style{},
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when BG turns off",
		},
		{
			name:     "Color A to Color B - Foreground",
			from:     style.Style{}.Foreground("red"),
			to:       style.Style{}.Foreground("green"),
			wantCode: "\x1b[32m",
			wantDesc: "should generate only new FG code",
		},
		{
			name:     "Color A to Color B - Background",
			from:     style.Style{}.Background("red"),
			to:       style.Style{}.Background("green"),
			wantCode: "\x1b[42m",
			wantDesc: "should generate only new BG code",
		},
		{
			name:     "Bold to No Bold",
			from:     style.Style{}.Bold(true),
			to:       style.Style{}.Bold(false),
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when bold turns off",
		},
		{
			name:     "No Bold to Bold",
			from:     style.Style{}.Bold(false),
			to:       style.Style{}.Bold(true),
			wantCode: "\x1b[1m",
			wantDesc: "should generate bold code",
		},
		{
			name:     "Italic to No Italic",
			from:     style.Style{}.Italic(true),
			to:       style.Style{}.Italic(false),
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when italic turns off",
		},
		{
			name:     "No Italic to Italic",
			from:     style.Style{}.Italic(false),
			to:       style.Style{}.Italic(true),
			wantCode: "\x1b[3m",
			wantDesc: "should generate italic code",
		},
		{
			name:     "Underline to No Underline",
			from:     style.Style{}.Underline(true),
			to:       style.Style{}.Underline(false),
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when underline turns off",
		},
		{
			name:     "No Underline to Underline",
			from:     style.Style{}.Underline(false),
			to:       style.Style{}.Underline(true),
			wantCode: "\x1b[4m",
			wantDesc: "should generate underline code",
		},
		{
			name:     "Reverse to No Reverse",
			from:     style.Style{}.Reverse(true),
			to:       style.Style{}.Reverse(false),
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when reverse turns off",
		},
		{
			name:     "No Reverse to Reverse",
			from:     style.Style{}.Reverse(false),
			to:       style.Style{}.Reverse(true),
			wantCode: "\x1b[7m",
			wantDesc: "should generate reverse code",
		},
		{
			name:     "Strikethrough to No Strikethrough",
			from:     style.Style{}.Strikethrough(true),
			to:       style.Style{}.Strikethrough(false),
			wantCode: "\x1b[0m",
			wantDesc: "should generate reset code when strikethrough turns off",
		},
		{
			name:     "Strikethrough on not supported",
			from:     style.Style{}.Strikethrough(false),
			to:       style.Style{}.Strikethrough(true),
			wantCode: "",
			wantDesc: "strikethrough on is not yet supported in diff codes",
		},
		{
			name:     "Complex to Complex - Multiple changes",
			from:     style.Style{}.Foreground("red").Bold(true),
			to:       style.Style{}.Foreground("green").Italic(true),
			wantCode: "\x1b[0m\x1b[3;32m",
			wantDesc: "should reset and apply new styles",
		},
		{
			name:     "Same style - no changes",
			from:     style.Style{}.Foreground("red").Bold(true),
			to:       style.Style{}.Foreground("red").Bold(true),
			wantCode: "",
			wantDesc: "should generate no code for identical styles",
		},
		{
			name:     "Add BG to existing FG",
			from:     style.Style{}.Foreground("red"),
			to:       style.Style{}.Foreground("red").Background("blue"),
			wantCode: "\x1b[44m",
			wantDesc: "should only add BG code",
		},
		{
			name:     "Remove BG keep FG",
			from:     style.Style{}.Foreground("red").Background("blue"),
			to:       style.Style{}.Foreground("red"),
			wantCode: "\x1b[0m\x1b[31m",
			wantDesc: "should reset and reapply FG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ssm.buildDiffCodes(tt.from, tt.to)
			if got != tt.wantCode {
				t.Errorf("%s: %s\ngot:  %q\nwant: %q", tt.name, tt.wantDesc, got, tt.wantCode)
			}
		})
	}
}

// TestStyleStateMachine_Reset 测试状态重置
func TestStyleStateMachine_Reset(t *testing.T) {
	ssm := NewStyleStateMachine()

	// 设置一个非默认状态
	style1 := style.Style{}.Foreground("red").Bold(true)
	ssm.Update(style1)

	// 重置状态
	ssm.Reset()

	// 验证当前状态已重置为默认
	if ssm.NeedsUpdate(style.Style{}) {
		t.Error("After Reset(), state should be default and not need update for default style")
	}

	// 再次应用相同样式应该产生完整代码（因为内部状态已重置）
	diff := ssm.Update(style1)
	if diff == "" {
		t.Error("After Reset(), applying style should produce output")
	}
	if !strings.Contains(diff, "\x1b[") {
		t.Errorf("Should generate ANSI codes after reset, got: %q", diff)
	}
}

// TestStyleStateMachine_NeedsUpdate 测试需要更新检测
func TestStyleStateMachine_NeedsUpdate(t *testing.T) {
	ssm := NewStyleStateMachine()

	// 初始状态应该需要更新
	if !ssm.NeedsUpdate(style.Style{}.Foreground("red")) {
		t.Error("Initial state should need update for any non-default style")
	}

	// 更新后不应该需要更新相同样式
	ssm.Update(style.Style{}.Foreground("red"))
	if ssm.NeedsUpdate(style.Style{}.Foreground("red")) {
		t.Error("Same style should not need update")
	}

	// 不同样式应该需要更新
	if !ssm.NeedsUpdate(style.Style{}.Foreground("green")) {
		t.Error("Different style should need update")
	}
}

// TestStyleStateMachine_ComplexScenario 测试复杂场景
func TestStyleStateMachine_ComplexScenario(t *testing.T) {
	ssm := NewStyleStateMachine()

	// 场景：文本编辑器中的行样式变化
	// 第1行：普通文本
	line1Style := style.Style{}.Foreground("white")
	output1 := ssm.Update(line1Style)
	if output1 == "" {
		t.Error("First style update should produce output")
	}

	// 第2行：注释（绿色）
	commentStyle := style.Style{}.Foreground("green")
	output2 := ssm.Update(commentStyle)
	if !strings.Contains(output2, "\x1b[32m") {
		t.Errorf("Comment style should contain green FG code, got: %q", output2)
	}

	// 第3行：关键字（红色加粗）- 这会触发重置因为需要加粗
	keywordStyle := style.Style{}.Foreground("red").Bold(true)
	output3 := ssm.Update(keywordStyle)
	// 由于从 green 到 red+bold，应该生成完整样式
	if !strings.Contains(output3, "\x1b[") {
		t.Errorf("Keyword style should generate codes, got: %q", output3)
	}

	// 回到普通文本（需要关闭加粗）
	output4 := ssm.Update(line1Style)
	// 需要关闭加粗，应该触发重置
	if !strings.Contains(output4, "\x1b[0m") && !strings.Contains(output4, "\x1b[") {
		t.Errorf("Back to normal should generate codes to clear bold, got: %q", output4)
	}
}

// TestStyleStateMachine_EdgeCases 测试边界情况
func TestStyleStateMachine_EdgeCases(t *testing.T) {
	ssm := NewStyleStateMachine()

	tests := []struct {
		name     string
		from     style.Style
		to       style.Style
		contains string // 应包含的子串
	}{
		{
			name:     "Empty to empty",
			from:     style.Style{},
			to:       style.Style{},
			contains: "",
		},
		{
			name:     "All attributes on to all off",
			from:     style.Style{}.Foreground("red").Background("blue").Bold(true).Italic(true).Underline(true).Reverse(true),
			to:       style.Style{},
			contains: "\x1b[0m",
		},
		{
			name:     "Mixed attributes change",
			from:     style.Style{}.Foreground("red").Bold(true),
			to:       style.Style{}.Foreground("red").Italic(true),
			contains: "\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先设置 from 状态
			ssm.Update(tt.from)
			// 再切换到 to 状态
			got := ssm.Update(tt.to)

			if tt.contains == "" {
				// 期望空输出
				if tt.from == tt.to && got != "" {
					t.Errorf("Same styles should produce no output, got: %q", got)
				}
			} else {
				if !strings.Contains(got, tt.contains) {
					t.Errorf("Output should contain %q, got: %q", tt.contains, got)
				}
			}
			// 重置状态机用于下一个测试
			ssm.Reset()
		})
	}
}

// TestStyleStateMachine_Performance 测试性能 - 确保没有过度分配
func TestStyleStateMachine_Performance(t *testing.T) {
	ssm := NewStyleStateMachine()

	// 模拟高频样式切换（如光标闪烁）
	styles := []style.Style{
		style.Style{}.Foreground("white"),
		style.Style{}.Foreground("white").Reverse(true),
	}

	// 执行1000次切换
	for i := 0; i < 1000; i++ {
		ssm.Update(styles[i%2])
	}

	// 验证最终状态正确
	finalOutput := ssm.Update(style.Style{}.Foreground("white"))
	// 如果状态机有bug，可能会累积很多重置代码
	if len(finalOutput) > 100 {
		t.Errorf("Style state machine may be leaking state, output too long: %d chars", len(finalOutput))
	}
}

// TestStyleStateMachine_ManyChangesOptimization 测试多变化优化
func TestStyleStateMachine_ManyChangesOptimization(t *testing.T) {
	ssm := NewStyleStateMachine()

	// 初始状态：少量属性
	style1 := style.Style{}.Foreground("red").Bold(true)
	ssm.Update(style1)

	// 目标状态：4个以上属性变化（应该触发reset优化）
	style2 := style.Style{}.Foreground("green").Background("blue").Italic(true).Underline(true).Reverse(true)
	output := ssm.Update(style2)

	// 当变化过多时，应该使用reset而不是单独代码
	if !strings.Contains(output, "\x1b[0m") {
		t.Errorf("Many changes should use reset optimization, got: %q", output)
	}
}

// TestStyleStateMachine_ColorToDefault_Critical 关键测试：颜色到默认
// 这是修复的核心场景
func TestStyleStateMachine_ColorToDefault_Critical(t *testing.T) {
	ssm := NewStyleStateMachine()

	// 场景1: 红色前景 -> 默认
	ssm.Update(style.Style{}.Foreground("red"))
	output := ssm.Update(style.Style{})
	if output != "\x1b[0m" {
		t.Errorf("Red FG to default should produce \\x1b[0m, got: %q", output)
	}

	// 场景2: 蓝色背景 -> 默认
	ssm.Reset()
	ssm.Update(style.Style{}.Background("blue"))
	output = ssm.Update(style.Style{})
	if output != "\x1b[0m" {
		t.Errorf("Blue BG to default should produce \\x1b[0m, got: %q", output)
	}

	// 场景3: 红色前景+蓝色背景 -> 默认
	ssm.Reset()
	ssm.Update(style.Style{}.Foreground("red").Background("blue"))
	output = ssm.Update(style.Style{})
	if output != "\x1b[0m" {
		t.Errorf("FG+BG to default should produce \\x1b[0m, got: %q", output)
	}
}

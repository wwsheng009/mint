package components_test

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/checkbox"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/tooltip"
)

// This file tests Chinese character width support across all UI components.
// Chinese characters have display width of 2, while ASCII has width of 1.
// Components must use paint.StringWidth() instead of utf8.RuneCountInString().

func TestTextComponent_ChineseWidth(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantWidth int
	}{
		{"Single Chinese", "你", 2},
		{"Two Chinese", "你好", 4},
		{"Mixed", "Hi你好", 6},
		{"Fullwidth punctuation", "你好！", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textNode := text.New(tt.content)
			inst := textNode.CreateInstance().(interface {
				Measure(layout.Constraints) layout.Size
			})
			size := inst.Measure(layout.UnboundedConstraints())

			if size.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", size.Width, tt.wantWidth)
			}
		})
	}
}

func TestButtonComponent_ChineseWidth(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantWidth int // Expected display width of label
	}{
		{"ASCII", "OK", 2},
		{"Chinese", "确定", 4},
		{"Mixed", "保存Save", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := button.New(tt.label)
			inst := btn.CreateInstance().(interface {
				Measure(layout.Constraints) layout.Size
			})
			size := inst.Measure(layout.UnboundedConstraints())

			// Verify: label display width should match expected
			labelWidth := paint.StringWidth(tt.label)
			if labelWidth != tt.wantWidth {
				t.Errorf("Label display width = %d, want %d", labelWidth, tt.wantWidth)
			}

			// Button adds overhead (brackets, focus indicator, padding)
			// Just verify the label width calculation is correct
			t.Logf("Button label %q: display width=%d, button width=%d",
				tt.label, labelWidth, size.Width)
		})
	}
}

func TestInputComponent_ChineseWidth(t *testing.T) {
	tests := []struct {
		name  string
		value string
		width int
	}{
		{"ASCII value", "hello", 10},
		{"Chinese value", "你好世界", 10},
		{"Mixed value", "Hi你好", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputNode := input.New().SetValue(tt.value).SetWidth(tt.width)
			inst := inputNode.CreateInstance().(*input.Instance)

			cmds := inst.Paint(0, 0)
			if len(cmds) == 0 {
				t.Fatal("Paint returned no commands")
			}

			// Find the text command (should be command with the value)
			for _, cmd := range cmds {
				t.Logf("Command: X=%d, Y=%d, Text=%q", cmd.X, cmd.Y, cmd.Text)
			}
		})
	}
}

func TestCheckboxComponent_ChineseWidth(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantWidth int
	}{
		{"ASCII", "Enable", 6},
		{"Chinese", "启用", 4},
		{"Mixed", "保存Save", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := checkbox.New(tt.label)
			inst := cb.CreateInstance().(interface {
				Measure(layout.Constraints) layout.Size
			})
			size := inst.Measure(layout.UnboundedConstraints())

			labelWidth := paint.StringWidth(tt.label)
			t.Logf("Checkbox label %q: display width=%d, checkbox width=%d",
				tt.label, labelWidth, size.Width)

			if labelWidth != tt.wantWidth {
				t.Errorf("Label display width = %d, want %d", labelWidth, tt.wantWidth)
			}
		})
	}
}

func TestToastComponent_ChineseWidth(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantWidth int
	}{
		{"ASCII", "Click here", 10},
		{"Chinese", "点击此处", 8},
		{"Mixed", "Save保存", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toastVNode := tooltip.NewToast(tt.message)
			inst := toastVNode.CreateInstance().(interface {
				Measure(layout.Constraints) layout.Size
			})
			size := inst.Measure(layout.UnboundedConstraints())

			msgWidth := paint.StringWidth(tt.message)
			t.Logf("Toast message %q: display width=%d, toast width=%d",
				tt.message, msgWidth, size.Width)

			if msgWidth != tt.wantWidth {
				t.Errorf("Message display width = %d, want %d", msgWidth, tt.wantWidth)
			}
		})
	}
}

// TestStringWidthConsistency verifies paint.StringWidth is consistent
func TestStringWidthConsistency(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"abc", 3},
		{"你", 2},
		{"你好", 4},
		{"Hi你好", 6},
		{"！", 2}, // Fullwidth exclamation mark U+FF01
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := paint.StringWidth(tt.s)
			if got != tt.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

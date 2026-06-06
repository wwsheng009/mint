package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestSemanticTextShortcuts(t *testing.T) {
	tests := []struct {
		name    string
		node    VNode
		content string
	}{
		{"muted", TextMuted("helper"), "helper"},
		{"subtle", TextSubtle("detail"), "detail"},
		{"success", TextSuccess("saved"), "saved"},
		{"warning", TextWarning("pending"), "pending"},
		{"danger", TextDanger("failed"), "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.node.Tag() != "text" {
				t.Fatalf("tag = %q, want text", tt.node.Tag())
			}
			if got := tt.node.Props()["content"]; got != tt.content {
				t.Fatalf("content = %v, want %s", got, tt.content)
			}
		})
	}
}

func TestTextMutedLinesShortcut(t *testing.T) {
	node := TextMutedLines("first", " ", "second")
	if node == nil {
		t.Fatal("TextMutedLines() returned nil")
	}
	if node.Tag() != "vstack" {
		t.Fatalf("tag = %q, want vstack", node.Tag())
	}
	children := node.Children()
	if len(children) != 2 {
		t.Fatalf("children = %d, want 2", len(children))
	}
	if got := children[0].Props()["content"]; got != "first" {
		t.Fatalf("first line = %v, want first", got)
	}
	if got := children[1].Props()["content"]; got != "second" {
		t.Fatalf("second line = %v, want second", got)
	}
	if empty := TextMutedLines("", " \t "); empty != nil {
		t.Fatalf("empty TextMutedLines() = %v, want nil", empty)
	}
}

func TestTextStateMapsOperationalStates(t *testing.T) {
	tests := []struct {
		state string
		color style.Color
	}{
		{"healthy", "green"},
		{"enabled", "green"},
		{"pending_restart", "yellow"},
		{"risk", "yellow"},
		{"failed", "red"},
		{"firing", "red"},
		{"unknown", "gray"},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			node := TextState("status", tt.state)
			if node.Tag() != "text" {
				t.Fatalf("state %s tag = %q, want text", tt.state, node.Tag())
			}
			if got := node.Props()["content"]; got != "status" {
				t.Fatalf("state %s content = %v, want status", tt.state, got)
			}
			lineStyle, ok := node.Props()["style"].(style.Style)
			if !ok {
				t.Fatalf("state %s style = %T, want style.Style", tt.state, node.Props()["style"])
			}
			if lineStyle.FG != tt.color {
				t.Fatalf("state %s fg = %q, want %q", tt.state, lineStyle.FG, tt.color)
			}
		})
	}
}

func TestMaskSensitiveTextShortcut(t *testing.T) {
	if got := MaskSensitiveText("account-billing-prod", 4, 4); got != "acco...prod" {
		t.Fatalf("masked = %q, want acco...prod", got)
	}
}

func TestDisplayTextShortcuts(t *testing.T) {
	if got := DisplayText(" \n active\t "); got != "active" {
		t.Fatalf("display text = %q, want active", got)
	}
	if got := DisplayFallbackText("", "unknown"); got != "unknown" {
		t.Fatalf("fallback text = %q, want unknown", got)
	}
	if got := FirstNonEmptyText("", " \n\t ", " active\n"); got != "active" {
		t.Fatalf("first non-empty text = %q, want active", got)
	}
	if got := FirstNonEmptyText("", " \n\t "); got != "" {
		t.Fatalf("empty first non-empty text = %q, want blank", got)
	}
	if got := IntDisplayText(-2); got != "0" {
		t.Fatalf("int display text = %q, want 0", got)
	}
	if got := NonZeroIntText(12); got != "12" {
		t.Fatalf("non-zero int text = %q, want 12", got)
	}
	if got := NonZeroIntText(0); got != "" {
		t.Fatalf("zero non-zero int text = %q, want blank", got)
	}
	if got := RatioText(3, 5); got != "3/5" {
		t.Fatalf("ratio text = %q, want 3/5", got)
	}
	if got := SlashText(" provider-a ", "", " key-1\n"); got != "provider-a/-/key-1" {
		t.Fatalf("slash text = %q, want provider-a/-/key-1", got)
	}
	if got := ListText(" default ", "", " backup\n"); got != "default, backup" {
		t.Fatalf("list text = %q, want default, backup", got)
	}
	if got := BoolText(true); got != "yes" {
		t.Fatalf("bool true text = %q, want yes", got)
	}
	if got := BoolText(false); got != "no" {
		t.Fatalf("bool false text = %q, want no", got)
	}
	if got := EnabledText(true); got != "enabled" {
		t.Fatalf("enabled true text = %q, want enabled", got)
	}
	if got := EnabledText(false); got != "disabled" {
		t.Fatalf("enabled false text = %q, want disabled", got)
	}
	if got := ShortTimeText(time.Date(2026, 5, 28, 9, 8, 7, 0, time.UTC)); got != "09:08:07" {
		t.Fatalf("short time text = %q, want 09:08:07", got)
	}
	if got := ShortTimeText(time.Time{}); got != "-" {
		t.Fatalf("zero short time text = %q, want -", got)
	}
	if got := SortScopeText("server", "runtime", true); got != "server runtime desc" {
		t.Fatalf("sort scope text = %q, want server runtime desc", got)
	}
	if got := SortScopeText("", "provider", false); got != "provider asc" {
		t.Fatalf("unscoped sort text = %q, want provider asc", got)
	}
	if got := ServerSortScopeText("runtime", true); got != "server runtime desc" {
		t.Fatalf("server sort scope = %q, want server runtime desc", got)
	}
	if got := CurrentPageSortScopeText("latency", false); got != "current page latency asc" {
		t.Fatalf("current page sort scope = %q, want current page latency asc", got)
	}
	if got := SortLabelForColumn(1, "provider", "avg wait"); got != "avg wait" {
		t.Fatalf("sort label for column = %q, want avg wait", got)
	}
	if got := ColumnSortScopeText("rows", 1, true, "provider", "avg wait"); got != "rows avg wait desc" {
		t.Fatalf("column sort scope = %q, want rows avg wait desc", got)
	}
	if got := ColumnSortScopeText("rows", -1, true, "provider"); got != "" {
		t.Fatalf("unmapped column sort scope = %q, want empty", got)
	}
	if got := ServerColumnSortScopeText(0, true, "runtime"); got != "server runtime desc" {
		t.Fatalf("server column sort scope = %q, want server runtime desc", got)
	}
	if got := CurrentPageColumnSortScopeText(0, false, "latency"); got != "current page latency asc" {
		t.Fatalf("current page column sort scope = %q, want current page latency asc", got)
	}
	if got := PercentText(0.5); got != "50.0%" {
		t.Fatalf("percent text = %q, want 50.0%%", got)
	}
	if got := MillisecondsText(125.4); got != "125ms" {
		t.Fatalf("milliseconds text = %q, want 125ms", got)
	}
	if got := DecimalText(3.14); got != "3.1" {
		t.Fatalf("decimal text = %q, want 3.1", got)
	}
	if got := TrimmedDecimalText(3.10, 2); got != "3.1" {
		t.Fatalf("trimmed decimal text = %q, want 3.1", got)
	}
	if got := OptionalTrimmedDecimalText(3.145, true, 2); got != "3.15" {
		t.Fatalf("optional trimmed decimal text = %q, want 3.15", got)
	}
	if got := OptionalTrimmedDecimalText(3.145, false, 2); got != "-" {
		t.Fatalf("absent optional trimmed decimal text = %q, want -", got)
	}
	if got := UnitText(3.14, "req/s"); got != "3.1req/s" {
		t.Fatalf("unit text = %q, want 3.1req/s", got)
	}
	if got := OptionalUnitText(3.14, true, "req/s"); got != "3.1req/s" {
		t.Fatalf("optional unit text = %q, want 3.1req/s", got)
	}
	if got := OptionalUnitText(3.14, false, "req/s"); got != "-" {
		t.Fatalf("absent optional unit text = %q, want -", got)
	}
	if got := SecondsText(-1); got != "0s" {
		t.Fatalf("seconds text = %q, want 0s", got)
	}
	if got := DurationSecondsText(0); got != "-" {
		t.Fatalf("duration seconds text = %q, want -", got)
	}
	if got := CountSummaryText(
		CountTextPart{Label: "active", Count: 1},
		CountTextPart{Label: "idle", Count: -2},
	); got != "1 active / 0 idle" {
		t.Fatalf("count summary text = %q, want 1 active / 0 idle", got)
	}
	if got := KeyValueSummaryText(
		KeyValueTextPart{Label: "retries", Value: "3"},
		KeyValueTextPart{Label: "delay", Value: ""},
	); got != "retries=3 / delay=-" {
		t.Fatalf("key value summary text = %q, want retries=3 / delay=-", got)
	}
}

func TestCompactTextShortcutUsesDisplayWidth(t *testing.T) {
	if got := CompactText("abcdefghijklmnopqrstuvwxyz", 10); got != "abcdefg..." {
		t.Fatalf("compact text = %q, want abcdefg...", got)
	}
	wide := strings.Repeat("界", 20)
	got := CompactFallbackText(wide, "-", 10)
	if paint.StringWidth(got) > 10 {
		t.Fatalf("compact wide width = %d, want <= 10 (%q)", paint.StringWidth(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("compact wide = %q, want ellipsis suffix", got)
	}
	if got := OptionalCompactText(" \n\t ", 10); got != "" {
		t.Fatalf("optional compact blank = %q, want blank", got)
	}
	if got := OptionalCompactText(" active\nprovider ", 20); got != "active provider" {
		t.Fatalf("optional compact normalized = %q, want active provider", got)
	}
	if got := OptionalCompactText("abcdefghijklmnopqrstuvwxyz", 10); got != "abcdefg..." {
		t.Fatalf("optional compact text = %q, want abcdefg...", got)
	}
	got = OptionalCompactText(wide, 10)
	if paint.StringWidth(got) > 10 {
		t.Fatalf("optional compact wide width = %d, want <= 10 (%q)", paint.StringWidth(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("optional compact wide = %q, want ellipsis suffix", got)
	}
}

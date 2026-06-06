package text

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

func TestSemanticTextStateMapsOperationalStates(t *testing.T) {
	tests := []struct {
		state string
		color style.Color
	}{
		{"healthy", "green"},
		{"in sync", "green"},
		{"running", "yellow"},
		{"pending_restart", "yellow"},
		{"firing", "red"},
		{"out of sync", "red"},
		{"unknown", "gray"},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			node := State("status", tt.state)
			if node.Tag() != "text" {
				t.Fatalf("tag = %q, want text", node.Tag())
			}
			if got := node.Props()["content"]; got != "status" {
				t.Fatalf("content = %v, want status", got)
			}
			lineStyle, ok := node.Props()["style"].(style.Style)
			if !ok {
				t.Fatalf("style = %T, want style.Style", node.Props()["style"])
			}
			if lineStyle.FG != tt.color {
				t.Fatalf("fg = %q, want %q", lineStyle.FG, tt.color)
			}
		})
	}
}

func TestMaskSensitiveText(t *testing.T) {
	if got := MaskSensitive("account-billing-prod", 4, 4); got != "acco...prod" {
		t.Fatalf("masked account = %q, want acco...prod", got)
	}
	if got := MaskSensitive("provider-key-demo", 6, 4); got != "provid...demo" {
		t.Fatalf("masked key = %q, want provid...demo", got)
	}
	if got := MaskSensitive("short", 3, 3); got != "****" {
		t.Fatalf("short masked = %q, want ****", got)
	}
	if got := MaskSensitive(" \n\t ", 2, 4); got != "-" {
		t.Fatalf("blank masked = %q, want -", got)
	}
	if got := MaskSensitive("agw_example_token", 0, 0); got != "****" {
		t.Fatalf("fully masked = %q, want ****", got)
	}
}

func TestDisplayTextHelpers(t *testing.T) {
	if got := DisplayText(" \n healthy\t "); got != "healthy" {
		t.Fatalf("display text = %q, want healthy", got)
	}
	if got := DisplayText(" \n\t "); got != "-" {
		t.Fatalf("blank display text = %q, want -", got)
	}
	if got := FallbackText("", "unknown"); got != "unknown" {
		t.Fatalf("fallback text = %q, want unknown", got)
	}
	if got := FirstNonEmptyText("", " \n\t ", " active\n"); got != "active" {
		t.Fatalf("first non-empty text = %q, want active", got)
	}
	if got := FirstNonEmptyText("", " \n\t "); got != "" {
		t.Fatalf("empty first non-empty text = %q, want blank", got)
	}
	if got := IntText(-3); got != "0" {
		t.Fatalf("negative int text = %q, want 0", got)
	}
	if got := NonZeroIntText(12); got != "12" {
		t.Fatalf("non-zero int text = %q, want 12", got)
	}
	if got := NonZeroIntText(0); got != "" {
		t.Fatalf("zero non-zero int text = %q, want blank", got)
	}
	if got := NonZeroIntText(-3); got != "" {
		t.Fatalf("negative non-zero int text = %q, want blank", got)
	}
	if got := RatioText(4, 7); got != "4/7" {
		t.Fatalf("ratio text = %q, want 4/7", got)
	}
	if got := RatioText(-1, -2); got != "0/0" {
		t.Fatalf("negative ratio text = %q, want 0/0", got)
	}
	if got := SlashText(" provider-a ", "", " key-1\n"); got != "provider-a/-/key-1" {
		t.Fatalf("slash text = %q, want provider-a/-/key-1", got)
	}
	if got := SlashText(); got != "-" {
		t.Fatalf("empty slash text = %q, want -", got)
	}
	if got := ListText(" default ", "", " backup\n"); got != "default, backup" {
		t.Fatalf("list text = %q, want default, backup", got)
	}
	if got := ListText("", " \n\t "); got != "-" {
		t.Fatalf("empty list text = %q, want -", got)
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
	if got := SortScopeText(" server ", " runtime\n", true); got != "server runtime desc" {
		t.Fatalf("sort scope text = %q, want server runtime desc", got)
	}
	if got := SortScopeText("", "provider", false); got != "provider asc" {
		t.Fatalf("unscoped sort text = %q, want provider asc", got)
	}
	if got := SortScopeText("current page", " \n\t ", true); got != "" {
		t.Fatalf("blank sort label = %q, want empty", got)
	}
	if got := ServerSortScopeText("runtime", true); got != "server runtime desc" {
		t.Fatalf("server sort scope = %q, want server runtime desc", got)
	}
	if got := CurrentPageSortScopeText("latency", false); got != "current page latency asc" {
		t.Fatalf("current page sort scope = %q, want current page latency asc", got)
	}
	if got := SortLabelForColumn(1, "provider", " avg wait\n"); got != "avg wait" {
		t.Fatalf("sort label for column = %q, want avg wait", got)
	}
	if got := SortLabelForColumn(-1, "provider"); got != "" {
		t.Fatalf("negative sort label = %q, want empty", got)
	}
	if got := SortLabelForColumn(2, "provider"); got != "" {
		t.Fatalf("out-of-range sort label = %q, want empty", got)
	}
	if got := ColumnSortScopeText("rows", 2, true, "provider", "load", "avg wait"); got != "rows avg wait desc" {
		t.Fatalf("column sort scope = %q, want rows avg wait desc", got)
	}
	if got := ColumnSortScopeText("rows", 1, false, "provider", " "); got != "" {
		t.Fatalf("blank column sort scope = %q, want empty", got)
	}
	if got := ServerColumnSortScopeText(0, true, "runtime"); got != "server runtime desc" {
		t.Fatalf("server column sort scope = %q, want server runtime desc", got)
	}
	if got := CurrentPageColumnSortScopeText(0, false, "latency"); got != "current page latency asc" {
		t.Fatalf("current page column sort scope = %q, want current page latency asc", got)
	}
	if got := PercentText(0.875); got != "87.5%" {
		t.Fatalf("ratio percent text = %q, want 87.5%%", got)
	}
	if got := PercentText(12.345); got != "12.3%" {
		t.Fatalf("absolute percent text = %q, want 12.3%%", got)
	}
	if got := MillisecondsText(12.6); got != "13ms" {
		t.Fatalf("milliseconds text = %q, want 13ms", got)
	}
	if got := MillisecondsText(0); got != "-" {
		t.Fatalf("zero milliseconds text = %q, want -", got)
	}
	if got := DecimalText(12.34); got != "12.3" {
		t.Fatalf("decimal text = %q, want 12.3", got)
	}
	if got := TrimmedDecimalText(12.30, 2); got != "12.3" {
		t.Fatalf("trimmed decimal text = %q, want 12.3", got)
	}
	if got := TrimmedDecimalText(12, 2); got != "12" {
		t.Fatalf("integer trimmed decimal text = %q, want 12", got)
	}
	if got := TrimmedDecimalText(12.345, 2); got != "12.35" {
		t.Fatalf("rounded trimmed decimal text = %q, want 12.35", got)
	}
	if got := TrimmedDecimalText(-0.004, 2); got != "0" {
		t.Fatalf("negative zero trimmed decimal text = %q, want 0", got)
	}
	if got := OptionalTrimmedDecimalText(12.345, true, 2); got != "12.35" {
		t.Fatalf("optional trimmed decimal text = %q, want 12.35", got)
	}
	if got := OptionalTrimmedDecimalText(12.345, false, 2); got != "-" {
		t.Fatalf("absent optional trimmed decimal text = %q, want -", got)
	}
	if got := UnitText(12.34, "widgets"); got != "12.3widgets" {
		t.Fatalf("unit text = %q, want 12.3widgets", got)
	}
	if got := UnitText(12.34, "%"); got != "12.3%" {
		t.Fatalf("percent unit text = %q, want 12.3%%", got)
	}
	if got := UnitText(12.6, "ms"); got != "13ms" {
		t.Fatalf("milliseconds unit text = %q, want 13ms", got)
	}
	if got := OptionalUnitText(12.6, true, "ms"); got != "13ms" {
		t.Fatalf("optional unit text = %q, want 13ms", got)
	}
	if got := OptionalUnitText(12.6, false, "ms"); got != "-" {
		t.Fatalf("absent optional unit text = %q, want -", got)
	}
	if got := SecondsText(-2); got != "0s" {
		t.Fatalf("seconds text = %q, want 0s", got)
	}
	if got := DurationSecondsText(0); got != "-" {
		t.Fatalf("empty duration seconds text = %q, want -", got)
	}
	if got := DurationSecondsText(12); got != "12s" {
		t.Fatalf("duration seconds text = %q, want 12s", got)
	}
	if got := CountSummaryText(
		CountPart{Label: "total", Count: -3},
		CountPart{Label: "failed", Count: 2},
		CountPart{Label: "  ", Count: 9},
	); got != "0 total / 2 failed" {
		t.Fatalf("count summary text = %q, want 0 total / 2 failed", got)
	}
	if got := CountSummaryText(); got != "-" {
		t.Fatalf("empty count summary text = %q, want -", got)
	}
	if got := KeyValueSummaryText(
		KeyValuePart{Label: " retries ", Value: "3"},
		KeyValuePart{Label: "delay", Value: ""},
		KeyValuePart{Label: " ", Value: "ignored"},
	); got != "retries=3 / delay=-" {
		t.Fatalf("key value summary text = %q, want retries=3 / delay=-", got)
	}
	if got := KeyValueSummaryText(); got != "-" {
		t.Fatalf("empty key value summary text = %q, want -", got)
	}
}

func TestCompactTextUsesDisplayWidth(t *testing.T) {
	if got := CompactText("abcdefghijklmnopqrstuvwxyz", 10); got != "abcdefg..." {
		t.Fatalf("compact ascii = %q, want abcdefg...", got)
	}
	wide := strings.Repeat("界", 20)
	got := CompactText(wide, 10)
	if paint.StringWidth(got) > 10 {
		t.Fatalf("compact wide width = %d, want <= 10 (%q)", paint.StringWidth(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("compact wide = %q, want ellipsis suffix", got)
	}
	if got := CompactFallbackText("", "fallback-value", 8); got != "fallb..." {
		t.Fatalf("compact fallback = %q, want fallb...", got)
	}
	if got := OptionalCompactText(" \n\t ", 10); got != "" {
		t.Fatalf("optional compact blank = %q, want blank", got)
	}
	if got := OptionalCompactText(" active\nprovider ", 20); got != "active provider" {
		t.Fatalf("optional compact normalized = %q, want active provider", got)
	}
	if got := OptionalCompactText("abcdefghijklmnopqrstuvwxyz", 10); got != "abcdefg..." {
		t.Fatalf("optional compact ascii = %q, want abcdefg...", got)
	}
	got = OptionalCompactText(wide, 10)
	if paint.StringWidth(got) > 10 {
		t.Fatalf("optional compact wide width = %d, want <= 10 (%q)", paint.StringWidth(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("optional compact wide = %q, want ellipsis suffix", got)
	}
}

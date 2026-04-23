package palette

import (
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
)

func TestSeriesColorOrder(t *testing.T) {
	tests := []struct {
		index int
		want  string
	}{
		{index: 0, want: string(theme.Primary())},
		{index: 1, want: string(theme.Accent())},
		{index: 2, want: string(theme.Secondary())},
		{index: 3, want: string(theme.Success())},
		{index: 4, want: string(theme.Warning())},
		{index: 5, want: string(theme.Error())},
		{index: 6, want: string(theme.Primary())},
	}

	for _, tt := range tests {
		if got := string(SeriesColor(tt.index)); got != tt.want {
			t.Fatalf("SeriesColor(%d) = %q, want %q", tt.index, got, tt.want)
		}
	}
}

func TestSemanticColors(t *testing.T) {
	if got := string(AxisColor()); got != string(theme.Border()) {
		t.Fatalf("AxisColor() = %q, want %q", got, theme.Border())
	}
	if got := string(LabelColor()); got != string(theme.Muted()) {
		t.Fatalf("LabelColor() = %q, want %q", got, theme.Muted())
	}
	if got := string(ReferenceLineColor()); got != string(theme.Warning()) {
		t.Fatalf("ReferenceLineColor() = %q, want %q", got, theme.Warning())
	}
	if got := string(ReferenceBandColor()); got != string(theme.Secondary()) {
		t.Fatalf("ReferenceBandColor() = %q, want %q", got, theme.Secondary())
	}
	if got := string(CollisionColor()); got != string(theme.Warning()) {
		t.Fatalf("CollisionColor() = %q, want %q", got, theme.Warning())
	}
	if got := string(TitleColor()); got != string(theme.Text()) {
		t.Fatalf("TitleColor() = %q, want %q", got, theme.Text())
	}
	if got := string(UpColor()); got != string(theme.Success()) {
		t.Fatalf("UpColor() = %q, want %q", got, theme.Success())
	}
	if got := string(DownColor()); got != string(theme.Error()) {
		t.Fatalf("DownColor() = %q, want %q", got, theme.Error())
	}
	if got := string(FlatColor()); got != string(theme.Secondary()) {
		t.Fatalf("FlatColor() = %q, want %q", got, theme.Secondary())
	}
}

package statusbar

import "testing"

func TestKeyValuePreset(t *testing.T) {
	section := KeyValue("endpoint", "http://localhost:8080")
	if section.Text != "endpoint: http://localhost:8080" {
		t.Fatalf("text = %q", section.Text)
	}

	empty := KeyValue("sync", "")
	if empty.Text != "sync: -" {
		t.Fatalf("empty text = %q", empty.Text)
	}

	noLabel := KeyValue("", "ready")
	if noLabel.Text != "ready" {
		t.Fatalf("no-label text = %q", noLabel.Text)
	}
}

func TestMutedKeyValuePreset(t *testing.T) {
	section := MutedKeyValue("selection", "-")
	if section.Text != "selection: -" {
		t.Fatalf("text = %q", section.Text)
	}
	if section.FgColor != "bright-black" {
		t.Fatalf("fg = %q", section.FgColor)
	}
}

func TestStateBadgePresetUsesOperationalColors(t *testing.T) {
	tests := []struct {
		status string
		fg     string
		bg     string
	}{
		{"healthy", "black", "green"},
		{"pending_restart", "black", "yellow"},
		{"failed", "white", "red"},
		{"syncing", "black", "cyan"},
		{"unknown", "bright-white", "bright-black"},
	}

	for _, tt := range tests {
		section := StateBadge(tt.status)
		if section.FgColor != tt.fg || section.BgColor != tt.bg {
			t.Fatalf("StateBadge(%q) colors = %q/%q, want %q/%q", tt.status, section.FgColor, section.BgColor, tt.fg, tt.bg)
		}
		if section.Text != " "+tt.status+" " {
			t.Fatalf("StateBadge(%q) text = %q", tt.status, section.Text)
		}
	}
}

func TestBusyAndErrorBadgePresets(t *testing.T) {
	busy := BusyBadge("")
	if busy.Text != " busy " || busy.FgColor != "black" || busy.BgColor != "yellow" {
		t.Fatalf("busy = %+v", busy)
	}

	err := ErrorBadge("failed")
	if err.Text != " failed " || err.FgColor != "white" || err.BgColor != "red" {
		t.Fatalf("error = %+v", err)
	}
}

func TestDefaultTone(t *testing.T) {
	tests := map[string]Tone{
		"effective":       ToneNormal,
		"rate limited":    ToneWarn,
		"out_of_sync":     ToneError,
		"refreshing":      ToneInfo,
		"custom_internal": ToneNeutral,
	}
	for status, want := range tests {
		if got := DefaultTone(status); got != want {
			t.Fatalf("DefaultTone(%q) = %q, want %q", status, got, want)
		}
	}
}

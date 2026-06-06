package statusbar

import (
	"testing"
	"time"
)

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

func TestOperationalStatusPresets(t *testing.T) {
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		got  Section
		want string
	}{
		{"endpoint", Endpoint("http://localhost:8080"), "endpoint: http://localhost:8080"},
		{"profile", Profile("local"), "profile: local"},
		{"user", User("admin"), "user: admin"},
		{"role", Role("ops"), "role: ops"},
		{"page", Page("jobs"), "page: jobs"},
		{"scope", Scope("provider"), "scope: provider"},
		{"target", Target("openai/key-1"), "target: openai/key-1"},
		{"selection", Selection("job-1"), "selection: job-1"},
		{"filter", Filter("failed"), "filter: failed"},
		{"count", Count("keys", 12), "keys: 12"},
		{"negative count", Count("keys", -1), "keys: 0"},
		{"latency ms", Latency(250 * time.Millisecond), "latency: 250ms"},
		{"uptime hours", Uptime(3 * time.Hour), "uptime: 3h"},
		{"hotkey", Hotkey("r", "reload"), "r reload"},
		{"hotkey empty", Hotkey("", ""), "-"},
		{"separator", Separator(), " | "},
		{"last sync never", LastSync(time.Time{}, now), "last sync: never"},
		{"last sync seconds", LastSync(now.Add(-30*time.Second), now), "last sync: 30s ago"},
		{"last sync minutes", LastSync(now.Add(-5*time.Minute), now), "last sync: 5m ago"},
		{"auto refresh", AutoRefresh(15 * time.Second), "refresh: 15s"},
		{"auto refresh now", AutoRefresh(0), "refresh: now"},
	}

	for _, tt := range tests {
		if tt.got.Text != tt.want {
			t.Fatalf("%s text = %q, want %q", tt.name, tt.got.Text, tt.want)
		}
	}
}

func TestSelectionTargetPreset(t *testing.T) {
	if got := SelectionTarget("job", "Sync", 12); got != "job Sync" {
		t.Fatalf("selection target = %q, want job Sync", got)
	}
	if got := SelectionTarget("", "openai/default/key-openai-1", 18); got != "openai/default/..." {
		t.Fatalf("selection target without kind = %q", got)
	}
	if got := SelectionTarget("key", "", 18); got != "-" {
		t.Fatalf("empty selection target = %q, want -", got)
	}
	if got := SelectionTarget("trace", "abcdefghijklmnopqrstuvwxyz", 10); got != "trace abcdefg..." {
		t.Fatalf("compact selection target = %q", got)
	}
}

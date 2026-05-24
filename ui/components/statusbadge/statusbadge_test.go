package statusbadge

import (
	"testing"

	"github.com/wwsheng009/mint/ui/components/badge"
)

func TestDefaultMapping(t *testing.T) {
	cases := []struct {
		status string
		want   badge.Status
	}{
		{status: "healthy", want: badge.StatusSuccess},
		{status: "rate_limited", want: badge.StatusWarning},
		{status: "unauthorized", want: badge.StatusError},
		{status: "syncing", want: badge.StatusProcessing},
		{status: "unknown", want: badge.StatusDefault},
	}

	for _, tt := range cases {
		node := New(tt.status)
		if got := node.Props()["status"]; got != tt.want {
			t.Fatalf("New(%q) status = %v, want %v", tt.status, got, tt.want)
		}
		if got := node.Props()["text"]; got != tt.status {
			t.Fatalf("New(%q) text = %v, want %v", tt.status, got, tt.status)
		}
	}
}

func TestOptions(t *testing.T) {
	node := New(
		"custom",
		Key("status.custom"),
		Label("Provider"),
		Text("CUSTOM"),
		Dot(),
		ForceTone(ToneWarn),
	)

	props := node.Props()
	if got := props["key"]; got != "status.custom" {
		t.Fatalf("key = %v, want status.custom", got)
	}
	if got := props["label"]; got != "Provider" {
		t.Fatalf("label = %v, want Provider", got)
	}
	if got := props["text"]; got != "CUSTOM" {
		t.Fatalf("text = %v, want CUSTOM", got)
	}
	if got := props["dot"]; got != true {
		t.Fatalf("dot = %v, want true", got)
	}
	if got := props["status"]; got != badge.StatusWarning {
		t.Fatalf("status = %v, want warning", got)
	}
}

func TestCustomMapper(t *testing.T) {
	node := New("paused", Mapper(func(string) Tone {
		return ToneWarn
	}))
	if got := node.Props()["status"]; got != badge.StatusWarning {
		t.Fatalf("status = %v, want warning", got)
	}
}

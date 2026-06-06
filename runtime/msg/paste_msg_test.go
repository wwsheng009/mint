package msg

import (
	"strings"
	"testing"
)

func TestPasteMsgStringDoesNotExposeText(t *testing.T) {
	msg := NewPasteMsg("https://example.test/login?token=agw_example_token")

	got := msg.String()
	if !strings.Contains(got, "runes=") || !strings.Contains(got, "bytes=") {
		t.Fatalf("paste msg string = %q, want rune/byte summary", got)
	}
	if strings.Contains(got, "https://example.test/login") || strings.Contains(got, "agw_example_token") {
		t.Fatalf("paste msg string leaked pasted text: %q", got)
	}
}

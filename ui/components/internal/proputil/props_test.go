package proputil

import (
	"reflect"
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type stubIntent struct {
	name string
}

func (i stubIntent) IntentType() string {
	return i.name
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name  string
		props rtui.Props
		want  string
	}{
		{name: "value", props: rtui.Props{"title": "Mint"}, want: "Mint"},
		{name: "missing", props: nil, want: "fallback"},
		{name: "wrong type", props: rtui.Props{"title": 42}, want: "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetString(tt.props, "title", "fallback"); got != tt.want {
				t.Fatalf("GetString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name  string
		props rtui.Props
		want  bool
	}{
		{name: "value", props: rtui.Props{"disabled": true}, want: true},
		{name: "missing", props: nil, want: false},
		{name: "wrong type", props: rtui.Props{"disabled": "true"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBool(tt.props, "disabled", false); got != tt.want {
				t.Fatalf("GetBool() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name  string
		props rtui.Props
		want  int
	}{
		{name: "value", props: rtui.Props{"width": 48}, want: 48},
		{name: "missing", props: nil, want: 12},
		{name: "wrong type", props: rtui.Props{"width": "48"}, want: 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInt(tt.props, "width", 12); got != tt.want {
				t.Fatalf("GetInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetStyle(t *testing.T) {
	valueStyle := style.NewStyle().Foreground(style.Color("green")).Bold(true)
	defaultStyle := style.NewStyle().Background(style.Color("black"))

	tests := []struct {
		name  string
		props rtui.Props
		want  style.Style
	}{
		{name: "value", props: rtui.Props{"labelStyle": valueStyle}, want: valueStyle},
		{name: "missing", props: nil, want: defaultStyle},
		{name: "wrong type", props: rtui.Props{"labelStyle": "green"}, want: defaultStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStyle(tt.props, "labelStyle", defaultStyle); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetStyle() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetIntent(t *testing.T) {
	valueIntent := stubIntent{name: "selection.change"}
	defaultIntent := stubIntent{name: "selection.default"}

	tests := []struct {
		name  string
		props rtui.Props
		want  string
	}{
		{name: "value", props: rtui.Props{"changeIntent": valueIntent}, want: "selection.change"},
		{name: "missing", props: nil, want: "selection.default"},
		{name: "wrong type", props: rtui.Props{"changeIntent": "selection.change"}, want: "selection.default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIntent(tt.props, "changeIntent", defaultIntent)
			if got == nil {
				t.Fatal("GetIntent() returned nil")
			}
			if got.IntentType() != tt.want {
				t.Fatalf("GetIntent().IntentType() = %q, want %q", got.IntentType(), tt.want)
			}
		})
	}
}

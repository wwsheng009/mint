package statistic

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "statistic" {
		t.Fatalf("Tag = %q, want statistic", v.Tag())
	}
	if v.precision != -1 {
		t.Fatalf("precision = %d, want -1", v.precision)
	}
	if v.groupSeparator != "," || v.decimalSeparator != "." {
		t.Fatalf("separators = (%q,%q), want (, .)", v.groupSeparator, v.decimalSeparator)
	}
}

func TestBuilderFluent(t *testing.T) {
	extra := textcomp.New("Compared to yesterday")
	v := NewBuilder().
		Key("revenue").
		Title("Revenue").
		Value(12345.67).
		Prefix("$").
		Suffix(" USD").
		Extra(extra).
		Precision(2).
		GroupSeparator(" ").
		DecimalSeparator(",").
		Loading(true).
		Bordered(true).
		Down().
		Width(32).
		TitleStyle(style.NewStyle().Bold(true)).
		ValueStyle(style.NewStyle().Foreground(style.Color("green"))).
		BuildVNode()

	if v.Key() != "revenue" || v.title != "Revenue" {
		t.Fatalf("key/title = (%q,%q)", v.Key(), v.title)
	}
	if v.prefix != "$" || v.suffix != " USD" || v.extra != extra {
		t.Fatalf("prefix/suffix/extra not set")
	}
	if v.precision != 2 || v.groupSeparator != " " || v.decimalSeparator != "," {
		t.Fatalf("format props = (%d,%q,%q)", v.precision, v.groupSeparator, v.decimalSeparator)
	}
	if !v.loading || !v.bordered || v.trend != TrendDown || v.width != 32 {
		t.Fatalf("state props = (loading=%v bordered=%v trend=%v width=%d)", v.loading, v.bordered, v.trend, v.width)
	}
}

func TestFormatValueIntegersAndFloats(t *testing.T) {
	if got := formatValue(1234567, -1, ",", "."); got != "1,234,567" {
		t.Fatalf("format int = %q, want 1,234,567", got)
	}
	if got := formatValue(12345.678, 2, " ", ","); got != "12 345,68" {
		t.Fatalf("format float = %q, want 12 345,68", got)
	}
	if got := formatValue("custom", 2, ",", "."); got != "custom" {
		t.Fatalf("format string = %q, want custom", got)
	}
}

func TestRuntimeChildrenRendersTitleValueAndExtra(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:     "Revenue",
		propValue:     12345.67,
		propPrefix:    "$",
		propSuffix:    " USD",
		propPrecision: 2,
		propTrend:     TrendUp,
		propExtra:     textcomp.New("Compared to yesterday"),
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "Revenue") {
		t.Fatal("expected title text")
	}
	if !containsVNodeText(root, "$") {
		t.Fatal("expected prefix text")
	}
	if !containsVNodeText(root, "12,345.67") {
		t.Fatal("expected formatted value text")
	}
	if !containsVNodeText(root, " USD") {
		t.Fatal("expected suffix text")
	}
	if !containsVNodeText(root, "↑") {
		t.Fatal("expected trend text")
	}
	if !containsVNodeText(root, "Compared to yesterday") {
		t.Fatal("expected extra text")
	}
}

func TestRuntimeChildrenLoadingUsesPlaceholder(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:   "Requests",
		propValue:   999,
		propLoading: true,
	})
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if !containsVNodeText(children[0], "...") {
		t.Fatal("expected loading placeholder")
	}
	if containsVNodeText(children[0], "999") {
		t.Fatal("did not expect formatted value while loading")
	}
}

func TestRuntimeChildrenBorderedRootHasBorderProp(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "Latency",
		propValue:    128,
		propBordered: true,
	})
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	props := children[0].Props()
	if props == nil || props.GetString("borderStyle") != "single" {
		t.Fatalf("borderStyle = %q, want single", props.GetString("borderStyle"))
	}
}

func containsVNodeText(node rtui.VNode, want string) bool {
	if node == nil {
		return false
	}
	if contentProvider, ok := node.(interface{ Content() string }); ok && contentProvider.Content() == want {
		return true
	}
	if props := node.Props(); props != nil && props.GetString("content") == want {
		return true
	}
	for _, child := range node.Children() {
		if containsVNodeText(child, want) {
			return true
		}
	}
	return false
}

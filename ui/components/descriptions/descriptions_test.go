package descriptions

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

func TestNew(t *testing.T) {
	v := New([]Item{Field("Name", "Mint")})
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "descriptions" {
		t.Fatalf("Tag = %q, want descriptions", v.Tag())
	}
	if v.column != 3 {
		t.Fatalf("column = %d, want 3", v.column)
	}
	if !v.colon {
		t.Fatal("colon should default to true")
	}
	if v.layout != LayoutHorizontal {
		t.Fatalf("layout = %v, want horizontal", v.layout)
	}
}

func TestBuilderFluent(t *testing.T) {
	extra := textcomp.New("Copy")
	v := NewBuilder().
		Key("build-meta").
		Title("Build Info").
		Extra(extra).
		Column(2).
		Bordered(true).
		Colon(false).
		Vertical().
		Width(64).
		LabelWidth(14).
		ContentWidth(28).
		EmptyText("-").
		MaskText("masked").
		LabelStyle(style.NewStyle().Bold(true)).
		ContentStyle(style.NewStyle().Foreground(style.Color("cyan"))).
		Item(Field("Version", "v1.2.3").WithKey("version")).
		BuildVNode()

	if v.Key() != "build-meta" {
		t.Fatalf("Key = %q, want build-meta", v.Key())
	}
	if v.title != "Build Info" || v.extra != extra {
		t.Fatalf("header props not set: title=%q extra=%v", v.title, v.extra)
	}
	if v.column != 2 || !v.bordered || v.colon {
		t.Fatalf("layout props = (column=%d bordered=%v colon=%v)", v.column, v.bordered, v.colon)
	}
	if v.layout != LayoutVertical {
		t.Fatalf("layout = %v, want vertical", v.layout)
	}
	if v.width != 64 {
		t.Fatalf("width = %d, want 64", v.width)
	}
	if v.labelWidth != 14 || v.contentWidth != 28 {
		t.Fatalf("width props = (label=%d content=%d), want (14,28)", v.labelWidth, v.contentWidth)
	}
	if v.emptyText != "-" || v.maskText != "masked" {
		t.Fatalf("text props = (empty=%q mask=%q)", v.emptyText, v.maskText)
	}
}

func TestNormalizeItemsAssignsKeysAndSpan(t *testing.T) {
	items := normalizeItems([]Item{
		Field("Name", "Mint").WithKey("dup").WithSpan(0),
		Field("Env", "Prod").WithKey("dup").WithSpan(2),
	})
	if items[0].Key != "dup" {
		t.Fatalf("items[0].Key = %q, want dup", items[0].Key)
	}
	if items[1].Key != "dup-1" {
		t.Fatalf("items[1].Key = %q, want dup-1", items[1].Key)
	}
	if items[0].Span != 1 {
		t.Fatalf("items[0].Span = %d, want 1", items[0].Span)
	}
}

func TestPackRowsRespectsSpan(t *testing.T) {
	rows := packRows([]Item{
		Field("Name", "Mint"),
		Field("Version", "v1.2.3").WithSpan(2),
		Field("Commit", "abc"),
	}, 3)

	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(rows))
	}
	if len(rows[0]) != 2 || rows[0][0].span != 1 || rows[0][1].span != 2 {
		t.Fatalf("row 0 = %#v", rows[0])
	}
	if len(rows[1]) != 1 || rows[1][0].item.Label != "Commit" {
		t.Fatalf("row 1 = %#v", rows[1])
	}
}

func TestValueItemsRenderEmptyAndSensitiveValues(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn:    1,
		propEmptyText: "-",
		propMaskText:  "hidden",
		propItems: []Item{
			Value("Missing", nil).WithKey("missing"),
			Value("Blank", "").WithKey("blank"),
			SensitiveField("Token", "agw_example_token").WithKey("token"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "-") {
		t.Fatal("expected empty placeholder")
	}
	if !containsVNodeText(root, "hidden") {
		t.Fatal("expected sensitive value to be masked")
	}
	if containsVNodeText(root, "agw_example_token") {
		t.Fatal("sensitive value should not be rendered")
	}
}

func TestRuntimeChildrenApplyStableLabelAndContentWidths(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn:       1,
		propLabelWidth:   12,
		propContentWidth: 24,
		propItems: []Item{
			Value("Runtime", "healthy").WithKey("runtime"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]

	label := findVNodeByKey(root, "descriptions-label-runtime")
	if label == nil {
		t.Fatal("expected label width wrapper")
	}
	if got := label.Props()["width"]; got != 12 {
		t.Fatalf("label width = %v, want 12", got)
	}

	content := findVNodeByKey(root, "descriptions-content-width-runtime")
	if content == nil {
		t.Fatal("expected content width wrapper")
	}
	if got := content.Props()["width"]; got != 24 {
		t.Fatalf("content width = %v, want 24", got)
	}
}

func TestInstanceRuntimeChildrenContainsHeaderAndValues(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:  "Build Info",
		propExtra:  textcomp.New("Copy"),
		propColumn: 2,
		propItems: []Item{
			Field("Name", "Mint").WithKey("name"),
			Field("Version", "v1.2.3").WithKey("version"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "Build Info") {
		t.Fatal("expected title text in runtime children")
	}
	if !containsVNodeText(root, "Copy") {
		t.Fatal("expected extra text in runtime children")
	}
	if !containsVNodeText(root, "Name:") {
		t.Fatal("expected label text with colon")
	}
	if !containsVNodeText(root, "Mint") || !containsVNodeText(root, "v1.2.3") {
		t.Fatal("expected values in runtime children")
	}
}

func TestInstanceVerticalLayoutOmitsColon(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn: 1,
		propColon:  false,
		propLayout: LayoutVertical,
		propItems: []Item{
			Field("Status", "Ready").WithKey("status"),
		},
	})

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if !containsVNodeText(root, "Status") {
		t.Fatal("expected label text without colon")
	}
	if containsVNodeText(root, "Status:") {
		t.Fatal("did not expect label text with colon")
	}
}

func TestInstanceEffectiveColumnsResponsiveToWidth(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propColumn: 3,
		propWidth:  60,
		propItems:  []Item{Field("A", "1")},
	})
	if got := inst.effectiveColumns(); got != 2 {
		t.Fatalf("effectiveColumns = %d, want 2", got)
	}

	inst.SetProps(rtui.Props{
		propColumn: 3,
		propWidth:  40,
		propItems:  []Item{Field("A", "1")},
	})
	if got := inst.effectiveColumns(); got != 1 {
		t.Fatalf("effectiveColumns = %d, want 1", got)
	}
}

func TestNormalizeItemsClampsNegativeWidths(t *testing.T) {
	items := normalizeItems([]Item{
		Value("Runtime", "healthy").WithLabelWidth(-1).WithContentWidth(-2),
	})
	if items[0].LabelWidth != 0 {
		t.Fatalf("label width = %d, want 0", items[0].LabelWidth)
	}
	if items[0].ContentWidth != 0 {
		t.Fatalf("content width = %d, want 0", items[0].ContentWidth)
	}
}

func containsVNodeText(node rtui.VNode, want string) bool {
	if node == nil {
		return false
	}
	if contentProvider, ok := node.(interface{ Content() string }); ok && contentProvider.Content() == want {
		return true
	}
	if props := node.Props(); props != nil {
		if props.GetString("content") == want {
			return true
		}
	}
	for _, child := range node.Children() {
		if containsVNodeText(child, want) {
			return true
		}
	}
	return false
}

func findVNodeByKey(node rtui.VNode, key string) rtui.VNode {
	if node == nil {
		return nil
	}
	if node.Key() == key {
		return node
	}
	for _, child := range node.Children() {
		if found := findVNodeByKey(child, key); found != nil {
			return found
		}
	}
	return nil
}

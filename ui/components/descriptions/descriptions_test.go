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

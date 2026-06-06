package ui

import "testing"

func TestNavDataShortcuts(t *testing.T) {
	listNode := List().
		Rows([]string{"A", "B"}).
		Build()
	if listNode == nil || listNode.Tag() != "list" {
		t.Fatalf("List().Build().Tag() = %q, want list", listNode.Tag())
	}

	treeNode := TreeView().
		Nodes([]TreeNode{{Content: "root", NodeID: 1}}).
		Build()
	if treeNode == nil || treeNode.Tag() != "treeview" {
		t.Fatalf("TreeView().Build().Tag() = %q, want treeview", treeNode.Tag())
	}

	drawerNode := Drawer(Text("Body"))
	if drawerNode == nil || drawerNode.Tag() != "drawer" {
		t.Fatalf("Drawer().Tag() = %q, want drawer", drawerNode.Tag())
	}

	titled := DrawerTitled("Settings", Text("Body"))
	if titled == nil || titled.Tag() != "drawer" {
		t.Fatalf("DrawerTitled().Tag() = %q, want drawer", titled.Tag())
	}

	modalNode := ModalOf(Text("Body"))
	if modalNode == nil || modalNode.Tag() != "modal" {
		t.Fatalf("ModalOf().Tag() = %q, want modal", modalNode.Tag())
	}

	openModal := ModalOpenedTitled("Picker", Text("Body"))
	if openModal == nil || openModal.Tag() != "modal" {
		t.Fatalf("ModalOpenedTitled().Tag() = %q, want modal", openModal.Tag())
	}
	if got := openModal.Props()["isOpen"]; got != true {
		t.Fatalf("ModalOpenedTitled().isOpen = %v, want true", got)
	}
	if got := openModal.Props()["title"]; got != "Picker" {
		t.Fatalf("ModalOpenedTitled().title = %v, want Picker", got)
	}
}

func TestListShortcutSortAscending(t *testing.T) {
	node := List().
		Rows([]string{"zeta", "alpha", "beta"}).
		SortAscending().
		Build()

	props := node.Props()
	rows, ok := props["rows"].([]string)
	if !ok {
		t.Fatalf("rows prop = %T, want []string", props["rows"])
	}
	want := []string{"alpha", "beta", "zeta"}
	for index := range want {
		if rows[index] != want[index] {
			t.Fatalf("rows = %v, want %v", rows, want)
		}
	}
	if props["sortRows"] != true || props["sortDescending"] != false {
		t.Fatalf("sort props = (%v,%v), want (true,false)", props["sortRows"], props["sortDescending"])
	}
}

func TestListItemShortcutsSupportStructuredSortedRows(t *testing.T) {
	node := List().
		Items([]ListItem{
			NewListItem("openai").WithPrefix("[ok]").WithDescription("healthy"),
			NewListItemWithDescription("anthropic", "degraded").WithPrefix("[warn]"),
		}).
		SortAscending().
		Build()

	props := node.Props()
	items, ok := props["items"].([]ListItem)
	if !ok {
		t.Fatalf("items prop = %T, want []ListItem", props["items"])
	}
	if len(items) != 2 || items[0].Title != "anthropic" || items[1].Title != "openai" {
		t.Fatalf("items = %+v, want sorted by structured title", items)
	}
	rows, ok := props["rows"].([]string)
	if !ok {
		t.Fatalf("rows prop = %T, want []string", props["rows"])
	}
	if len(rows) != 2 || rows[0] != "[warn] anthropic - degraded" || rows[1] != "[ok] openai - healthy" {
		t.Fatalf("rows = %#v, want flattened structured rows", rows)
	}
}

func TestNewDrawerBuilder(t *testing.T) {
	vnode := NewDrawerBuilder().
		Title("Settings").
		Content(Text("Body")).
		Opened().
		Build()
	if vnode == nil {
		t.Fatal("NewDrawerBuilder().Build() returned nil")
	}
	if vnode.Tag() != "drawer" {
		t.Fatalf("Tag() = %q, want drawer", vnode.Tag())
	}
}

func TestDescriptionsPanelWithContextShortcut(t *testing.T) {
	node := DescriptionsPanelWithContext(
		"jobs.selection",
		"Job Detail",
		62,
		14,
		40,
		DescriptionsContextStripConfig{
			LabelWidth:   7,
			ContentWidth: 12,
			Items: []DescriptionsItem{
				NewDescriptionsCompactValue("ID", "job-1", 12),
				NewDescriptionsStateValue("Status", "running", "running"),
			},
		},
		[]DescriptionsItem{NewDescriptionsValue("Current Run", "run-1")},
	)

	if node == nil {
		t.Fatal("DescriptionsPanelWithContext() returned nil")
	}
	if node.Tag() != "panel" {
		t.Fatalf("DescriptionsPanelWithContext().Tag() = %q, want panel", node.Tag())
	}
	content, ok := node.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", node.Props()["content"])
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want context and details", len(children))
	}
	if children[0].Key() != "jobs.selection.context" {
		t.Fatalf("context key = %q, want jobs.selection.context", children[0].Key())
	}
}

func TestDetailPanelShortcut(t *testing.T) {
	node := DetailPanel(DetailPanelConfig{
		Key:          "logs.selection",
		Title:        "Log Detail",
		Width:        56,
		LabelWidth:   12,
		ContentWidth: 36,
		Context: DescriptionsContextStripConfig{
			Column:       2,
			LabelWidth:   8,
			ContentWidth: 18,
			Items: []DescriptionsItem{
				NewDescriptionsCompactValue("Request", "request-1", 18),
				NewDescriptionsStateValue("Status", "failed", "failed"),
			},
		},
		Items: []DescriptionsItem{
			NewDescriptionsCompactValue("Path", "/v1/chat/completions", 36),
		},
	})

	if node == nil {
		t.Fatal("DetailPanel() returned nil")
	}
	if node.Tag() != "panel" {
		t.Fatalf("DetailPanel().Tag() = %q, want panel", node.Tag())
	}
	content, ok := node.Props()["content"].(VNode)
	if !ok {
		t.Fatalf("content = %T, want VNode", node.Props()["content"])
	}
	if content.Tag() != "vstack" {
		t.Fatalf("content tag = %q, want vstack", content.Tag())
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("children len = %d, want context and details", len(children))
	}
	if children[0].Key() != "logs.selection.context" {
		t.Fatalf("context key = %q, want logs.selection.context", children[0].Key())
	}
	if children[1].Key() != "logs.selection.details" {
		t.Fatalf("details key = %q, want logs.selection.details", children[1].Key())
	}
}

func TestDetailPanelEmptyHintShortcut(t *testing.T) {
	got := DetailPanelEmptyHint("Refresh logs or clear search.",
		KeyValueTextPart{Label: "source", Value: "page snapshot"},
		KeyValueTextPart{Label: "search", Value: "trace-1"},
	)
	want := "Refresh logs or clear search. Scope: source=page snapshot / search=trace-1"
	if got != want {
		t.Fatalf("DetailPanelEmptyHint() = %q, want %q", got, want)
	}

	got = DetailPanelEmptyHintWithScopeWidth("Clear search.", 18,
		KeyValueTextPart{Label: "search", Value: "abcdefghijklmnopqrstuvwxyz"},
	)
	if got != "Clear search. Scope: search=abcdefgh..." {
		t.Fatalf("DetailPanelEmptyHintWithScopeWidth() = %q, want compact scope", got)
	}
}

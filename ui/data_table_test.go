package ui

import "testing"

func TestDataTableOperationalStateShortcuts(t *testing.T) {
	node := DataTable(
		[]TableColumn{{Title: "Provider", Width: 16}},
		[][]string{{"openai"}},
		DataTableLoading(true),
		DataTableLoadingText("Loading providers..."),
		DataTableServerPagination(1, 25, 40),
	)

	props := node.Props()
	if got := props["emptyText"]; got != "Loading providers..." {
		t.Fatalf("emptyText = %v, want loading text", got)
	}
	if got := props["statusText"]; got != "Loading" {
		t.Fatalf("statusText = %v, want Loading", got)
	}
}

func TestDataTableServerPaginationShortcut(t *testing.T) {
	node := DataTable(
		[]TableColumn{{Title: "Provider", Width: 16}},
		[][]string{{"openai"}, {"azure"}},
		DataTableServerPagination(2, 25, 76),
	)

	if got := node.Props()["statusText"]; got != "Page 2/4 · Total 76 · Size 25" {
		t.Fatalf("statusText = %v, want server pagination summary", got)
	}
}

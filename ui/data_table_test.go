package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
)

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

func TestOperationalDataTableShortcut(t *testing.T) {
	node := OperationalDataTable(
		[]TableColumn{{Title: "Provider", Width: 16}},
		[][]string{{"openai"}, {"azure"}},
		1,
		25,
		"selectedProviderIndex",
		DataTableRowKeys([]string{"provider.openai", "provider.azure"}),
		DataTableCurrentPage(2),
		DataTableSelectedKey("provider.azure"),
		DataTableSelectedKeyField("selectedProviderKey"),
		DataTablePageField("providerPage"),
		DataTableActivateKeyField("activatedProviderKey"),
		DataTableSortBy(0, true),
		DataTableOnChange(dataTableTestIntent{name: "datatable.change"}),
		DataTableServerPagination(2, 25, 76),
	)

	props := node.Props()
	if got := props["pageSize"]; got != 25 {
		t.Fatalf("pageSize = %v, want 25", got)
	}
	if got := props["showFooter"]; got != true {
		t.Fatalf("showFooter = %v, want true", got)
	}
	if got := props["showScrollbar"]; got != true {
		t.Fatalf("showScrollbar = %v, want true", got)
	}
	if got := props["currentPage"]; got != 2 {
		t.Fatalf("currentPage = %v, want 2", got)
	}
	if _, ok := props["changeIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("changeIntentField = %T, want intent.FieldIntent", props["changeIntentField"])
	}
	if _, ok := props["pageIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("pageIntentField = %T, want intent.FieldIntent", props["pageIntentField"])
	}
	if got := props["sortColumn"]; got != 0 {
		t.Fatalf("sortColumn = %v, want 0", got)
	}
	if got := props["sortDescending"]; got != true {
		t.Fatalf("sortDescending = %v, want true", got)
	}
	if got := props["statusText"]; got != "Page 2/4 · Total 76 · Size 25" {
		t.Fatalf("statusText = %v, want server pagination summary", got)
	}
}

func TestDataTableSortStateShortcut(t *testing.T) {
	node := DataTable(
		[]TableColumn{{Title: "Provider", Width: 16, Sortable: true}},
		[][]string{{"openai"}},
		DataTableSortState(0, false),
	)

	props := node.Props()
	if got := props["sortColumn"]; got != 0 {
		t.Fatalf("sortColumn = %v, want first column", got)
	}
	if got := props["sortDescending"]; got != false {
		t.Fatalf("sortDescending = %v, want false", got)
	}
	if got := props["sortControlled"]; got != true {
		t.Fatalf("sortControlled = %v, want true", got)
	}

	unset := DataTable(
		[]TableColumn{{Title: "Provider", Width: 16, Sortable: true}},
		[][]string{{"openai"}},
		DataTableSortState(-1, true),
	)
	if got := unset.Props()["sortControlled"]; got != false {
		t.Fatalf("unset sortControlled = %v, want false", got)
	}
}

type dataTableTestIntent struct{ name string }

func (i dataTableTestIntent) IntentType() string { return i.name }

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

func TestDataTableStableKeyShortcuts(t *testing.T) {
	node := DataTable(
		[]TableColumn{{Title: "Provider", Width: 16}},
		[][]string{{"openai"}, {"azure"}},
		DataTableRowKeys([]string{"provider.openai", "provider.azure"}),
		DataTableSelectedKey("provider.azure"),
		DataTableSelectedKeyField("selected_provider_key"),
		DataTableActivateKeyField("activated_provider_key"),
	)

	props := node.Props()
	if rowKeys, ok := props["rowKeys"].([]string); !ok || len(rowKeys) != 2 || rowKeys[1] != "provider.azure" {
		t.Fatalf("rowKeys = %#v, want provider row keys", props["rowKeys"])
	}
	if got := props["selectedRowKey"]; got != "provider.azure" {
		t.Fatalf("selectedRowKey = %v, want provider.azure", got)
	}
	if _, ok := props["selectedKeyIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("selectedKeyIntentField = %T, want intent.FieldIntent", props["selectedKeyIntentField"])
	}
	if _, ok := props["activateKeyIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("activateKeyIntentField = %T, want intent.FieldIntent", props["activateKeyIntentField"])
	}
}

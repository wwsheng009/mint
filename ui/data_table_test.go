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

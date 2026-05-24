package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/table"
)

func TestDataTableShortcut(t *testing.T) {
	node := DataTable(
		[]TableColumn{
			{Title: "Name", Width: 12},
			{Title: "Status", Width: 10},
		},
		[][]string{
			{"provider-a", "healthy"},
			{"provider-b", "degraded"},
		},
		DataTableKey("providers"),
		DataTableComponentID("providers.table"),
		DataTablePageSize(8),
		DataTableSelectedIndex(1),
		DataTableSelectedField("selectedProvider"),
		DataTableSearch("provider"),
		DataTableEmptyText("No providers"),
		DataTableShowFooter(true),
		DataTableShowScrollbar(true),
		DataTableOperationalStyle(),
	)

	props := node.Props()
	if got := props["key"]; got != "providers" {
		t.Fatalf("key = %v, want providers", got)
	}
	if got := props["componentID"]; got != "providers.table" {
		t.Fatalf("componentID = %v, want providers.table", got)
	}
	if got := props["pageSize"]; got != 8 {
		t.Fatalf("pageSize = %v, want 8", got)
	}
	if got := props["selectedIndex"]; got != 1 {
		t.Fatalf("selectedIndex = %v, want 1", got)
	}
	if got := props["selectedIndexControlled"]; got != true {
		t.Fatalf("selectedIndexControlled = %v, want true", got)
	}
	if _, ok := props["changeIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("changeIntentField = %T, want intent.FieldIntent", props["changeIntentField"])
	}
	if got := props["searchQuery"]; got != "provider" {
		t.Fatalf("searchQuery = %v, want provider", got)
	}
	if got := props["emptyText"]; got != "No providers" {
		t.Fatalf("emptyText = %v, want No providers", got)
	}
	if got := props["showFooter"]; got != true {
		t.Fatalf("showFooter = %v, want true", got)
	}
	if got := props["showScrollbar"]; got != true {
		t.Fatalf("showScrollbar = %v, want true", got)
	}
	if columns, ok := props["columns"].([]table.TableColumn); !ok || len(columns) != 2 {
		t.Fatalf("columns = %#v, want two columns", props["columns"])
	}
	if rows, ok := props["rows"].([][]string); !ok || len(rows) != 2 {
		t.Fatalf("rows = %#v, want two rows", props["rows"])
	}
	if headerStyle, ok := props["headerStyle"].(style.Style); !ok || !headerStyle.IsBold() {
		t.Fatalf("headerStyle = %#v, want bold style", props["headerStyle"])
	}
	if selectedStyle, ok := props["selectedStyle"].(style.Style); !ok || selectedStyle.BG != style.Color("cyan") {
		t.Fatalf("selectedStyle = %#v, want cyan background", props["selectedStyle"])
	}
}

func TestDataTableCustomOption(t *testing.T) {
	node := DataTable(
		[]TableColumn{{Title: "Name"}},
		[][]string{{"a"}},
		func(cfg *DataTableConfig) {
			cfg.EmptyText = "custom empty"
			cfg.ShowFooter = boolPtr(false)
		},
	)

	props := node.Props()
	if got := props["emptyText"]; got != "custom empty" {
		t.Fatalf("emptyText = %v, want custom empty", got)
	}
	if got := props["showFooter"]; got != false {
		t.Fatalf("showFooter = %v, want false", got)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

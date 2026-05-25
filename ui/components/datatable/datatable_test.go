package datatable

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui/components/table"
)

func TestNew(t *testing.T) {
	node := New(
		[]table.TableColumn{
			{Title: "Name", Width: 12},
			{Title: "Status", Width: 10},
		},
		[][]string{
			{"provider-a", "healthy"},
			{"provider-b", "degraded"},
		},
		Key("providers"),
		ComponentID("providers.table"),
		PageSize(8),
		SelectedIndex(1),
		SelectedField("selectedProvider"),
		Search("provider"),
		EmptyText("No providers"),
		ShowFooter(true),
		ShowScrollbar(true),
		OperationalStyle(),
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

func TestCustomOption(t *testing.T) {
	node := New(
		[]table.TableColumn{{Title: "Name"}},
		[][]string{{"a"}},
		func(cfg *Config) {
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

func TestLoadingState(t *testing.T) {
	node := New(
		[]table.TableColumn{{Title: "Name"}},
		[][]string{{"provider-a"}},
		Loading(true),
		LoadingText("Loading providers..."),
	)

	props := node.Props()
	if rows, ok := props["rows"].([][]string); !ok || len(rows) != 0 {
		t.Fatalf("rows = %#v, want empty rows while loading", props["rows"])
	}
	if got := props["emptyText"]; got != "Loading providers..." {
		t.Fatalf("emptyText = %v, want loading text", got)
	}
	if got := props["statusText"]; got != "Loading" {
		t.Fatalf("statusText = %v, want Loading", got)
	}
}

func TestErrorState(t *testing.T) {
	node := New(
		[]table.TableColumn{{Title: "Name"}},
		[][]string{{"provider-a"}},
		ErrorText("gateway API unavailable"),
	)

	props := node.Props()
	if rows, ok := props["rows"].([][]string); !ok || len(rows) != 0 {
		t.Fatalf("rows = %#v, want empty rows for error state", props["rows"])
	}
	if got := props["emptyText"]; got != "gateway API unavailable" {
		t.Fatalf("emptyText = %v, want error text", got)
	}
	if got := props["statusText"]; got != "Error · gateway API unavailable" {
		t.Fatalf("statusText = %v, want error status", got)
	}
}

func TestServerPaginationStatus(t *testing.T) {
	node := New(
		[]table.TableColumn{{Title: "Name"}},
		[][]string{{"provider-a"}, {"provider-b"}},
		ServerPagination(2, 25, 76),
	)

	props := node.Props()
	if got := props["statusText"]; got != "Page 2/4 · Total 76 · Size 25" {
		t.Fatalf("statusText = %v, want server pagination summary", got)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

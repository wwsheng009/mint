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
		CurrentPage(2),
		SelectedIndex(1),
		SelectedField("selectedProvider"),
		PageField("providerPage"),
		Search("provider"),
		SortBy(1, true),
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
	if got := props["currentPage"]; got != 2 {
		t.Fatalf("currentPage = %v, want 2", got)
	}
	if got := props["currentPageControlled"]; got != true {
		t.Fatalf("currentPageControlled = %v, want true", got)
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
	if _, ok := props["pageIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("pageIntentField = %T, want intent.FieldIntent", props["pageIntentField"])
	}
	if got := props["searchQuery"]; got != "provider" {
		t.Fatalf("searchQuery = %v, want provider", got)
	}
	if got := props["sortColumn"]; got != 1 {
		t.Fatalf("sortColumn = %v, want 1", got)
	}
	if got := props["sortDescending"]; got != true {
		t.Fatalf("sortDescending = %v, want true", got)
	}
	if got := props["sortControlled"]; got != true {
		t.Fatalf("sortControlled = %v, want true", got)
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

func TestOperationalPreset(t *testing.T) {
	node := Operational(
		[]table.TableColumn{{Title: "Provider", Width: 16}},
		[][]string{{"openai"}, {"azure"}},
		1,
		25,
		"selectedProviderIndex",
		RowKeys([]string{"provider.openai", "provider.azure"}),
		SelectedKey("provider.azure"),
		SelectedKeyField("selectedProviderKey"),
		ActivateKeyField("activatedProviderKey"),
		ServerPagination(2, 25, 76),
	)

	props := node.Props()
	if got := props["pageSize"]; got != 25 {
		t.Fatalf("pageSize = %v, want 25", got)
	}
	if got := props["selectedIndex"]; got != 1 {
		t.Fatalf("selectedIndex = %v, want 1", got)
	}
	if got := props["showFooter"]; got != true {
		t.Fatalf("showFooter = %v, want true", got)
	}
	if got := props["showScrollbar"]; got != true {
		t.Fatalf("showScrollbar = %v, want true", got)
	}
	if _, ok := props["changeIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("changeIntentField = %T, want intent.FieldIntent", props["changeIntentField"])
	}
	if rowKeys, ok := props["rowKeys"].([]string); !ok || len(rowKeys) != 2 || rowKeys[1] != "provider.azure" {
		t.Fatalf("rowKeys = %#v, want provider row keys", props["rowKeys"])
	}
	if got := props["selectedRowKey"]; got != "provider.azure" {
		t.Fatalf("selectedRowKey = %v, want provider.azure", got)
	}
	if got := props["statusText"]; got != "Page 2/4 · Total 76 · Size 25" {
		t.Fatalf("statusText = %v, want server pagination summary", got)
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

func TestSortByOption(t *testing.T) {
	node := New(
		[]table.TableColumn{
			{Title: "Provider", Sortable: true},
			{Title: "Status", Sortable: true},
		},
		[][]string{{"openai", "healthy"}, {"azure", "degraded"}},
		SortBy(1, false),
	)

	props := node.Props()
	if got := props["sortColumn"]; got != 1 {
		t.Fatalf("sortColumn = %v, want 1", got)
	}
	if got := props["sortDescending"]; got != false {
		t.Fatalf("sortDescending = %v, want false", got)
	}
	if got := props["sortControlled"]; got != true {
		t.Fatalf("sortControlled = %v, want true", got)
	}
}

func TestSortStateOption(t *testing.T) {
	firstColumn := New(
		[]table.TableColumn{
			{Title: "Provider", Sortable: true},
			{Title: "Status", Sortable: true},
		},
		[][]string{{"openai", "healthy"}, {"azure", "degraded"}},
		SortState(0, false),
	)

	props := firstColumn.Props()
	if got := props["sortColumn"]; got != 0 {
		t.Fatalf("sortColumn = %v, want first column", got)
	}
	if got := props["sortDescending"]; got != false {
		t.Fatalf("sortDescending = %v, want false", got)
	}
	if got := props["sortControlled"]; got != true {
		t.Fatalf("sortControlled = %v, want true", got)
	}

	unset := New(
		[]table.TableColumn{{Title: "Provider", Sortable: true}},
		[][]string{{"openai"}},
		SortState(-1, true),
	)
	unsetProps := unset.Props()
	if got := unsetProps["sortColumn"]; got != -1 {
		t.Fatalf("unset sortColumn = %v, want -1", got)
	}
	if got := unsetProps["sortControlled"]; got != false {
		t.Fatalf("unset sortControlled = %v, want false", got)
	}
}

func TestControlledStateOptions(t *testing.T) {
	changeIntent := datatableTestIntent{name: "datatable.change"}
	node := New(
		[]table.TableColumn{
			{Title: "Provider", Sortable: true},
			{Title: "Status", Sortable: true},
		},
		[][]string{{"openai", "healthy"}, {"azure", "degraded"}},
		ComponentID("tokens.table"),
		PageSize(25),
		CurrentPage(3),
		PageField("tokenPage"),
		SortBy(1, true),
		OnChange(changeIntent),
	)

	props := node.Props()
	if got := props["componentID"]; got != "tokens.table" {
		t.Fatalf("componentID = %v, want tokens.table", got)
	}
	if got := props["currentPage"]; got != 3 {
		t.Fatalf("currentPage = %v, want 3", got)
	}
	if got := props["currentPageControlled"]; got != true {
		t.Fatalf("currentPageControlled = %v, want true", got)
	}
	if _, ok := props["pageIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("pageIntentField = %T, want intent.FieldIntent", props["pageIntentField"])
	}
	if got := props["sortColumn"]; got != 1 {
		t.Fatalf("sortColumn = %v, want 1", got)
	}
	if got := props["sortDescending"]; got != true {
		t.Fatalf("sortDescending = %v, want true", got)
	}
	if got := props["changeIntent"]; got != changeIntent {
		t.Fatalf("changeIntent = %#v, want %#v", got, changeIntent)
	}
}

func TestStableRowKeyOptions(t *testing.T) {
	node := New(
		[]table.TableColumn{{Title: "Provider"}},
		[][]string{{"openai"}, {"azure"}},
		RowKeys([]string{"provider.openai", "provider.azure"}),
		SelectedKey("provider.azure"),
		SelectedKeyField("selected_provider_key"),
		ActivateKeyField("activated_provider_key"),
	)

	props := node.Props()
	if rowKeys, ok := props["rowKeys"].([]string); !ok || len(rowKeys) != 2 || rowKeys[1] != "provider.azure" {
		t.Fatalf("rowKeys = %#v, want provider row keys", props["rowKeys"])
	}
	if got := props["selectedRowKey"]; got != "provider.azure" {
		t.Fatalf("selectedRowKey = %v, want provider.azure", got)
	}
	if got := props["selectedRowKeyControlled"]; got != true {
		t.Fatalf("selectedRowKeyControlled = %v, want true", got)
	}
	if _, ok := props["selectedKeyIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("selectedKeyIntentField = %T, want intent.FieldIntent", props["selectedKeyIntentField"])
	}
	if _, ok := props["activateKeyIntentField"].(intent.FieldIntent); !ok {
		t.Fatalf("activateKeyIntentField = %T, want intent.FieldIntent", props["activateKeyIntentField"])
	}
}

type datatableTestIntent struct{ name string }

func (i datatableTestIntent) IntentType() string { return i.name }

func boolPtr(v bool) *bool {
	return &v
}

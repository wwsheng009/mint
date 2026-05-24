package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/table"
)

// DataTableOption configures a DataTable shortcut.
type DataTableOption func(*DataTableConfig)

// DataTableConfig stores configuration for DataTable.
type DataTableConfig struct {
	Key              string
	ComponentID      string
	PageSize         int
	SelectedIndex    int
	HasSelectedIndex bool
	SelectedField    string
	SearchQuery      string
	EmptyText        string
	ShowFooter       *bool
	ShowScrollbar    *bool
	HeaderStyle      style.Style
	HasHeaderStyle   bool
	SelectedStyle    style.Style
	HasSelectedStyle bool
	TableStyle       style.Style
	HasTableStyle    bool
}

// DataTable builds a table with common data-application options.
func DataTable(columns []TableColumn, rows [][]string, opts ...DataTableOption) rtui.VNode {
	cfg := DataTableConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	builder := table.NewBuilder().
		Columns(columns).
		Rows(rows)
	if cfg.Key != "" {
		builder.Key(cfg.Key)
	}
	if cfg.ComponentID != "" {
		builder.ComponentID(cfg.ComponentID)
	}
	if cfg.PageSize > 0 {
		builder.PageSize(cfg.PageSize)
	}
	if cfg.HasSelectedIndex {
		builder.SelectedIndex(cfg.SelectedIndex)
	}
	if cfg.SelectedField != "" {
		builder.ForField(intent.BindField(cfg.SelectedField))
	}
	if cfg.SearchQuery != "" {
		builder.SearchQuery(cfg.SearchQuery)
	}
	if cfg.EmptyText != "" {
		builder.EmptyText(cfg.EmptyText)
	}
	if cfg.ShowFooter != nil {
		builder.ShowFooter(*cfg.ShowFooter)
	}
	if cfg.ShowScrollbar != nil {
		builder.ShowScrollbar(*cfg.ShowScrollbar)
	}
	if cfg.HasHeaderStyle {
		builder.HeaderStyle(cfg.HeaderStyle)
	}
	if cfg.HasSelectedStyle {
		builder.SelectedStyle(cfg.SelectedStyle)
	}
	if cfg.HasTableStyle {
		builder.TableStyle(cfg.TableStyle)
	}
	return builder.Build()
}

// DataTableKey sets the VNode key.
func DataTableKey(key string) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.Key = key
	}
}

// DataTableComponentID sets the table component ID for state-change intents.
func DataTableComponentID(componentID string) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.ComponentID = componentID
	}
}

// DataTablePageSize enables pagination with the given page size.
func DataTablePageSize(pageSize int) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.PageSize = pageSize
	}
}

// DataTableSelectedIndex controls the selected source row index.
func DataTableSelectedIndex(index int) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.SelectedIndex = index
		cfg.HasSelectedIndex = true
	}
}

// DataTableSelectedField binds the selected row index to FieldChangeIntent.
func DataTableSelectedField(field string) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.SelectedField = field
	}
}

// DataTableSearch applies case-insensitive search across all columns.
func DataTableSearch(query string) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.SearchQuery = query
	}
}

// DataTableEmptyText sets the empty-state text.
func DataTableEmptyText(text string) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.EmptyText = text
	}
}

// DataTableShowFooter toggles the table footer.
func DataTableShowFooter(show bool) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.ShowFooter = &show
	}
}

// DataTableShowScrollbar toggles the table scrollbar indicator.
func DataTableShowScrollbar(show bool) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.ShowScrollbar = &show
	}
}

// DataTableHeaderStyle sets the header style.
func DataTableHeaderStyle(headerStyle style.Style) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.HeaderStyle = headerStyle
		cfg.HasHeaderStyle = true
	}
}

// DataTableSelectedStyle sets the selected row style.
func DataTableSelectedStyle(selectedStyle style.Style) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.SelectedStyle = selectedStyle
		cfg.HasSelectedStyle = true
	}
}

// DataTableStyle sets the table body style.
func DataTableStyle(tableStyle style.Style) DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.TableStyle = tableStyle
		cfg.HasTableStyle = true
	}
}

// DataTableOperationalStyle applies a compact cyan-highlight style suitable for operational tools.
func DataTableOperationalStyle() DataTableOption {
	return func(cfg *DataTableConfig) {
		cfg.HeaderStyle = style.NewStyle().Foreground(style.Color("cyan")).Bold(true)
		cfg.HasHeaderStyle = true
		cfg.SelectedStyle = style.NewStyle().
			Foreground(style.Color("black")).
			Background(style.Color("cyan"))
		cfg.HasSelectedStyle = true
	}
}

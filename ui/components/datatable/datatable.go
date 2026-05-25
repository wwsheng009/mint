package datatable

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/table"
)

// Option configures a DataTable shortcut.
type Option func(*Config)

// Config stores DataTable configuration.
type Config struct {
	Key              string
	ComponentID      string
	PageSize         int
	SelectedIndex    int
	HasSelectedIndex bool
	SelectedField    string
	SearchQuery      string
	EmptyText        string
	Loading          bool
	LoadingText      string
	ErrorText        string
	StatusText       string
	ServerPage       int
	ServerPageSize   int
	ServerTotal      int
	HasServerPage    bool
	ShowFooter       *bool
	ShowScrollbar    *bool
	HeaderStyle      style.Style
	HasHeaderStyle   bool
	SelectedStyle    style.Style
	HasSelectedStyle bool
	TableStyle       style.Style
	HasTableStyle    bool
}

// New builds a table with common data-application options.
func New(columns []table.TableColumn, rows [][]string, opts ...Option) rtui.VNode {
	cfg := Config{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	rows = normalizeRowsForState(rows, cfg)
	builder := table.NewBuilder().Columns(columns).Rows(rows)
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
	if emptyText := emptyTextForState(cfg); emptyText != "" {
		builder.EmptyText(emptyText)
	}
	if statusText := statusTextForConfig(cfg); statusText != "" {
		builder.StatusText(statusText)
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

func Key(key string) Option {
	return func(cfg *Config) {
		cfg.Key = key
	}
}

func ComponentID(componentID string) Option {
	return func(cfg *Config) {
		cfg.ComponentID = componentID
	}
}

func PageSize(pageSize int) Option {
	return func(cfg *Config) {
		cfg.PageSize = pageSize
	}
}

func SelectedIndex(index int) Option {
	return func(cfg *Config) {
		cfg.SelectedIndex = index
		cfg.HasSelectedIndex = true
	}
}

func SelectedField(field string) Option {
	return func(cfg *Config) {
		cfg.SelectedField = field
	}
}

func Search(query string) Option {
	return func(cfg *Config) {
		cfg.SearchQuery = query
	}
}

func EmptyText(text string) Option {
	return func(cfg *Config) {
		cfg.EmptyText = text
	}
}

func Loading(loading bool) Option {
	return func(cfg *Config) {
		cfg.Loading = loading
	}
}

func LoadingText(text string) Option {
	return func(cfg *Config) {
		cfg.LoadingText = text
	}
}

func ErrorText(text string) Option {
	return func(cfg *Config) {
		cfg.ErrorText = text
	}
}

func StatusText(text string) Option {
	return func(cfg *Config) {
		cfg.StatusText = text
	}
}

func ServerPagination(page, pageSize, total int) Option {
	return func(cfg *Config) {
		cfg.ServerPage = page
		cfg.ServerPageSize = pageSize
		cfg.ServerTotal = total
		cfg.HasServerPage = true
	}
}

func ShowFooter(show bool) Option {
	return func(cfg *Config) {
		cfg.ShowFooter = &show
	}
}

func ShowScrollbar(show bool) Option {
	return func(cfg *Config) {
		cfg.ShowScrollbar = &show
	}
}

func HeaderStyle(headerStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.HeaderStyle = headerStyle
		cfg.HasHeaderStyle = true
	}
}

func SelectedStyle(selectedStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.SelectedStyle = selectedStyle
		cfg.HasSelectedStyle = true
	}
}

func normalizeRowsForState(rows [][]string, cfg Config) [][]string {
	if cfg.Loading || strings.TrimSpace(cfg.ErrorText) != "" {
		return nil
	}
	return rows
}

func emptyTextForState(cfg Config) string {
	if cfg.Loading {
		if text := strings.TrimSpace(cfg.LoadingText); text != "" {
			return text
		}
		return "Loading..."
	}
	if errorText := strings.TrimSpace(cfg.ErrorText); errorText != "" {
		return errorText
	}
	return cfg.EmptyText
}

func statusTextForConfig(cfg Config) string {
	if statusText := strings.TrimSpace(cfg.StatusText); statusText != "" {
		return statusText
	}
	if cfg.Loading {
		return "Loading"
	}
	if errorText := strings.TrimSpace(cfg.ErrorText); errorText != "" {
		return "Error · " + errorText
	}
	if cfg.HasServerPage {
		return formatServerPagination(cfg.ServerPage, cfg.ServerPageSize, cfg.ServerTotal)
	}
	return ""
}

func formatServerPagination(page, pageSize, total int) string {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if total < 0 {
		total = 0
	}
	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if page > totalPages {
		page = totalPages
	}
	return fmt.Sprintf("Page %d/%d · Total %d · Size %d", page, totalPages, total, pageSize)
}

func TableStyle(tableStyle style.Style) Option {
	return func(cfg *Config) {
		cfg.TableStyle = tableStyle
		cfg.HasTableStyle = true
	}
}

func OperationalStyle() Option {
	return func(cfg *Config) {
		cfg.HeaderStyle = style.NewStyle().Foreground(style.Color("cyan")).Bold(true)
		cfg.HasHeaderStyle = true
		cfg.SelectedStyle = style.NewStyle().
			Foreground(style.Color("black")).
			Background(style.Color("cyan"))
		cfg.HasSelectedStyle = true
	}
}

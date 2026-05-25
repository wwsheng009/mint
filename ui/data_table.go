package ui

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/datatable"
)

type DataTableOption = datatable.Option
type DataTableConfig = datatable.Config

func DataTable(columns []TableColumn, rows [][]string, opts ...DataTableOption) rtui.VNode {
	return datatable.New(columns, rows, opts...)
}

func DataTableKey(key string) DataTableOption {
	return datatable.Key(key)
}

func DataTableComponentID(componentID string) DataTableOption {
	return datatable.ComponentID(componentID)
}

func DataTableRowKeys(keys []string) DataTableOption {
	return datatable.RowKeys(keys)
}

func DataTablePageSize(pageSize int) DataTableOption {
	return datatable.PageSize(pageSize)
}

func DataTableSelectedIndex(index int) DataTableOption {
	return datatable.SelectedIndex(index)
}

func DataTableSelectedKey(key string) DataTableOption {
	return datatable.SelectedKey(key)
}

func DataTableSelectedField(field string) DataTableOption {
	return datatable.SelectedField(field)
}

func DataTableSelectedKeyField(field string) DataTableOption {
	return datatable.SelectedKeyField(field)
}

func DataTableSearch(query string) DataTableOption {
	return datatable.Search(query)
}

func DataTableEmptyText(text string) DataTableOption {
	return datatable.EmptyText(text)
}

func DataTableLoading(loading bool) DataTableOption {
	return datatable.Loading(loading)
}

func DataTableLoadingText(text string) DataTableOption {
	return datatable.LoadingText(text)
}

func DataTableErrorText(text string) DataTableOption {
	return datatable.ErrorText(text)
}

func DataTableStatusText(text string) DataTableOption {
	return datatable.StatusText(text)
}

func DataTableServerPagination(page, pageSize, total int) DataTableOption {
	return datatable.ServerPagination(page, pageSize, total)
}

func DataTableShowFooter(show bool) DataTableOption {
	return datatable.ShowFooter(show)
}

func DataTableShowScrollbar(show bool) DataTableOption {
	return datatable.ShowScrollbar(show)
}

func DataTableHeaderStyle(headerStyle style.Style) DataTableOption {
	return datatable.HeaderStyle(headerStyle)
}

func DataTableSelectedStyle(selectedStyle style.Style) DataTableOption {
	return datatable.SelectedStyle(selectedStyle)
}

func DataTableStyle(tableStyle style.Style) DataTableOption {
	return datatable.TableStyle(tableStyle)
}

func DataTableOnActivate(activateIntent intent.Intent) DataTableOption {
	return datatable.OnActivate(activateIntent)
}

func DataTableActivateField(field string) DataTableOption {
	return datatable.ActivateField(field)
}

func DataTableActivateKeyField(field string) DataTableOption {
	return datatable.ActivateKeyField(field)
}

func DataTableOperationalStyle() DataTableOption {
	return datatable.OperationalStyle()
}

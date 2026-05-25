package table

import "github.com/wwsheng009/mint/runtime/intent"

// StateChangeIntent is emitted when interactive table state changes.
// It bubbles locally through the instance tree.
type StateChangeIntent struct {
	ComponentID           string
	SelectedIndex         int
	SelectedSourceIndex   int
	SelectedRowKey        string
	ExpandedSourceIndices []int
	CurrentPage           int
	PageCount             int
	PageSize              int
	SortColumn            int
	SortDescending        bool
	SearchQuery           string
	Filters               map[int]string
	VisibleRows           int
	FilteredRows          int
	TotalRows             int
}

func (StateChangeIntent) IntentType() string {
	return "Table:StateChange"
}

func (StateChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

func (StateChangeIntent) IsTransition() bool {
	return false
}

func (StateChangeIntent) IsGlobal() bool {
	return true
}

func StateChange(
	componentID string,
	selectedIndex int,
	selectedSourceIndex int,
	selectedRowKey string,
	expandedSourceIndices []int,
	currentPage int,
	pageCount int,
	pageSize int,
	sortColumn int,
	sortDescending bool,
	searchQuery string,
	filters map[int]string,
	visibleRows int,
	filteredRows int,
	totalRows int,
) StateChangeIntent {
	return StateChangeIntent{
		ComponentID:           componentID,
		SelectedIndex:         selectedIndex,
		SelectedSourceIndex:   selectedSourceIndex,
		SelectedRowKey:        selectedRowKey,
		ExpandedSourceIndices: append([]int(nil), expandedSourceIndices...),
		CurrentPage:           currentPage,
		PageCount:             pageCount,
		PageSize:              pageSize,
		SortColumn:            sortColumn,
		SortDescending:        sortDescending,
		SearchQuery:           searchQuery,
		Filters:               cloneFilters(filters),
		VisibleRows:           visibleRows,
		FilteredRows:          filteredRows,
		TotalRows:             totalRows,
	}
}

// ActivateIntent is emitted when the current table row is explicitly activated.
// It is separate from StateChangeIntent so callers can distinguish navigation
// from an affirmative enter/click action on the focused row.
type ActivateIntent struct {
	ComponentID         string
	SelectedIndex       int
	SelectedSourceIndex int
	SelectedRowKey      string
	Row                 []string
}

func (ActivateIntent) IntentType() string {
	return "Table:Activate"
}

func (ActivateIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

func (ActivateIntent) IsTransition() bool {
	return false
}

func (ActivateIntent) IsGlobal() bool {
	return true
}

func Activate(
	componentID string,
	selectedIndex int,
	selectedSourceIndex int,
	selectedRowKey string,
	row []string,
) ActivateIntent {
	return ActivateIntent{
		ComponentID:         componentID,
		SelectedIndex:       selectedIndex,
		SelectedSourceIndex: selectedSourceIndex,
		SelectedRowKey:      selectedRowKey,
		Row:                 append([]string(nil), row...),
	}
}

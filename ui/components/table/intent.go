package table

import "github.com/wwsheng009/mint/runtime/intent"

// StateChangeIntent is emitted when interactive table state changes.
// It bubbles locally through the instance tree.
type StateChangeIntent struct {
	ComponentID         string
	SelectedIndex       int
	SelectedSourceIndex int
	CurrentPage         int
	PageCount           int
	PageSize            int
	SortColumn          int
	SortDescending      bool
	SearchQuery         string
	Filters             map[int]string
	VisibleRows         int
	FilteredRows        int
	TotalRows           int
}

func (StateChangeIntent) IntentType() string {
	return "Table:StateChange"
}

func (StateChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (StateChangeIntent) IsTransition() bool {
	return true
}

func (StateChangeIntent) IsGlobal() bool {
	return true
}

func StateChange(
	componentID string,
	selectedIndex int,
	selectedSourceIndex int,
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
		ComponentID:         componentID,
		SelectedIndex:       selectedIndex,
		SelectedSourceIndex: selectedSourceIndex,
		CurrentPage:         currentPage,
		PageCount:           pageCount,
		PageSize:            pageSize,
		SortColumn:          sortColumn,
		SortDescending:      sortDescending,
		SearchQuery:         searchQuery,
		Filters:             cloneFilters(filters),
		VisibleRows:         visibleRows,
		FilteredRows:        filteredRows,
		TotalRows:           totalRows,
	}
}

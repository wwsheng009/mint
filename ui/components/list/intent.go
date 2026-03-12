package list

import "github.com/wwsheng009/mint/runtime/intent"

// StateChangeIntent is emitted when interactive list state changes.
// It is intended for external controllers that track list state by ComponentID.
type StateChangeIntent struct {
	ComponentID    string
	SelectedIndex  int
	SelectedRow    string
	ScrollOffset   int
	ViewportHeight int
	VisibleRows    int
	TotalRows      int
	SelectionMode  SelectionMode
	CheckedIndices []int
	CheckedRows    []string
}

func (StateChangeIntent) IntentType() string {
	return "List:StateChange"
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

func (i StateChangeIntent) GetComponentID() string {
	return i.ComponentID
}

// RowSelectIntent is emitted when the selected row changes.
type RowSelectIntent struct {
	SelectedIndex int
	SelectedRow   string
	ComponentID   string
}

func (RowSelectIntent) IntentType() string              { return "list.RowSelectIntent" }
func (RowSelectIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (RowSelectIntent) IsTransition() bool              { return false }
func (RowSelectIntent) IsGlobal() bool                  { return false }
func (i RowSelectIntent) GetComponentID() string        { return i.ComponentID }

// NavigationIntent is emitted when keyboard or command navigation changes selection.
type NavigationIntent struct {
	Direction   string
	FromIndex   int
	ToIndex     int
	ComponentID string
}

func (NavigationIntent) IntentType() string              { return "list.NavigationIntent" }
func (NavigationIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (NavigationIntent) IsTransition() bool              { return false }
func (NavigationIntent) IsGlobal() bool                  { return false }
func (i NavigationIntent) GetComponentID() string        { return i.ComponentID }

// ScrollIntent is emitted when the viewport scroll offset changes.
type ScrollIntent struct {
	Offset      int
	Delta       int
	ViewSize    int
	ContentSize int
	ComponentID string
}

func (ScrollIntent) IntentType() string              { return "list.ScrollIntent" }
func (ScrollIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ScrollIntent) IsTransition() bool              { return false }
func (ScrollIntent) IsGlobal() bool                  { return false }
func (i ScrollIntent) GetComponentID() string        { return i.ComponentID }

// SelectionChangeIntent is emitted when checkbox selection changes.
type SelectionChangeIntent struct {
	SelectionMode  SelectionMode
	CheckedIndices []int
	CheckedRows    []string
	ComponentID    string
}

func (SelectionChangeIntent) IntentType() string              { return "list.SelectionChangeIntent" }
func (SelectionChangeIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SelectionChangeIntent) IsTransition() bool              { return false }
func (SelectionChangeIntent) IsGlobal() bool                  { return false }
func (i SelectionChangeIntent) GetComponentID() string        { return i.ComponentID }

// SearchNextIntent requests moving selection to the next search match.
type SearchNextIntent struct {
	ComponentID string
}

func (SearchNextIntent) IntentType() string              { return "list.SearchNextIntent" }
func (SearchNextIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SearchNextIntent) IsTransition() bool              { return false }
func (SearchNextIntent) IsGlobal() bool                  { return false }
func (i SearchNextIntent) GetComponentID() string        { return i.ComponentID }

// SearchPrevIntent requests moving selection to the previous search match.
type SearchPrevIntent struct {
	ComponentID string
}

func (SearchPrevIntent) IntentType() string              { return "list.SearchPrevIntent" }
func (SearchPrevIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SearchPrevIntent) IsTransition() bool              { return false }
func (SearchPrevIntent) IsGlobal() bool                  { return false }
func (i SearchPrevIntent) GetComponentID() string        { return i.ComponentID }

// SearchStatsIntent reports current search query and match stats.
type SearchStatsIntent struct {
	Query       string
	Total       int
	Selected    int
	ComponentID string
}

func (SearchStatsIntent) IntentType() string              { return "list.SearchStatsIntent" }
func (SearchStatsIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SearchStatsIntent) IsTransition() bool              { return false }
func (SearchStatsIntent) IsGlobal() bool                  { return false }
func (i SearchStatsIntent) GetComponentID() string        { return i.ComponentID }

func StateChange(
	componentID string,
	selectedIndex int,
	selectedRow string,
	scrollOffset int,
	viewportHeight int,
	visibleRows int,
	totalRows int,
	selectionMode SelectionMode,
	checkedIndices []int,
	checkedRows []string,
) StateChangeIntent {
	return StateChangeIntent{
		ComponentID:    componentID,
		SelectedIndex:  selectedIndex,
		SelectedRow:    selectedRow,
		ScrollOffset:   scrollOffset,
		ViewportHeight: viewportHeight,
		VisibleRows:    visibleRows,
		TotalRows:      totalRows,
		SelectionMode:  selectionMode,
		CheckedIndices: append([]int(nil), checkedIndices...),
		CheckedRows:    append([]string(nil), checkedRows...),
	}
}

func RowSelect(selectedIndex int, selectedRow string) RowSelectIntent {
	return RowSelectIntent{
		SelectedIndex: selectedIndex,
		SelectedRow:   selectedRow,
	}
}

func RowSelectWithID(componentID string, selectedIndex int, selectedRow string) RowSelectIntent {
	return RowSelectIntent{
		SelectedIndex: selectedIndex,
		SelectedRow:   selectedRow,
		ComponentID:   componentID,
	}
}

func Navigation(direction string, fromIndex, toIndex int) NavigationIntent {
	return NavigationIntent{
		Direction: direction,
		FromIndex: fromIndex,
		ToIndex:   toIndex,
	}
}

func NavigationWithID(componentID, direction string, fromIndex, toIndex int) NavigationIntent {
	return NavigationIntent{
		Direction:   direction,
		FromIndex:   fromIndex,
		ToIndex:     toIndex,
		ComponentID: componentID,
	}
}

func Scroll(offset, delta, viewSize, contentSize int) ScrollIntent {
	return ScrollIntent{
		Offset:      offset,
		Delta:       delta,
		ViewSize:    viewSize,
		ContentSize: contentSize,
	}
}

func ScrollWithID(componentID string, offset, delta, viewSize, contentSize int) ScrollIntent {
	return ScrollIntent{
		Offset:      offset,
		Delta:       delta,
		ViewSize:    viewSize,
		ContentSize: contentSize,
		ComponentID: componentID,
	}
}

func SelectionChange(mode SelectionMode, checkedIndices []int, checkedRows []string) SelectionChangeIntent {
	return SelectionChangeIntent{
		SelectionMode:  mode,
		CheckedIndices: append([]int(nil), checkedIndices...),
		CheckedRows:    append([]string(nil), checkedRows...),
	}
}

func SelectionChangeWithID(componentID string, mode SelectionMode, checkedIndices []int, checkedRows []string) SelectionChangeIntent {
	return SelectionChangeIntent{
		SelectionMode:  mode,
		CheckedIndices: append([]int(nil), checkedIndices...),
		CheckedRows:    append([]string(nil), checkedRows...),
		ComponentID:    componentID,
	}
}

func SearchNext(componentID string) SearchNextIntent {
	return SearchNextIntent{ComponentID: componentID}
}

func SearchPrev(componentID string) SearchPrevIntent {
	return SearchPrevIntent{ComponentID: componentID}
}

func SearchStats(query string, total, selected int) SearchStatsIntent {
	return SearchStatsIntent{
		Query:    query,
		Total:    total,
		Selected: selected,
	}
}

func SearchStatsWithID(componentID, query string, total, selected int) SearchStatsIntent {
	return SearchStatsIntent{
		Query:       query,
		Total:       total,
		Selected:    selected,
		ComponentID: componentID,
	}
}

// SelectNextIntent requests moving the selected row forward by one.
type SelectNextIntent struct {
	ComponentID string
}

func (SelectNextIntent) IntentType() string              { return "list.SelectNextIntent" }
func (SelectNextIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SelectNextIntent) IsTransition() bool              { return false }
func (SelectNextIntent) IsGlobal() bool                  { return false }
func (i SelectNextIntent) GetComponentID() string        { return i.ComponentID }

// SelectPrevIntent requests moving the selected row backward by one.
type SelectPrevIntent struct {
	ComponentID string
}

func (SelectPrevIntent) IntentType() string              { return "list.SelectPrevIntent" }
func (SelectPrevIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SelectPrevIntent) IsTransition() bool              { return false }
func (SelectPrevIntent) IsGlobal() bool                  { return false }
func (i SelectPrevIntent) GetComponentID() string        { return i.ComponentID }

// SelectByIndexIntent requests selecting a specific row index.
// Index = -1 clears the row selection.
type SelectByIndexIntent struct {
	Index       int
	ComponentID string
}

func (SelectByIndexIntent) IntentType() string              { return "list.SelectByIndexIntent" }
func (SelectByIndexIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (SelectByIndexIntent) IsTransition() bool              { return false }
func (SelectByIndexIntent) IsGlobal() bool                  { return false }
func (i SelectByIndexIntent) GetComponentID() string        { return i.ComponentID }

// ClearSelectionIntent clears the currently selected row.
type ClearSelectionIntent struct {
	ComponentID string
}

func (ClearSelectionIntent) IntentType() string              { return "list.ClearSelectionIntent" }
func (ClearSelectionIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ClearSelectionIntent) IsTransition() bool              { return false }
func (ClearSelectionIntent) IsGlobal() bool                  { return false }
func (i ClearSelectionIntent) GetComponentID() string        { return i.ComponentID }

// ScrollToIntent requests moving the viewport to an absolute offset.
type ScrollToIntent struct {
	Offset      int
	ComponentID string
}

func (ScrollToIntent) IntentType() string              { return "list.ScrollToIntent" }
func (ScrollToIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ScrollToIntent) IsTransition() bool              { return false }
func (ScrollToIntent) IsGlobal() bool                  { return false }
func (i ScrollToIntent) GetComponentID() string        { return i.ComponentID }

// ScrollByIntent requests moving the viewport by a delta.
type ScrollByIntent struct {
	Delta       int
	ComponentID string
}

func (ScrollByIntent) IntentType() string              { return "list.ScrollByIntent" }
func (ScrollByIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ScrollByIntent) IsTransition() bool              { return false }
func (ScrollByIntent) IsGlobal() bool                  { return false }
func (i ScrollByIntent) GetComponentID() string        { return i.ComponentID }

// ToggleCheckedIntent toggles a checkbox selection for the given row index.
type ToggleCheckedIntent struct {
	Index       int
	ComponentID string
}

func (ToggleCheckedIntent) IntentType() string              { return "list.ToggleCheckedIntent" }
func (ToggleCheckedIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ToggleCheckedIntent) IsTransition() bool              { return false }
func (ToggleCheckedIntent) IsGlobal() bool                  { return false }
func (i ToggleCheckedIntent) GetComponentID() string        { return i.ComponentID }

// ClearCheckedIntent clears all checkbox selections.
type ClearCheckedIntent struct {
	ComponentID string
}

func (ClearCheckedIntent) IntentType() string              { return "list.ClearCheckedIntent" }
func (ClearCheckedIntent) Priority() intent.ActionPriority { return intent.PriorityNormal }
func (ClearCheckedIntent) IsTransition() bool              { return false }
func (ClearCheckedIntent) IsGlobal() bool                  { return false }
func (i ClearCheckedIntent) GetComponentID() string        { return i.ComponentID }

func SelectNext(componentID string) SelectNextIntent {
	return SelectNextIntent{ComponentID: componentID}
}

func SelectPrev(componentID string) SelectPrevIntent {
	return SelectPrevIntent{ComponentID: componentID}
}

func SelectByIndex(componentID string, index int) SelectByIndexIntent {
	return SelectByIndexIntent{ComponentID: componentID, Index: index}
}

func ClearSelection(componentID string) ClearSelectionIntent {
	return ClearSelectionIntent{ComponentID: componentID}
}

func ScrollTo(componentID string, offset int) ScrollToIntent {
	return ScrollToIntent{ComponentID: componentID, Offset: offset}
}

func ScrollBy(componentID string, delta int) ScrollByIntent {
	return ScrollByIntent{ComponentID: componentID, Delta: delta}
}

func ToggleChecked(componentID string, index int) ToggleCheckedIntent {
	return ToggleCheckedIntent{ComponentID: componentID, Index: index}
}

func ClearChecked(componentID string) ClearCheckedIntent {
	return ClearCheckedIntent{ComponentID: componentID}
}

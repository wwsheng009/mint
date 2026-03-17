package pagination

import "github.com/wwsheng009/mint/runtime/intent"

// PageChangeIntent is emitted when pagination changes page.
type PageChangeIntent struct {
	ComponentID string
	FromPage    int
	ToPage      int
	PageCount   int
	PageSize    int
	Total       int
}

func (PageChangeIntent) IntentType() string {
	return "pagination.PageChangeIntent"
}

func (PageChangeIntent) Priority() intent.ActionPriority {
	return intent.PriorityNormal
}

func (PageChangeIntent) IsTransition() bool {
	return false
}

func (PageChangeIntent) IsGlobal() bool {
	return false
}

func (i PageChangeIntent) GetComponentID() string {
	return i.ComponentID
}

func PageChange(fromPage, toPage, pageCount, pageSize, total int) PageChangeIntent {
	return PageChangeIntent{
		FromPage:  fromPage,
		ToPage:    toPage,
		PageCount: pageCount,
		PageSize:  pageSize,
		Total:     total,
	}
}

func PageChangeWithID(componentID string, fromPage, toPage, pageCount, pageSize, total int) PageChangeIntent {
	intent := PageChange(fromPage, toPage, pageCount, pageSize, total)
	intent.ComponentID = componentID
	return intent
}

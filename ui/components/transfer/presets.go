package transfer

import "strings"

const (
	defaultOperationalListWidth  = 28
	defaultOperationalListHeight = 8
	defaultOperationalPageSize   = 20
)

// NewItemWithDescription creates a transfer item with secondary searchable text.
func NewItemWithDescription(key, title, description string) Item {
	return NewItem(key, title).WithDescription(description)
}

// DisabledItem creates a disabled transfer item with an optional description.
func DisabledItem(key, title, description string) Item {
	return NewItemWithDescription(key, title, description).WithDisabled(true)
}

// WithDescription returns a copy of the item with secondary searchable text.
func (item Item) WithDescription(description string) Item {
	item.Description = normalizeInlineText(description)
	return item
}

// WithDisabled returns a copy of the item with the disabled state set.
func (item Item) WithDisabled(disabled bool) Item {
	item.Disabled = disabled
	return item
}

// OperationalAssignment builds a transfer for common operational assignment
// flows such as scope, permission, provider, key, or alert target selection.
func OperationalAssignment(componentID string, items []Item, targetKeys []string) *Builder {
	builder := NewBuilder().
		ComponentID(strings.TrimSpace(componentID)).
		Titles("Available", "Selected").
		Operations("Add", "Remove").
		BulkOperations(true).
		BulkOperationLabels("Add visible", "Remove visible").
		Searchable(true).
		SearchPlaceholders("Search available", "Search selected").
		PageSize(defaultOperationalPageSize).
		ListWidth(defaultOperationalListWidth).
		ListHeight(defaultOperationalListHeight).
		Items(items).
		TargetKeys(targetKeys)
	if strings.TrimSpace(componentID) != "" {
		builder.Key(componentID)
	}
	return builder
}

func normalizeInlineText(value string) string {
	return strings.TrimSpace(strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	).Replace(value))
}

package list

import "strings"

// RowItem is the structured row model for List.
// It can be flattened into the legacy string row format for compatibility.
type RowItem struct {
	Title       string
	Description string
	Prefix      string
	Suffix      string
}

// Item creates a structured list item with the provided title.
func Item(title string) RowItem {
	return RowItem{Title: title}
}

// WithDescription appends a secondary description to the row text.
func (i RowItem) WithDescription(description string) RowItem {
	i.Description = description
	return i
}

// WithPrefix prepends a short leading marker to the row text.
func (i RowItem) WithPrefix(prefix string) RowItem {
	i.Prefix = prefix
	return i
}

// WithSuffix appends a short trailing marker to the row text.
func (i RowItem) WithSuffix(suffix string) RowItem {
	i.Suffix = suffix
	return i
}

// Text flattens the structured row into the legacy single-line representation.
func (i RowItem) Text() string {
	parts := make([]string, 0, 2)
	if i.Prefix != "" {
		parts = append(parts, i.Prefix)
	}

	main := i.Title
	if main == "" {
		main = i.Description
	}
	if i.Title != "" && i.Description != "" {
		main += " - " + i.Description
	}
	if main != "" {
		parts = append(parts, main)
	}
	if i.Suffix != "" {
		parts = append(parts, i.Suffix)
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

func cloneItems(items []RowItem) []RowItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]RowItem, len(items))
	copy(out, items)
	return out
}

func itemsFromRows(rows []string) []RowItem {
	if len(rows) == 0 {
		return nil
	}
	items := make([]RowItem, len(rows))
	for i, row := range rows {
		items[i] = Item(row)
	}
	return items
}

func rowsFromItems(items []RowItem) []string {
	if len(items) == 0 {
		return nil
	}
	rows := make([]string, len(items))
	for i, item := range items {
		rows[i] = item.Text()
	}
	return rows
}

func normalizeItemsAndRows(items []RowItem, rows []string) ([]RowItem, []string) {
	switch {
	case items != nil:
		clonedItems := cloneItems(items)
		return clonedItems, rowsFromItems(clonedItems)
	case rows != nil:
		clonedRows := append([]string(nil), rows...)
		return itemsFromRows(clonedRows), clonedRows
	default:
		return nil, nil
	}
}

func equalItems(left, right []RowItem) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

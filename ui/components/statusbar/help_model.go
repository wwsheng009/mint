package statusbar

import "sync"

type helpEntry struct {
	text    string
	hovered bool
	focused bool
	order   int
	bounds  [4]int
}

type helpModel struct {
	mu       sync.RWMutex
	fallback string
	prefix   string
	entries  map[string]helpEntry
}

func newHelpModel(fallback, prefix string) *helpModel {
	return &helpModel{
		fallback: fallback,
		prefix:   prefix,
		entries:  make(map[string]helpEntry),
	}
}

func (m *helpModel) Update(key string, order int, text string, hovered, focused bool, bounds [4]int) {
	if m == nil || key == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	text = normalizeStatusText(text)
	if text == "" {
		delete(m.entries, key)
		return
	}
	m.entries[key] = helpEntry{
		text:    text,
		hovered: hovered,
		focused: focused,
		order:   order,
		bounds:  bounds,
	}
}

func (m *helpModel) Remove(key string) {
	if m == nil || key == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, key)
}

func (m *helpModel) Current() string {
	if m == nil {
		return ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	current := m.pickText(true)
	if current == "" {
		current = m.pickText(false)
	}
	if current == "" {
		current = m.fallback
	}
	if current == "" {
		return ""
	}
	if m.prefix != "" {
		return m.prefix + current
	}
	return current
}

func (m *helpModel) Active() (string, [4]int, bool) {
	if m == nil {
		return "", [4]int{}, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.pickEntry(true)
	if !ok {
		entry, ok = m.pickEntry(false)
	}
	return m.prefixedEntry(entry, ok)
}

func (m *helpModel) HoveredActive() (string, [4]int, bool) {
	if m == nil {
		return "", [4]int{}, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.pickEntry(true)
	return m.prefixedEntry(entry, ok)
}

func (m *helpModel) HasContent() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.fallback != "" {
		return true
	}
	for _, entry := range m.entries {
		if entry.text != "" {
			return true
		}
	}
	return false
}

func (m *helpModel) pickText(hover bool) string {
	entry, ok := m.pickEntry(hover)
	if !ok {
		return ""
	}
	return entry.text
}

func (m *helpModel) pickEntry(hover bool) (helpEntry, bool) {
	bestOrder := int(^uint(0) >> 1)
	var best helpEntry
	found := false
	for _, entry := range m.entries {
		active := entry.focused
		if hover {
			active = entry.hovered
		}
		if !active || entry.text == "" {
			continue
		}
		if entry.order < bestOrder {
			bestOrder = entry.order
			best = entry
			found = true
		}
	}
	return best, found
}

func (m *helpModel) prefixedEntry(entry helpEntry, ok bool) (string, [4]int, bool) {
	if !ok || entry.text == "" {
		return "", [4]int{}, false
	}
	text := entry.text
	if m.prefix != "" {
		text = m.prefix + text
	}
	return text, entry.bounds, true
}

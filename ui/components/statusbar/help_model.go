package statusbar

import "sync"

type helpEntry struct {
	text    string
	hovered bool
	focused bool
	order   int
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

func (m *helpModel) Update(key string, order int, text string, hovered, focused bool) {
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
	bestOrder := int(^uint(0) >> 1)
	bestText := ""
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
			bestText = entry.text
		}
	}
	return bestText
}

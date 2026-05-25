package menu

import (
	"strings"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type Variant string

const (
	VariantMenuBar  Variant = "menubar"
	VariantDropdown Variant = "dropdown"
	VariantContext  Variant = "context"
	VariantPopup    Variant = "popup"
)

type ItemKind string

const (
	ItemAction    ItemKind = "action"
	ItemSeparator ItemKind = "separator"
	ItemCheckbox  ItemKind = "checkbox"
	ItemRadio     ItemKind = "radio"
	ItemSubmenu   ItemKind = "submenu"
	ItemCustom    ItemKind = "custom"
	ItemLabel     ItemKind = "label"
)

type ShortcutScope string

const (
	ShortcutLocal  ShortcutScope = "local"
	ShortcutGlobal ShortcutScope = "global"
)

type Placement string

const (
	PlacementBottomStart Placement = "bottom-start"
	PlacementBottomEnd   Placement = "bottom-end"
	PlacementTopStart    Placement = "top-start"
	PlacementTopEnd      Placement = "top-end"
	PlacementRightStart  Placement = "right-start"
	PlacementLeftStart   Placement = "left-start"
	PlacementAuto        Placement = "auto"
)

type Shortcut struct {
	Key            string
	Combo          string
	Display        string
	Scope          ShortcutScope
	Enabled        bool
	Priority       int
	When           string
	PreventDefault bool
}

func (s Shortcut) NormalizedCombo() string {
	return normalizeCombo(s.Combo)
}

func (s Shortcut) DisplayText() string {
	if strings.TrimSpace(s.Display) != "" {
		return s.Display
	}
	return s.Combo
}

type MenuItem struct {
	Key           string
	Label         string
	SecondaryText string
	Description   string
	Icon          string
	Shortcut      Shortcut
	Kind          ItemKind
	Disabled      bool
	Hidden        bool
	Checked       bool
	Danger        bool
	Group         string
	Children      []MenuItem
	Intent        intent.Intent
	OnSelect      func() intent.Intent
	CloseOnSelect bool
	KeepOpen      bool
	TestID        string
	ThemeSlot     string
	Metadata      map[string]any
}

type Model struct {
	ID                string
	ComponentID       string
	Variant           Variant
	Title             string
	Items             []MenuItem
	PathPrefix        []int
	Open              bool
	Layer             rtui.Layer
	PortalRoot        string
	AnchorID          string
	Anchor            rttypes.Anchor
	PortalPosition    rttypes.PositionType
	PortalPriority    int
	PortalOffsetX     int
	PortalOffsetY     int
	Placement         Placement
	ActivePath        []int
	ActiveIndex       int
	SelectedIndex     int
	ScrollOffset      int
	MaxWidth          int
	MaxHeight         int
	MinWidth          int
	ShowShortcuts     bool
	ShowDescriptions  bool
	ShowCheckMarks    bool
	ShowIcons         bool
	Scrollable        bool
	CloseOnOutside    bool
	CloseOnEscape     bool
	Typeahead         bool
	TypeaheadTimeout  time.Duration
	RegisterShortcuts bool
}

type OpenMenuIntent struct {
	MenuID string
	Path   []int
}

func (OpenMenuIntent) IntentType() string { return "menu.open" }
func (OpenMenuIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

type CloseMenuIntent struct {
	MenuID string
}

func (CloseMenuIntent) IntentType() string { return "menu.close" }
func (CloseMenuIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

type ActivateMenuItemIntent struct {
	MenuID  string
	ItemKey string
	Path    []int
}

func (ActivateMenuItemIntent) IntentType() string { return "menu.activate_item" }
func (ActivateMenuItemIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

type NavigateMenuIntent struct {
	MenuID    string
	Direction string
	FromIndex int
	ToIndex   int
}

func (NavigateMenuIntent) IntentType() string { return "menu.navigate" }
func (NavigateMenuIntent) Priority() intent.ActionPriority {
	return intent.PriorityUserBlocking
}

func Items(items ...MenuItem) []MenuItem {
	return cloneItems(items)
}

func Action(key, label string, pressIntent intent.Intent) MenuItem {
	return MenuItem{
		Key:           key,
		Label:         label,
		Kind:          ItemAction,
		Intent:        pressIntent,
		CloseOnSelect: true,
	}
}

func Checkbox(key, label string, checked bool, pressIntent intent.Intent) MenuItem {
	return MenuItem{
		Key:           key,
		Label:         label,
		Kind:          ItemCheckbox,
		Checked:       checked,
		Intent:        pressIntent,
		CloseOnSelect: true,
	}
}

func Radio(key, label, group string, checked bool, pressIntent intent.Intent) MenuItem {
	return MenuItem{
		Key:           key,
		Label:         label,
		Kind:          ItemRadio,
		Group:         group,
		Checked:       checked,
		Intent:        pressIntent,
		CloseOnSelect: true,
	}
}

func Submenu(key, label string, children ...MenuItem) MenuItem {
	return MenuItem{
		Key:      key,
		Label:    label,
		Kind:     ItemSubmenu,
		Children: cloneItems(children),
	}
}

func Separator() MenuItem {
	return MenuItem{Kind: ItemSeparator}
}

func LabelItem(key, label string) MenuItem {
	return MenuItem{Key: key, Label: label, Kind: ItemLabel}
}

func (m MenuItem) Normalize() MenuItem {
	if m.Kind == "" {
		if len(m.Children) > 0 {
			m.Kind = ItemSubmenu
		} else {
			m.Kind = ItemAction
		}
	}
	m.Children = cloneItems(m.Children)
	return m
}

func (m MenuItem) WithShortcut(combo string) MenuItem {
	m.Shortcut.Combo = combo
	if m.Shortcut.Display == "" {
		m.Shortcut.Display = combo
	}
	return m
}

func (m MenuItem) WithShortcutDisplay(display string) MenuItem {
	m.Shortcut.Display = display
	return m
}

func (m MenuItem) WithShortcutScope(scope ShortcutScope) MenuItem {
	m.Shortcut.Scope = scope
	return m
}

func (m MenuItem) WithDescription(text string) MenuItem {
	m.Description = text
	return m
}

func (m MenuItem) WithSecondaryText(text string) MenuItem {
	m.SecondaryText = text
	return m
}

func (m MenuItem) WithIcon(icon string) MenuItem {
	m.Icon = icon
	return m
}

func (m MenuItem) WithChildren(children ...MenuItem) MenuItem {
	m.Children = cloneItems(children)
	if len(children) > 0 && m.Kind == "" {
		m.Kind = ItemSubmenu
	}
	return m
}

func (m MenuItem) WithChecked(checked bool) MenuItem {
	m.Checked = checked
	return m
}

func (m MenuItem) WithDisabled(disabled bool) MenuItem {
	m.Disabled = disabled
	return m
}

// WithDisabledReason disables the item and records a user-facing reason.
//
// The reason is stored as metadata under "disabledReason". If the item has no
// description yet, the reason is also used as the description so menus rendered
// with ShowDescriptions(true) can explain why the item is unavailable.
func (m MenuItem) WithDisabledReason(reason string) MenuItem {
	m.Disabled = true
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return m
	}
	if strings.TrimSpace(m.Description) == "" {
		m.Description = reason
	}
	return m.WithMetadata("disabledReason", reason)
}

func (m MenuItem) WithDanger(danger bool) MenuItem {
	m.Danger = danger
	return m
}

func (m MenuItem) WithKeepOpen(keepOpen bool) MenuItem {
	m.KeepOpen = keepOpen
	return m
}

func (m MenuItem) WithCloseOnSelect(closeOnSelect bool) MenuItem {
	m.CloseOnSelect = closeOnSelect
	return m
}

func (m MenuItem) WithHidden(hidden bool) MenuItem {
	m.Hidden = hidden
	return m
}

func (m MenuItem) WithThemeSlot(slot string) MenuItem {
	m.ThemeSlot = slot
	return m
}

func (m MenuItem) WithOnSelect(fn func() intent.Intent) MenuItem {
	m.OnSelect = fn
	return m
}

func (m MenuItem) WithMetadata(key string, value any) MenuItem {
	if m.Metadata == nil {
		m.Metadata = map[string]any{}
	}
	m.Metadata[key] = value
	return m
}

func (m MenuItem) IsVisible() bool {
	return !m.Hidden
}

func (m MenuItem) IsSeparator() bool {
	return m.Normalize().Kind == ItemSeparator
}

func (m MenuItem) IsLabel() bool {
	return m.Normalize().Kind == ItemLabel
}

func (m MenuItem) HasSubmenu() bool {
	normalized := m.Normalize()
	return normalized.Kind == ItemSubmenu || len(normalized.Children) > 0
}

func (m MenuItem) IsSelectable() bool {
	normalized := m.Normalize()
	return normalized.IsVisible() && !normalized.Disabled && normalized.Kind != ItemSeparator && normalized.Kind != ItemLabel
}

func (m MenuItem) EffectiveIntent() intent.Intent {
	if m.OnSelect != nil {
		return m.OnSelect()
	}
	return m.Intent
}

func (m MenuItem) EffectiveCloseOnSelect() bool {
	if m.KeepOpen {
		return false
	}
	if m.HasSubmenu() {
		return false
	}
	if m.CloseOnSelect {
		return true
	}
	return true
}

// Package toolbar provides a Fiber-first operation toolbar for data and admin pages.
package toolbar

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
	statusbarcomp "github.com/wwsheng009/mint/ui/components/statusbar"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

const (
	propActions      = "actions"
	propCenterItems  = "centerItems"
	propDense        = "dense"
	propGap          = "gap"
	propKey          = "key"
	propLeftItems    = "leftItems"
	propRightItems   = "rightItems"
	propSeparator    = "separator"
	propStyle        = "style"
	propTitle        = "title"
	propTitleWidth   = "titleWidth"
	propUseStatusBar = "useStatusBar"
	propWidth        = "width"
)

// ItemKind controls how a Toolbar item renders.
type ItemKind string

const (
	ItemText      ItemKind = "text"
	ItemBadge     ItemKind = "badge"
	ItemButton    ItemKind = "button"
	ItemMenu      ItemKind = "menu"
	ItemSeparator ItemKind = "separator"
	ItemCustom    ItemKind = "custom"
)

// Item describes one visible Toolbar entry.
type Item struct {
	Key            string
	Label          string
	Kind           ItemKind
	PressIntent    intent.Intent
	Variant        button.Variant
	Disabled       bool
	HelpText       string
	DisabledReason string
	Width          int
	FgColor        string
	BgColor        string
	Bold           bool
	Custom         rtui.VNode

	MenuID               string
	MenuItems            []menucomp.MenuItem
	MenuOpen             bool
	MenuPlacement        menucomp.Placement
	MenuActivePath       []int
	MenuMinWidth         int
	MenuMaxHeight        int
	MenuShowShortcuts    bool
	MenuShowDescriptions bool
}

// SelectionConfig describes one Up/Down navigation surface for a selected item.
type SelectionConfig struct {
	Key           string
	Title         string
	Width         int
	PrevIntent    intent.Intent
	NextIntent    intent.Intent
	Busy          bool
	Index         int
	Total         int
	ItemLabel     string
	LoadingReason string
}

// PaginationConfig describes one Prev/Next page navigation surface.
type PaginationConfig struct {
	Key           string
	Width         int
	Summary       Item
	PrevIntent    intent.Intent
	NextIntent    intent.Intent
	Busy          bool
	Page          int
	TotalPages    int
	LoadingReason string
}

// PaginationScope holds normalized server-pagination numbers for one page
// navigation surface.
type PaginationScope struct {
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

// PageSummaryPart describes one extra "label value" context segment appended to
// a pagination summary.
type PageSummaryPart struct {
	Label string
	Value string
}

// ActionTargetPart describes one "label value" segment in an operation target
// summary. It is intended for target context near dangerous or auditable action
// groups, not for filter or pagination scope.
type ActionTargetPart struct {
	Label string
	Value string
}

// VNode is the declarative description of a Toolbar.
type VNode struct {
	*rtui.ElementVNode

	key          string
	title        string
	titleWidth   int
	width        int
	gap          int
	dense        bool
	separator    string
	useStatusBar bool
	leftItems    []Item
	centerItems  []Item
	rightItems   []Item
	rootStyle    style.Style
}

// ActionGroupConfig describes one titled group in a grouped operation surface.
type ActionGroupConfig struct {
	Key        string
	Title      string
	Width      int
	TitleWidth int
	Items      []Item
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a Toolbar VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("toolbar"),
		gap:          1,
		separator:    "|",
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to the key.
func (v *VNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *VNode) Tag() string { return "toolbar" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propActions:      cloneItems(v.rightItems),
		propCenterItems:  cloneItems(v.centerItems),
		propDense:        v.dense,
		propGap:          v.gap,
		propKey:          v.key,
		propLeftItems:    cloneItems(v.leftItems),
		propRightItems:   cloneItems(v.rightItems),
		propSeparator:    v.separator,
		propStyle:        v.rootStyle,
		propTitle:        v.title,
		propTitleWidth:   v.titleWidth,
		propUseStatusBar: v.useStatusBar,
		propWidth:        v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	v.key = getStringProp(props, propKey, v.key)
	v.title = getStringProp(props, propTitle, v.title)
	v.titleWidth = getIntProp(props, propTitleWidth, v.titleWidth)
	v.width = getIntProp(props, propWidth, v.width)
	v.gap = getIntProp(props, propGap, v.gap)
	v.dense = getBoolProp(props, propDense, v.dense)
	v.separator = getStringProp(props, propSeparator, v.separator)
	v.useStatusBar = getBoolProp(props, propUseStatusBar, v.useStatusBar)
	v.leftItems = normalizeItems(getItemsProp(props, propLeftItems, v.leftItems))
	v.centerItems = normalizeItems(getItemsProp(props, propCenterItems, v.centerItems))
	v.rightItems = normalizeItems(getItemsProp(props, propRightItems, getItemsProp(props, propActions, v.rightItems)))
	v.rootStyle = getStyleProp(props, propStyle, v.rootStyle)
	v.normalize()
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetTitleWidth(width int) *VNode {
	v.titleWidth = width
	v.normalize()
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	v.normalize()
	return v
}

func (v *VNode) SetGap(gap int) *VNode {
	v.gap = gap
	v.normalize()
	return v
}

func (v *VNode) SetDense(dense bool) *VNode {
	v.dense = dense
	return v
}

func (v *VNode) SetSeparator(separator string) *VNode {
	v.separator = separator
	return v
}

func (v *VNode) SetUseStatusBar(use bool) *VNode {
	v.useStatusBar = use
	return v
}

func (v *VNode) SetRootStyle(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) SetLeftItems(items []Item) *VNode {
	v.leftItems = normalizeItems(items)
	return v
}

func (v *VNode) AddLeftItem(item Item) *VNode {
	v.leftItems = normalizeItems(append(v.leftItems, item))
	return v
}

func (v *VNode) SetCenterItems(items []Item) *VNode {
	v.centerItems = normalizeItems(items)
	return v
}

func (v *VNode) AddCenterItem(item Item) *VNode {
	v.centerItems = normalizeItems(append(v.centerItems, item))
	return v
}

func (v *VNode) SetRightItems(items []Item) *VNode {
	v.rightItems = normalizeItems(items)
	return v
}

func (v *VNode) AddRightItem(item Item) *VNode {
	v.rightItems = normalizeItems(append(v.rightItems, item))
	return v
}

func (v *VNode) LeftItems() []Item { return cloneItems(v.leftItems) }

func (v *VNode) CenterItems() []Item { return cloneItems(v.centerItems) }

func (v *VNode) RightItems() []Item { return cloneItems(v.rightItems) }

func (v *VNode) normalize() {
	v.leftItems = normalizeItems(v.leftItems)
	v.centerItems = normalizeItems(v.centerItems)
	v.rightItems = normalizeItems(v.rightItems)
	if v.width < 0 {
		v.width = 0
	}
	if v.titleWidth < 0 {
		v.titleWidth = 0
	}
	if v.gap < 0 {
		v.gap = 0
	}
}

// Text creates a plain toolbar item.
func Text(key, label string) Item {
	return Item{Key: key, Label: label, Kind: ItemText}
}

// Badge creates a highlighted toolbar item.
func Badge(key, label string) Item {
	return Item{Key: key, Label: label, Kind: ItemBadge, Bold: true}
}

// Button creates a command item.
func Button(key, label string, pressIntent intent.Intent) Item {
	return Item{Key: key, Label: label, Kind: ItemButton, PressIntent: pressIntent}
}

// Dropdown creates a toolbar button that can render an anchored menu popup.
//
// The open flag is intentionally controlled by the application state. Pressing
// the dropdown emits menu.OpenMenuIntent by default; reducers should set
// MenuOpen(true) on the next render and close it through the existing menu
// close/outside-click flow.
func Dropdown(key, label string, items []menucomp.MenuItem, open bool) Item {
	return Item{
		Key:               key,
		Label:             label,
		Kind:              ItemMenu,
		MenuItems:         menucomp.NormalizeItems(items),
		MenuOpen:          open,
		MenuPlacement:     menucomp.PlacementBottomStart,
		MenuShowShortcuts: true,
	}
}

// Menu is an alias for Dropdown.
func Menu(key, label string, items []menucomp.MenuItem, open bool) Item {
	return Dropdown(key, label, items, open)
}

// Separator creates a visual separator item.
func Separator(key string) Item {
	return Item{Key: key, Kind: ItemSeparator}
}

// Custom creates a custom toolbar item.
func Custom(key string, node rtui.VNode) Item {
	return Item{Key: key, Kind: ItemCustom, Custom: node}
}

// KeyValue creates a compact "label: value" toolbar item for operational scopes.
func KeyValue(key, label, value string) Item {
	label = strings.TrimSpace(normalizeToolbarText(label))
	value = strings.TrimSpace(normalizeToolbarText(value))
	if value == "" {
		value = "-"
	}
	text := value
	if label != "" {
		text = label + ": " + value
	}
	return Text(key, text)
}

// MutedKeyValue creates a compact low-emphasis "label: value" toolbar item.
func MutedKeyValue(key, label, value string) Item {
	return KeyValue(key, label, value).WithForeground("bright-black")
}

// StateBadge creates a highlighted toolbar item using common operational tone mapping.
func StateBadge(key, status string) Item {
	label := strings.TrimSpace(normalizeToolbarText(status))
	if label == "" {
		label = "-"
	}
	fg, bg := toneColors(statusbarcomp.DefaultTone(label))
	return Badge(key, label).WithColors(fg, bg)
}

// BusyBadge creates a warning-colored toolbar item for running operations.
func BusyBadge(key, label string) Item {
	label = strings.TrimSpace(normalizeToolbarText(label))
	if label == "" {
		label = "busy"
	}
	return Badge(key, label).WithColors("black", "yellow")
}

// ErrorBadge creates an error-colored toolbar item.
func ErrorBadge(key, label string) Item {
	label = strings.TrimSpace(normalizeToolbarText(label))
	if label == "" {
		label = "error"
	}
	return Badge(key, label).WithColors("white", "red")
}

// Endpoint creates a standard toolbar item for the active API endpoint.
func Endpoint(value string) Item {
	return KeyValue("endpoint", "endpoint", value)
}

// Scope creates a standard toolbar item for the current page scope.
func Scope(value string) Item {
	return KeyValue("scope", "scope", value)
}

// Selection creates a low-emphasis toolbar item for the current selection.
func Selection(value string) Item {
	return MutedKeyValue("selection", "selection", value)
}

// ShellHeader creates a standard top-level operations shell header.
//
// It keeps the first global row ordered as application identity, endpoint or
// scope context, auth/session state, then global actions such as refresh/logout.
func ShellHeader(key, title, endpoint, auth string, actions ...Item) rtui.VNode {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "shell.header"
	}
	title = strings.TrimSpace(normalizeToolbarText(title))
	if title == "" {
		title = "-"
	}
	endpoint = strings.TrimSpace(normalizeToolbarText(endpoint))
	if endpoint == "" {
		endpoint = "-"
	}
	auth = strings.TrimSpace(normalizeToolbarText(auth))
	if auth == "" {
		auth = "-"
	}
	return NewBuilder().
		Key(key).
		Title(title).
		TitleWidth(22).
		Left(Text("base-url", endpoint).WithForeground("bright-black")).
		Center(Badge("auth", auth).WithColors("black", "cyan")).
		RightItems(actions).
		Build()
}

// ShellNav creates a dense top-level navigation toolbar from precomputed items.
//
// Callers keep ownership of route, permission, active-state and intent logic;
// the preset only standardizes the shell navigation row layout.
func ShellNav(key string, items []Item) rtui.VNode {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "shell.nav"
	}
	return NewBuilder().
		Key(key).
		Dense(true).
		LeftItems(items).
		Build()
}

// PageSummary creates a standard pagination summary item with optional scope
// context such as provider, search, status, or source.
func PageSummary(page, totalPages, total, pageSize, width int, parts ...PageSummaryPart) Item {
	if page <= 0 {
		page = 1
	}
	if totalPages <= 0 {
		totalPages = 1
	}
	if total < 0 {
		total = 0
	}
	if pageSize < 0 {
		pageSize = 0
	}
	segments := []string{fmt.Sprintf("page %d/%d total %d size %d", page, totalPages, total, pageSize)}
	for _, part := range parts {
		label := strings.TrimSpace(part.Label)
		value := strings.TrimSpace(part.Value)
		if label == "" && value == "" {
			continue
		}
		if value == "" {
			value = "-"
		}
		if label == "" {
			segments = append(segments, value)
			continue
		}
		segments = append(segments, label+" "+value)
	}
	if width <= 0 {
		width = 72
	}
	return Text("page", strings.Join(segments, " ")).WithWidth(width)
}

// CompactPageSummaryPart creates a display-width-bounded pagination scope
// segment. It is intended for user-entered search, provider, status, or source
// values that share one toolbar row with Prev/Next controls.
func CompactPageSummaryPart(label, value string, maxWidth int) PageSummaryPart {
	return PageSummaryPart{
		Label: label,
		Value: textcomp.CompactFallbackText(value, "-", maxWidth),
	}
}

// PageSummaryPartIfValue creates a pagination scope segment only when value is
// non-empty after trimming. It is useful for optional filters such as search
// text, where omitting the inactive scope is clearer than rendering "search -".
func PageSummaryPartIfValue(label, value string) PageSummaryPart {
	if strings.TrimSpace(value) == "" {
		return PageSummaryPart{}
	}
	return PageSummaryPart{Label: label, Value: value}
}

// PageSummaryPartUnless creates a pagination scope segment only when value is
// non-empty and differs from defaultValue after trimming.
func PageSummaryPartUnless(label, value, defaultValue string) PageSummaryPart {
	if pageSummaryValueIsDefault(value, defaultValue) {
		return PageSummaryPart{}
	}
	return PageSummaryPart{Label: label, Value: value}
}

// CompactPageSummaryPartIfValue creates a display-width-bounded pagination
// scope segment only when value is non-empty after trimming.
func CompactPageSummaryPartIfValue(label, value string, maxWidth int) PageSummaryPart {
	if strings.TrimSpace(value) == "" {
		return PageSummaryPart{}
	}
	return CompactPageSummaryPart(label, value, maxWidth)
}

// CompactPageSummaryPartUnless creates a display-width-bounded pagination
// scope segment only when value is non-empty and differs from defaultValue.
func CompactPageSummaryPartUnless(label, value, defaultValue string, maxWidth int) PageSummaryPart {
	if pageSummaryValueIsDefault(value, defaultValue) {
		return PageSummaryPart{}
	}
	return CompactPageSummaryPart(label, value, maxWidth)
}

func pageSummaryValueIsDefault(value, defaultValue string) bool {
	value = strings.TrimSpace(value)
	defaultValue = strings.TrimSpace(defaultValue)
	return value == "" || value == defaultValue
}

// ActionTargetSummary creates a compact operation target summary from label/value
// parts, for example "group default · provider openai · key ****". Blank values
// render as "-" so missing target context is visible before an operation is
// prepared.
func ActionTargetSummary(parts ...ActionTargetPart) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		label := strings.TrimSpace(normalizeToolbarText(part.Label))
		value := textcomp.DisplayText(part.Value)
		if label == "" && value == "-" {
			continue
		}
		if label == "" {
			segments = append(segments, value)
			continue
		}
		segments = append(segments, label+" "+value)
	}
	return strings.Join(segments, " · ")
}

// ActionTargetSummaryWithScope creates an operation target summary with an
// explicit active operation scope prefix, for example
// "scope provider · group default · provider openai · key ****".
func ActionTargetSummaryWithScope(scope string, parts ...ActionTargetPart) string {
	scope = strings.TrimSpace(normalizeToolbarText(scope))
	if scope == "" {
		return ActionTargetSummary(parts...)
	}
	scoped := make([]ActionTargetPart, 0, len(parts)+1)
	scoped = append(scoped, ActionTargetPart{Label: "scope", Value: scope})
	scoped = append(scoped, parts...)
	return ActionTargetSummary(scoped...)
}

// ActionTargetSummaryWithScopes creates an operation target summary for an
// action area that may expose one or more operation scopes at the same time.
// A single scope keeps the ActionTargetSummaryWithScope format; multiple
// scopes render as a plural "scopes" prefix, for example
// "scopes global/group/key · endpoint http://127.0.0.1:8080 · group default".
func ActionTargetSummaryWithScopes(scopes []string, parts ...ActionTargetPart) string {
	normalizedScopes := normalizeActionTargetScopes(scopes)
	switch len(normalizedScopes) {
	case 0:
		return ActionTargetSummary(parts...)
	case 1:
		return ActionTargetSummaryWithScope(normalizedScopes[0], parts...)
	default:
		scoped := make([]ActionTargetPart, 0, len(parts)+1)
		scoped = append(scoped, ActionTargetPart{Label: "scopes", Value: strings.Join(normalizedScopes, "/")})
		scoped = append(scoped, parts...)
		return ActionTargetSummary(scoped...)
	}
}

func normalizeActionTargetScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(normalizeToolbarText(scope))
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	return normalized
}

// CompactActionTargetPart creates a display-width-bounded operation target
// segment for group/provider/key/token identifiers that share one action area.
func CompactActionTargetPart(label, value string, maxWidth int) ActionTargetPart {
	return ActionTargetPart{
		Label: label,
		Value: textcomp.CompactFallbackText(value, "-", maxWidth),
	}
}

// NormalizePaginationScope clamps server-pagination fields and applies a
// fallback total when the server total is not available.
func NormalizePaginationScope(page, pageSize, total, totalPages, fallbackTotal int) PaginationScope {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 25
	}
	if total <= 0 {
		total = fallbackTotal
	}
	if total < 0 {
		total = 0
	}
	if totalPages <= 0 {
		totalPages = 1
	}
	return PaginationScope{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// PaginationControlsWithScope creates pagination controls from normalized
// server-pagination fields and a scope-aware summary.
func PaginationControlsWithScope(key string, width int, scope PaginationScope, prevIntent, nextIntent intent.Intent, busy bool, loadingReason string, summaryWidth int, parts ...PageSummaryPart) rtui.VNode {
	return PaginationConfig{
		Key:           key,
		Width:         width,
		Summary:       PageSummary(scope.Page, scope.TotalPages, scope.Total, scope.PageSize, summaryWidth, parts...),
		PrevIntent:    prevIntent,
		NextIntent:    nextIntent,
		Busy:          busy,
		Page:          scope.Page,
		TotalPages:    scope.TotalPages,
		LoadingReason: loadingReason,
	}.Controls()
}

// PaginationPrev creates a Prev pagination action with common disabled reasons.
func PaginationPrev(pressIntent intent.Intent, busy bool, page int, loadingReason string) Item {
	item := Button("prev", "Prev", pressIntent)
	if busy {
		return item.WithDisabledReason(firstNonEmpty(strings.TrimSpace(loadingReason), "Data is loading."))
	}
	if page <= 1 {
		return item.WithDisabledReason("Already at the first page.")
	}
	return item
}

// PaginationNext creates a Next pagination action with common disabled reasons.
func PaginationNext(pressIntent intent.Intent, busy bool, page, totalPages int, loadingReason string) Item {
	item := Button("next", "Next", pressIntent)
	if busy {
		return item.WithDisabledReason(firstNonEmpty(strings.TrimSpace(loadingReason), "Data is loading."))
	}
	if page >= totalPages {
		return item.WithDisabledReason("Already at the last page.")
	}
	return item
}

// PaginationControls creates a standard pagination toolbar with summary and Prev/Next actions.
func PaginationControls(key string, width int, summary Item, prevIntent, nextIntent intent.Intent, busy bool, page, totalPages int, loadingReason string) rtui.VNode {
	return PaginationConfig{
		Key:           key,
		Width:         width,
		Summary:       summary,
		PrevIntent:    prevIntent,
		NextIntent:    nextIntent,
		Busy:          busy,
		Page:          page,
		TotalPages:    totalPages,
		LoadingReason: loadingReason,
	}.Controls()
}

// Prev creates a standard Prev pagination action for this pagination config.
func (c PaginationConfig) Prev() Item {
	return PaginationPrev(c.PrevIntent, c.Busy, c.Page, c.LoadingReason)
}

// Next creates a standard Next pagination action for this pagination config.
func (c PaginationConfig) Next() Item {
	return PaginationNext(c.NextIntent, c.Busy, c.Page, c.TotalPages, c.LoadingReason)
}

// Controls creates a standard pagination toolbar with summary and Prev/Next actions.
func (c PaginationConfig) Controls() rtui.VNode {
	width := c.Width
	if width <= 0 {
		width = 96
	}
	summary := c.Summary
	if summary.Key == "" && summary.Label == "" && summary.Kind == "" {
		summary = Text("page", "-").WithWidth(72)
	}
	return NewBuilder().
		Key(c.Key).
		Width(width).
		Dense(true).
		Left(summary).
		Right(c.Prev()).
		Right(c.Next()).
		Build()
}

// SelectionPrev creates an Up selection action with common disabled reasons.
func SelectionPrev(pressIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string) Item {
	item := Button("select-prev", "Up", pressIntent)
	if busy {
		return item.WithDisabledReason(firstNonEmpty(strings.TrimSpace(loadingReason), "Data is loading."))
	}
	label := selectionItemLabel(itemLabel)
	if total <= 0 {
		return item.WithDisabledReason("No " + label + " available.")
	}
	if index <= 0 {
		return item.WithDisabledReason("Already at the first " + label + ".")
	}
	return item
}

// SelectionNext creates a Down selection action with common disabled reasons.
func SelectionNext(pressIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string) Item {
	item := Button("select-next", "Down", pressIntent)
	if busy {
		return item.WithDisabledReason(firstNonEmpty(strings.TrimSpace(loadingReason), "Data is loading."))
	}
	label := selectionItemLabel(itemLabel)
	if total <= 0 {
		return item.WithDisabledReason("No " + label + " available.")
	}
	if index >= total-1 {
		return item.WithDisabledReason("Already at the last " + label + ".")
	}
	return item
}

// SelectionControls creates a standard two-button selection navigation toolbar.
func SelectionControls(key, title string, prevIntent, nextIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string) rtui.VNode {
	return SelectionConfig{
		Key:           key,
		Title:         title,
		PrevIntent:    prevIntent,
		NextIntent:    nextIntent,
		Busy:          busy,
		Index:         index,
		Total:         total,
		ItemLabel:     itemLabel,
		LoadingReason: loadingReason,
	}.Controls()
}

// Prev creates a standard Up selection action for this selection config.
func (c SelectionConfig) Prev() Item {
	return SelectionPrev(c.PrevIntent, c.Busy, c.Index, c.Total, c.ItemLabel, c.LoadingReason)
}

// Next creates a standard Down selection action for this selection config.
func (c SelectionConfig) Next() Item {
	return SelectionNext(c.NextIntent, c.Busy, c.Index, c.Total, c.ItemLabel, c.LoadingReason)
}

// Controls creates a standard two-button selection navigation toolbar.
func (c SelectionConfig) Controls() rtui.VNode {
	width := c.Width
	if width <= 0 {
		width = 54
	}
	return NewBuilder().
		Key(c.Key).
		Title(c.Title).
		TitleWidth(10).
		Width(width).
		Dense(true).
		Left(c.Prev()).
		Left(c.Next()).
		Build()
}

// SelectionActionControls creates a selection navigation toolbar with contextual actions.
func SelectionActionControls(key, title string, width int, prevIntent, nextIntent intent.Intent, busy bool, index, total int, itemLabel, loadingReason string, actions ...Item) rtui.VNode {
	return SelectionConfig{
		Key:           key,
		Title:         title,
		Width:         width,
		PrevIntent:    prevIntent,
		NextIntent:    nextIntent,
		Busy:          busy,
		Index:         index,
		Total:         total,
		ItemLabel:     itemLabel,
		LoadingReason: loadingReason,
	}.ActionControls(actions...)
}

// ActionControls creates a selection navigation toolbar with contextual actions.
func (c SelectionConfig) ActionControls(actions ...Item) rtui.VNode {
	items := []Item{
		c.Prev(),
		c.Next(),
	}
	items = append(items, actions...)
	width := c.Width
	if width <= 0 {
		width = 64
	}
	return NewBuilder().
		Key(c.Key).
		Title(c.Title).
		TitleWidth(10).
		Width(width).
		Dense(true).
		LeftItems(items).
		Build()
}

// ActionControls creates a compact action toolbar from caller-supplied items.
func ActionControls(key string, width int, items ...Item) rtui.VNode {
	if width <= 0 {
		width = 64
	}
	return NewBuilder().
		Key(key).
		Width(width).
		Dense(true).
		LeftItems(items).
		Build()
}

// ActionGroup creates a standard toolbar for a titled group of operation actions.
func ActionGroup(key, title string, items []Item) rtui.VNode {
	return ActionGroupWithLayout(key, title, 118, 16, items)
}

// ActionGroupWithLayout creates a titled operation toolbar with caller-provided
// width and title width. Non-positive values fall back to ActionGroup defaults.
func ActionGroupWithLayout(key, title string, width, titleWidth int, items []Item) rtui.VNode {
	if width <= 0 {
		width = 118
	}
	if titleWidth <= 0 {
		titleWidth = 16
	}
	return NewBuilder().
		Key(key).
		Title(title).
		TitleWidth(titleWidth).
		Width(width).
		Dense(true).
		LeftItems(items).
		Build()
}

// ActionGroups creates a grouped operation surface with optional summary text.
//
// Empty groups are skipped. If no groups have actions, emptyText is rendered so
// operators can distinguish "filtered out" from a broken operation panel.
func ActionGroups(key string, groups []ActionGroupConfig, summary, emptyText string) rtui.VNode {
	children := make([]rtui.VNode, 0, len(groups)+2)
	for _, group := range groups {
		if len(group.Items) == 0 {
			continue
		}
		children = append(children, ActionGroupWithLayout(group.Key, group.Title, group.Width, group.TitleWidth, group.Items))
	}
	if len(children) == 0 {
		message := strings.TrimSpace(normalizeToolbarText(emptyText))
		if message == "" {
			message = "No actions available."
		}
		children = append(children, actionSummaryText(key+"-empty", message))
	}
	if summary = strings.TrimSpace(normalizeToolbarText(summary)); summary != "" {
		children = append(children, actionSummaryText(key+"-summary", summary))
	}
	root := rtui.VStackBuilder(children...).Gap(1).AlignCross(rtui.AlignStart).Build()
	root.SetKey(key)
	return root
}

func actionSummaryText(key, content string) rtui.VNode {
	return textcomp.NewBuilder(content).Key(key).FgColor("gray").Build()
}

func selectionItemLabel(label string) string {
	label = strings.TrimSpace(normalizeToolbarText(label))
	if label == "" {
		return "item"
	}
	return label
}

func toneColors(tone statusbarcomp.Tone) (fgColor, bgColor string) {
	switch tone {
	case statusbarcomp.ToneNormal:
		return "black", "green"
	case statusbarcomp.ToneWarn:
		return "black", "yellow"
	case statusbarcomp.ToneError:
		return "white", "red"
	case statusbarcomp.ToneInfo:
		return "black", "cyan"
	default:
		return "bright-white", "bright-black"
	}
}

func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

func (i Item) WithLabel(label string) Item {
	i.Label = label
	return i
}

func (i Item) OnPress(pressIntent intent.Intent) Item {
	i.PressIntent = pressIntent
	if i.Kind == ItemText || i.Kind == ItemBadge {
		i.Kind = ItemButton
	}
	return i
}

func (i Item) WithVariant(variant button.Variant) Item {
	i.Variant = variant
	return i
}

func (i Item) Primary() Item {
	i.Variant = button.VariantPrimary
	return i
}

func (i Item) Secondary() Item {
	i.Variant = button.VariantSecondary
	return i
}

func (i Item) Danger() Item {
	i.Variant = button.VariantDanger
	return i
}

func (i Item) Success() Item {
	i.Variant = button.VariantSuccess
	return i
}

func (i Item) WithDisabled(disabled bool) Item {
	i.Disabled = disabled
	return i
}

func (i Item) WithDisabledReason(reason string) Item {
	i.Disabled = true
	i.DisabledReason = normalizeToolbarText(reason)
	return i
}

// WithDisabledReasonIf disables the item with reason when condition is true.
// When the item already has a disabled reason, the new reason is appended so
// chained condition checks expose all matching prerequisites.
func (i Item) WithDisabledReasonIf(condition bool, reason string) Item {
	if !condition {
		return i
	}
	if strings.TrimSpace(normalizeToolbarText(i.DisabledReason)) != "" {
		return i.WithDisabledReason(JoinDisabledReasons(i.DisabledReason, reason))
	}
	return i.WithDisabledReason(reason)
}

// WithDisabledReasons disables the item when at least one non-empty reason is
// supplied and joins all reasons into one concise explanation. It is intended
// for operation buttons with multiple missing prerequisites, such as required
// reason plus selected target.
func (i Item) WithDisabledReasons(reasons ...string) Item {
	joined := JoinDisabledReasons(reasons...)
	if joined == "" {
		return i
	}
	return i.WithDisabledReason(joined)
}

// JoinDisabledReasons normalizes, filters, and joins multiple disabled reasons.
func JoinDisabledReasons(reasons ...string) string {
	parts := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		reason = strings.TrimSpace(normalizeToolbarText(reason))
		if reason == "" {
			continue
		}
		parts = append(parts, reason)
	}
	return strings.Join(parts, " ")
}

func (i Item) WithHelp(helpText string) Item {
	i.HelpText = normalizeToolbarText(helpText)
	return i
}

func (i Item) WithTooltip(tooltipText string) Item {
	return i.WithHelp(tooltipText)
}

func (i Item) WithWidth(width int) Item {
	i.Width = width
	return i
}

func (i Item) WithColors(fgColor, bgColor string) Item {
	i.FgColor = fgColor
	i.BgColor = bgColor
	return i
}

func (i Item) WithForeground(fgColor string) Item {
	i.FgColor = fgColor
	return i
}

func (i Item) WithBackground(bgColor string) Item {
	i.BgColor = bgColor
	return i
}

func (i Item) WithBold(bold bool) Item {
	i.Bold = bold
	return i
}

func (i Item) WithMenuID(menuID string) Item {
	i.MenuID = strings.TrimSpace(menuID)
	return i
}

func (i Item) WithMenuItems(items []menucomp.MenuItem) Item {
	i.MenuItems = menucomp.NormalizeItems(items)
	if i.Kind == "" || i.Kind == ItemText || i.Kind == ItemButton {
		i.Kind = ItemMenu
	}
	return i
}

func (i Item) WithMenuOpen(open bool) Item {
	i.MenuOpen = open
	return i
}

func (i Item) WithMenuPlacement(placement menucomp.Placement) Item {
	i.MenuPlacement = placement
	return i
}

func (i Item) WithMenuActivePath(path ...int) Item {
	i.MenuActivePath = append([]int(nil), path...)
	return i
}

func (i Item) WithMenuMinWidth(width int) Item {
	i.MenuMinWidth = width
	return i
}

func (i Item) WithMenuMaxHeight(height int) Item {
	i.MenuMaxHeight = height
	return i
}

func (i Item) WithMenuShortcuts(show bool) Item {
	i.MenuShowShortcuts = show
	return i
}

func (i Item) WithMenuDescriptions(show bool) Item {
	i.MenuShowDescriptions = show
	return i
}

func normalizeItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	normalized := cloneItems(items)
	seen := make(map[string]int, len(normalized))
	for index := range normalized {
		key := strings.TrimSpace(normalized[index].Key)
		if key == "" {
			key = fmt.Sprintf("item-%d", index)
		}
		base := key
		if count, exists := seen[base]; exists {
			count++
			seen[base] = count
			key = fmt.Sprintf("%s-%d", base, count)
		} else {
			seen[base] = 0
		}
		normalized[index].Key = key
		if normalized[index].Kind == "" {
			normalized[index].Kind = ItemText
		}
		switch normalized[index].Kind {
		case ItemText, ItemBadge, ItemButton, ItemMenu, ItemSeparator, ItemCustom:
		default:
			normalized[index].Kind = ItemText
		}
		if normalized[index].Width < 0 {
			normalized[index].Width = 0
		}
		normalized[index].Label = normalizeToolbarText(normalized[index].Label)
		normalized[index].HelpText = normalizeToolbarText(normalized[index].HelpText)
		normalized[index].DisabledReason = normalizeToolbarText(normalized[index].DisabledReason)
		normalized[index].MenuID = strings.TrimSpace(normalized[index].MenuID)
		normalized[index].MenuItems = menucomp.NormalizeItems(normalized[index].MenuItems)
		normalized[index].MenuActivePath = append([]int(nil), normalized[index].MenuActivePath...)
		if normalized[index].MenuPlacement == "" {
			normalized[index].MenuPlacement = menucomp.PlacementBottomStart
		}
		if normalized[index].MenuMinWidth < 0 {
			normalized[index].MenuMinWidth = 0
		}
		if normalized[index].MenuMaxHeight < 0 {
			normalized[index].MenuMaxHeight = 0
		}
	}
	return normalized
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}

func normalizeToolbarText(content string) string {
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(content)
}

func menuAnchorForPlacement(placement menucomp.Placement) rttypes.Anchor {
	switch placement {
	case menucomp.PlacementTopStart, menucomp.PlacementTopEnd:
		return rttypes.AnchorTopLeft
	case menucomp.PlacementRightStart:
		return rttypes.AnchorTopRight
	case menucomp.PlacementLeftStart:
		return rttypes.AnchorTopLeft
	default:
		return rttypes.AnchorBottomLeft
	}
}

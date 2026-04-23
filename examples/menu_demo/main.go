package main

import (
	"fmt"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	"github.com/wwsheng009/mint/ui"
)

const (
	menuComponentID = "menu-demo"
	menuBarID       = "menu-demo-bar"
)

type AppState struct {
	MenuOpen     bool
	OpenRoot     int
	ReadOnly     bool
	Theme        string
	LastAction   string
	ShortcutInfo string
}

type NewDocumentIntent struct{}
type OpenDocumentIntent struct{}
type SaveDocumentIntent struct{}
type ToggleReadonlyIntent struct{}
type RefreshPreviewIntent struct{}
type ShowAboutIntent struct{}

type SetThemeIntent struct {
	Name string
}

func (NewDocumentIntent) IntentType() string    { return "MenuDemoNewDocument" }
func (OpenDocumentIntent) IntentType() string   { return "MenuDemoOpenDocument" }
func (SaveDocumentIntent) IntentType() string   { return "MenuDemoSaveDocument" }
func (ToggleReadonlyIntent) IntentType() string { return "MenuDemoToggleReadonly" }
func (RefreshPreviewIntent) IntentType() string { return "MenuDemoRefreshPreview" }
func (ShowAboutIntent) IntentType() string      { return "MenuDemoShowAbout" }
func (i SetThemeIntent) IntentType() string     { return "MenuDemoSetTheme" }

func (NewDocumentIntent) StayPressed() bool    { return true }
func (OpenDocumentIntent) StayPressed() bool   { return true }
func (SaveDocumentIntent) StayPressed() bool   { return true }
func (ToggleReadonlyIntent) StayPressed() bool { return true }
func (RefreshPreviewIntent) StayPressed() bool { return true }
func (ShowAboutIntent) StayPressed() bool      { return true }
func (SetThemeIntent) StayPressed() bool       { return true }

var appStore = store.NewStore(AppState{
	MenuOpen:     false,
	OpenRoot:     0,
	ReadOnly:     false,
	Theme:        "dark",
	LastAction:   "Ready",
	ShortcutInfo: "Ctrl+N/O/S/R, F1, F5",
})

func init() {
	reducer.NewBuilder[AppState]().
		On(ui.OpenMenuIntent{}, func(s AppState, i intent.Intent) AppState {
			openIntent, ok := i.(ui.OpenMenuIntent)
			if !ok {
				return s
			}
			if len(openIntent.Path) > 0 {
				s.OpenRoot = openIntent.Path[0]
			}
			s.MenuOpen = true
			return s
		}).
		On(ui.CloseMenuIntent{}, func(s AppState, i intent.Intent) AppState {
			s.MenuOpen = false
			return s
		}).
		On(ui.ActivateMenuItemIntent{}, func(s AppState, i intent.Intent) AppState {
			act, ok := i.(ui.ActivateMenuItemIntent)
			if !ok {
				return s
			}
			s.LastAction = fmt.Sprintf("activate %s %v", act.ItemKey, act.Path)
			return s
		}).
		On(NewDocumentIntent{}, func(s AppState, i intent.Intent) AppState {
			s.LastAction = "Created a new document"
			return s
		}).
		On(OpenDocumentIntent{}, func(s AppState, i intent.Intent) AppState {
			s.LastAction = "Opened an existing document"
			return s
		}).
		On(SaveDocumentIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.ReadOnly {
				s.LastAction = "Save is disabled while Read Only is enabled"
				return s
			}
			s.LastAction = "Saved document"
			return s
		}).
		On(ToggleReadonlyIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ReadOnly = !s.ReadOnly
			s.LastAction = fmt.Sprintf("Read Only = %t", s.ReadOnly)
			return s
		}).
		On(RefreshPreviewIntent{}, func(s AppState, i intent.Intent) AppState {
			s.LastAction = "Refreshed preview"
			return s
		}).
		On(ShowAboutIntent{}, func(s AppState, i intent.Intent) AppState {
			s.LastAction = "Mint Menu Demo: standalone menu example"
			return s
		}).
		On(SetThemeIntent{}, func(s AppState, i intent.Intent) AppState {
			setTheme, ok := i.(SetThemeIntent)
			if !ok {
				return s
			}
			s.Theme = setTheme.Name
			s.LastAction = fmt.Sprintf("Theme = %s", s.Theme)
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

func main() {
	err := ui.Run(App,
		ui.WithWidth(88),
		ui.WithHeight(28),
		ui.WithTitle("Mint Menu Demo"),
		ui.WithPluginSetup(func(app *framework.App) {
			ui.InstallMenu(app,
				ui.NewMenuBarBuilder(buildMenuItems(appStore.Get())).
					ComponentID(menuComponentID).
					RegisterShortcuts(true),
			)
		}),
	)
	if err != nil {
		panic(err)
	}
}

func App() ui.VNode {
	state := appStore.Get()
	items := buildMenuItems(state)

	contentNodes := []ui.VNode{
		ui.NewMenuBarBuilder(items).
			SetID(menuBarID).
			ComponentID(menuComponentID).
			Theme(ui.MenuThemeDefault()).
			Build(),
		ui.Text(""),
		ui.TextBold("Menu Demo"),
		ui.Text("Tab focus menu bar, ←/→ switch root menus, ↓ or Enter opens popup."),
		ui.Text("Popup supports hover, Enter, typeahead, outside-click close, and global shortcuts."),
		ui.Text(""),
		ui.Text(fmt.Sprintf("Theme: %s", state.Theme)),
		ui.Text(fmt.Sprintf("Read Only: %t", state.ReadOnly)),
		ui.Text(fmt.Sprintf("Last Action: %s", state.LastAction)),
		ui.Text(fmt.Sprintf("Shortcuts: %s", state.ShortcutInfo)),
	}

	if state.MenuOpen && state.OpenRoot >= 0 && state.OpenRoot < len(items) {
		root := items[state.OpenRoot]
		contentNodes = append(contentNodes,
			ui.NewMenuPopupBuilder(root.Children).
				SetID("menu-demo-popup").
				ComponentID(menuComponentID).
				PathPrefix(state.OpenRoot).
				Title(root.Label).
				Theme(ui.MenuThemeDefault()).
				Open(true).
				MinWidth(34).
				MaxHeight(12).
				ShowDescriptions(true).
				ShowIcons(true).
				ShowCheckMarks(true).
				CloseOnOutside(true).
				CloseOnEscape(true).
				Typeahead(true).
				AnchorTo(menuBarID, rttypes.AnchorTopLeft).
				PortalPosition(rttypes.PositionAbsolute).
				PortalOffset(rootMenuOffset(items, state.OpenRoot), 1).
				Build(),
		)
	}

	return ui.VStack(contentNodes...)
}

func buildMenuItems(state AppState) []ui.MenuItem {
	return ui.MenuItems(
		ui.MenuSubmenu("file", "File",
			ui.MenuAction("new", "New Document", NewDocumentIntent{}).
				WithIcon("📄").
				WithDescription("Create a blank document").
				WithShortcut("ctrl+n"),
			ui.MenuAction("open", "Open...", OpenDocumentIntent{}).
				WithIcon("📂").
				WithDescription("Load an existing document").
				WithShortcut("ctrl+o"),
			ui.MenuAction("save", "Save", SaveDocumentIntent{}).
				WithIcon("💾").
				WithDescription("Write changes to disk").
				WithShortcut("ctrl+s").
				WithDisabled(state.ReadOnly),
			ui.MenuSeparator(),
			ui.MenuCheckbox("readonly", "Read Only", state.ReadOnly, ToggleReadonlyIntent{}).
				WithDescription("Prevent save operations").
				WithShortcut("ctrl+r"),
		),
		ui.MenuSubmenu("view", "View",
			ui.MenuSubmenu("theme", "Theme",
				ui.MenuRadio("theme-dark", "Dark", "theme", state.Theme == "dark", SetThemeIntent{Name: "dark"}),
				ui.MenuRadio("theme-light", "Light", "theme", state.Theme == "light", SetThemeIntent{Name: "light"}),
				ui.MenuRadio("theme-system", "System", "theme", state.Theme == "system", SetThemeIntent{Name: "system"}),
			).WithDescription("Nested submenu example"),
			ui.MenuAction("refresh", "Refresh Preview", RefreshPreviewIntent{}).
				WithIcon("↻").
				WithDescription("Re-render the current preview").
				WithShortcut("f5"),
		),
		ui.MenuSubmenu("help", "Help",
			ui.MenuAction("about", "About", ShowAboutIntent{}).
				WithIcon("ⓘ").
				WithDescription("Show demo metadata").
				WithShortcut("f1"),
		),
	)
}

func rootMenuOffset(items []ui.MenuItem, rootIndex int) int {
	cursor := 0
	for index, item := range items {
		if item.Hidden {
			continue
		}
		if index == rootIndex {
			return cursor
		}
		cursor += paint.StringWidth(" "+item.Label+" ") + 1
	}
	return 0
}

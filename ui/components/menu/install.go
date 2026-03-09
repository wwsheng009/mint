package menu

import (
	"reflect"
	"sync"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
)

type InstallerHost interface {
	ShortcutRegistrar
	AddMiddleware(action.ActionMiddleware)
}

type installState struct {
	mu                  sync.Mutex
	middlewareInstalled bool
	shortcutCombos      map[string]bool
}

var installStateRegistry sync.Map

func Install(host InstallerHost, emit func(intent.Intent), builders ...*Builder) int {
	if host == nil {
		return 0
	}

	state := getInstallState(host)
	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.middlewareInstalled {
		host.AddMiddleware(NewMiddleware())
		state.middlewareInstalled = true
	}

	if emit == nil {
		return 0
	}

	registered := 0
	for _, builder := range builders {
		if builder == nil || !builder.model.RegisterShortcuts {
			continue
		}
		model := builder.BuildModel()
		menuID := firstNonEmpty(model.ComponentID, model.ID, "menu")
		for _, binding := range CollectShortcuts(model.Items) {
			if binding.Shortcut.Combo == "" || binding.Shortcut.Scope == ShortcutLocal {
				continue
			}
			combo := normalizeCombo(binding.Shortcut.Combo)
			if combo == "" || state.shortcutCombos[combo] {
				continue
			}
			state.shortcutCombos[combo] = true
			b := binding
			host.OnKeyCombo(b.Shortcut.Combo, func() {
				emit(ActivateMenuItemIntent{MenuID: menuID, ItemKey: b.Item.Key, Path: append([]int(nil), b.Path...)})
				if b.Item.HasSubmenu() {
					emit(OpenMenuIntent{MenuID: menuID, Path: append([]int(nil), b.Path...)})
					return
				}
				if itemIntent := b.Item.EffectiveIntent(); itemIntent != nil {
					emit(itemIntent)
				}
			})
			registered++
		}
	}

	return registered
}

func getInstallState(host InstallerHost) *installState {
	key := installerHostKey(host)
	if key == 0 {
		return &installState{shortcutCombos: map[string]bool{}}
	}
	if existing, ok := installStateRegistry.Load(key); ok {
		return existing.(*installState)
	}
	state := &installState{shortcutCombos: map[string]bool{}}
	actual, _ := installStateRegistry.LoadOrStore(key, state)
	return actual.(*installState)
}

func installerHostKey(host InstallerHost) uintptr {
	value := reflect.ValueOf(host)
	if !value.IsValid() {
		return 0
	}
	if value.Kind() == reflect.Pointer || value.Kind() == reflect.UnsafePointer {
		return value.Pointer()
	}
	return 0
}

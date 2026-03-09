package menu

import "github.com/wwsheng009/mint/runtime/intent"

type ShortcutRegistrar interface {
	OnKeyCombo(combo string, handler func())
}

func RegisterGlobalShortcuts(registrar ShortcutRegistrar, menuID string, items []MenuItem, emit func(intent.Intent)) int {
	if registrar == nil || emit == nil {
		return 0
	}
	bindings := CollectShortcuts(items)
	registered := 0
	for _, binding := range bindings {
		if binding.Shortcut.Combo == "" {
			continue
		}
		if binding.Shortcut.Scope == ShortcutLocal {
			continue
		}
		b := binding
		registrar.OnKeyCombo(b.Shortcut.Combo, func() {
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
	return registered
}

func (b *Builder) BindGlobalShortcuts(registrar ShortcutRegistrar, emit func(intent.Intent)) int {
	if b == nil || !b.model.RegisterShortcuts {
		return 0
	}
	model := b.BuildModel()
	return RegisterGlobalShortcuts(registrar, firstNonEmpty(model.ComponentID, model.ID, "menu"), model.Items, emit)
}

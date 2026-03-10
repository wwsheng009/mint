package capability

import (
	"github.com/wwsheng009/mint/runtime/ui"
)

// SetValue attempts to assign a component value through stable instance APIs.
func SetValue(fiber *ui.Fiber, value interface{}) bool {
	if fiber == nil || fiber.Instance == nil {
		return false
	}

	if setter, ok := fiber.Instance.(interface{ SetValue(string) }); ok {
		switch v := value.(type) {
		case string:
			setter.SetValue(v)
			return true
		}
	}

	if setter, ok := fiber.Instance.(interface {
		SetValue(field string, value interface{})
	}); ok {
		// Form-like instances require a field key, so a plain SetValue(locator, value)
		// should not route here.
		_ = setter
	}

	if setter, ok := fiber.Instance.(interface {
		SetProp(key string, value interface{})
	}); ok {
		setter.SetProp("value", value)
		return true
	}

	return false
}

// SetProp attempts to assign a component prop through stable instance APIs.
func SetProp(fiber *ui.Fiber, key string, value interface{}) bool {
	if fiber == nil || fiber.Instance == nil {
		return false
	}
	if setter, ok := fiber.Instance.(interface {
		SetProp(key string, value interface{})
	}); ok {
		setter.SetProp(key, value)
		return true
	}
	return false
}

func SelectIndex(fiber *ui.Fiber, index int) bool {
	if fiber == nil || fiber.Instance == nil {
		return false
	}
	if selector, ok := fiber.Instance.(interface{ SelectIndex(int) bool }); ok {
		return selector.SelectIndex(index)
	}
	if selector, ok := fiber.Instance.(interface{ SelectIndexForAI(int) bool }); ok {
		return selector.SelectIndexForAI(index)
	}
	return false
}

func ToggleSelectionIndex(fiber *ui.Fiber, index int) bool {
	if fiber == nil || fiber.Instance == nil {
		return false
	}
	if toggler, ok := fiber.Instance.(interface{ ToggleSelectionAt(int) bool }); ok {
		return toggler.ToggleSelectionAt(index)
	}
	if toggler, ok := fiber.Instance.(interface{ ToggleSelectionAtSourceIndex(int) bool }); ok {
		return toggler.ToggleSelectionAtSourceIndex(index)
	}
	return false
}

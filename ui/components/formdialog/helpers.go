package formdialog

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/form"
)

func normalizeLayout(layout form.FormLayout) form.FormLayout {
	switch layout {
	case form.LayoutHorizontal, form.LayoutInline:
		return layout
	default:
		return form.LayoutVertical
	}
}

func cloneChildren(children []rtui.VNode) []rtui.VNode {
	if len(children) == 0 {
		return nil
	}
	out := make([]rtui.VNode, len(children))
	copy(out, children)
	return out
}

func cloneValues(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func getChildrenProp(props rtui.Props, key string, def []rtui.VNode) []rtui.VNode {
	if children, ok := props[key].([]rtui.VNode); ok {
		return cloneChildren(children)
	}
	return cloneChildren(def)
}

func getValuesProp(props rtui.Props, key string, def map[string]interface{}) map[string]interface{} {
	if values, ok := props[key].(map[string]interface{}); ok {
		return cloneValues(values)
	}
	return cloneValues(def)
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if value, ok := props[key]; ok {
		if number, ok := value.(int); ok {
			return number
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key]; ok {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	return def
}

func getLayoutProp(props rtui.Props, key string, def form.FormLayout) form.FormLayout {
	if value, ok := props[key]; ok {
		switch layout := value.(type) {
		case form.FormLayout:
			return normalizeLayout(layout)
		case string:
			return normalizeLayout(form.FormLayout(layout))
		}
	}
	return normalizeLayout(def)
}

func getStyleProp(props rtui.Props, key string, def style.Style) style.Style {
	if value, ok := props[key]; ok {
		if s, ok := value.(style.Style); ok {
			return s
		}
	}
	return def
}

func getIntentProp(props rtui.Props, key string, def intent.Intent) intent.Intent {
	if value, ok := props[key]; ok {
		if i, ok := value.(intent.Intent); ok {
			return i
		}
	}
	return def
}

func getButtonVariantProp(props rtui.Props, key string, def button.Variant) button.Variant {
	if value, ok := props[key]; ok {
		if variant, ok := value.(button.Variant); ok {
			return variant
		}
	}
	return def
}

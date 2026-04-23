package spin

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func getSizeProp(props rtui.Props, def Size) Size {
	if v, ok := props[propSize]; ok {
		if s, ok := v.(Size); ok {
			return s
		}
	}
	return def
}

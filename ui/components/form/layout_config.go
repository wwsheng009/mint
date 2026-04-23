package form

// FormLayout controls how FormItem arranges label, field and error text.
type FormLayout string

const (
	LayoutVertical   FormLayout = "vertical"
	LayoutHorizontal FormLayout = "horizontal"
	LayoutInline     FormLayout = "inline"
)

func normalizeLayout(layout FormLayout) FormLayout {
	switch layout {
	case LayoutHorizontal:
		return LayoutHorizontal
	case LayoutInline:
		return LayoutInline
	default:
		return LayoutVertical
	}
}

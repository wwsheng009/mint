package layout

// ==============================================================================
// Wrap Layout (flex-wrap: wrap)
// ==============================================================================
// Wrap 布局容器，子元素在达到容器宽度时自动换行

// WrapAlignment 定义行内元素的对齐方式
type WrapAlignment int

const (
	// WrapAlignStart 行内元素左对齐
	WrapAlignStart WrapAlignment = iota
	// WrapAlignCenter 行内元素居中对齐
	WrapAlignCenter
	// WrapAlignEnd 行内元素右对齐
	WrapAlignEnd
)

// WrapStyle 定义 Wrap 布局容器的样式
type WrapStyle struct {
	// Width 容器宽度（决定何时换行）
	Width int

	// Gap 行内元素之间的间距
	Gap int

	// RowGap 行与行之间的间距（0 表示使用 Gap）
	RowGap int

	// Align 行内元素的对齐方式
	Align WrapAlignment

	// Padding 内边距
	Padding Padding

	// FillWidth 是否拉伸每行填满容器宽度
	FillWidth bool

	// FillHeight 是否拉伸容器填满父容器高度
	FillHeight bool
}

// WrapStyleProvider 定义获取 Wrap 布局样式的接口
// 实现此接口的节点将被 Engine 使用 WrapLayout 进行布局
type WrapStyleProvider interface {
	GetWrapStyle() *WrapStyle
}

// DefaultWrapStyle 默认 Wrap 样式
func DefaultWrapStyle() *WrapStyle {
	return &WrapStyle{
		Width:     80,
		Gap:       1,
		RowGap:    0,
		Align:     WrapAlignStart,
		Padding:   Padding{},
		FillWidth: false,
	}
}

// WrapLayout Wrap 布局节点
type WrapLayout struct {
	id       string
	children []Node
	style    *WrapStyle

	// 计算结果
	rows       [][]int // 每行包含的子节点索引
	rowHeights []int   // 每行的高度
	rowWidths  []int   // 每行的宽度
}

// NewWrapLayout 创建 Wrap 布局
func NewWrapLayout(id string, style *WrapStyle) *WrapLayout {
	if style == nil {
		style = DefaultWrapStyle()
	}
	return &WrapLayout{
		id:     id,
		style:  style,
		rows:   nil,
		rowHeights: nil,
		rowWidths:  nil,
	}
}

// SetChildren 设置子节点
func (w *WrapLayout) SetChildren(children []Node) {
	w.children = children
}

// ID 返回节点 ID
func (w *WrapLayout) ID() string {
	return w.id
}

// LayoutChildren 布局子节点，返回每个子节点的 LayoutBox
func (w *WrapLayout) LayoutChildren(containerWidth, containerHeight int) []*LayoutBox {
	if len(w.children) == 0 {
		return nil
	}

	// 使用样式中定义的宽度，如果未设置则使用容器宽度
	width := w.style.Width
	if width <= 0 {
		width = containerWidth
	}

	// 计算内容可用宽度（减去内边距）
	contentWidth := width - w.style.Padding.Left - w.style.Padding.Right
	if contentWidth < 0 {
		contentWidth = 0
	}

	// 行间距（如果 RowGap 为 0，使用 Gap）
	rowGap := w.style.RowGap
	if rowGap == 0 {
		rowGap = w.style.Gap
	}

	// 第一遍：测量所有子节点尺寸
	childSizes := make([]Size, len(w.children))
	for i, child := range w.children {
		childSizes[i] = w.measureChild(child, contentWidth)
	}

	// 第二遍：计算行划分
	w.calculateRows(childSizes, contentWidth)

	// 第三遍：计算每个子节点的位置
	boxes := make([]*LayoutBox, len(w.children))

	y := w.style.Padding.Top

	for rowIdx, row := range w.rows {
		rowHeight := w.rowHeights[rowIdx]
		rowWidth := w.rowWidths[rowIdx]

		// 计算行的起始 X 位置（根据对齐方式）
		x := w.style.Padding.Left
		remainingSpace := contentWidth - rowWidth

		switch w.style.Align {
		case WrapAlignCenter:
			x += remainingSpace / 2
		case WrapAlignEnd:
			x += remainingSpace
		}

		// 放置行内元素
		for _, childIdx := range row {
			childSize := childSizes[childIdx]
			child := w.children[childIdx]

			// 计算子节点 Y 位置（行内垂直对齐，目前是顶部对齐）
			childY := y

			boxes[childIdx] = &LayoutBox{
				ID:     child.ID(),
				X:      x,
				Y:      childY,
				Width:  childSize.Width,
				Height: rowHeight,
			}

			x += childSize.Width + w.style.Gap
		}

		y += rowHeight
		if rowIdx < len(w.rows)-1 {
			y += rowGap
		}
	}

	return boxes
}

// measureChild 测量子节点尺寸
func (w *WrapLayout) measureChild(child Node, maxWidth int) Size {
	// 先尝试从节点获取尺寸
	cw, ch := child.GetSize()
	if cw > 0 && ch > 0 {
		return Size{Width: cw, Height: ch}
	}

	// 尝试使用 Measurable 接口
	if measurable, ok := child.(Measurable); ok {
		constraints := Constraints{
			MinWidth:  0,
			MaxWidth:  maxWidth,
			MinHeight: 0,
			MaxHeight: MaxInt,
		}
		return measurable.Measure(constraints)
	}

	// 默认尺寸
	if cw <= 0 {
		cw = 10
	}
	if ch <= 0 {
		ch = 1
	}

	return Size{Width: cw, Height: ch}
}

// calculateRows 计算行划分
func (w *WrapLayout) calculateRows(childSizes []Size, availableWidth int) {
	w.rows = nil
	w.rowHeights = nil
	w.rowWidths = nil

	if len(childSizes) == 0 {
		return
	}

	var currentRow []int
	currentWidth := 0
	currentHeight := 0

	for i, size := range childSizes {
		// 检查是否需要换行
		needWrap := false
		if len(currentRow) > 0 {
			// 当前行已有元素，检查添加新元素后是否超出
			newWidth := currentWidth + w.style.Gap + size.Width
			if newWidth > availableWidth {
				needWrap = true
			}
		}

		if needWrap {
			// 保存当前行
			w.rows = append(w.rows, currentRow)
			w.rowHeights = append(w.rowHeights, currentHeight)
			w.rowWidths = append(w.rowWidths, currentWidth)

			// 开始新行
			currentRow = []int{i}
			currentWidth = size.Width
			currentHeight = size.Height
		} else {
			// 添加到当前行
			if len(currentRow) > 0 {
				currentWidth += w.style.Gap
			}
			currentRow = append(currentRow, i)
			currentWidth += size.Width
			if size.Height > currentHeight {
				currentHeight = size.Height
			}
		}
	}

	// 保存最后一行
	if len(currentRow) > 0 {
		w.rows = append(w.rows, currentRow)
		w.rowHeights = append(w.rowHeights, currentHeight)
		w.rowWidths = append(w.rowWidths, currentWidth)
	}
}

// GetRows 返回计算后的行划分（用于调试）
func (w *WrapLayout) GetRows() [][]int {
	return w.rows
}

// GetRowHeights 返回每行高度（用于调试）
func (w *WrapLayout) GetRowHeights() []int {
	return w.rowHeights
}

// GetTotalHeight 计算总高度
func (w *WrapLayout) GetTotalHeight() int {
	if len(w.rowHeights) == 0 {
		return w.style.Padding.Top + w.style.Padding.Bottom
	}

	rowGap := w.style.RowGap
	if rowGap == 0 {
		rowGap = w.style.Gap
	}

	total := w.style.Padding.Top + w.style.Padding.Bottom
	for i, h := range w.rowHeights {
		total += h
		if i < len(w.rowHeights)-1 {
			total += rowGap
		}
	}

	return total
}

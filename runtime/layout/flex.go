package layout

// ==============================================================================
// Flexbox Layout Algorithm (V3)
// ==============================================================================
// 简化的 Flexbox 布局算法实现

// FlexDirection 弹性方向
type FlexDirection int

const (
	FlexRow FlexDirection = iota
	FlexColumn
	FlexRowReverse
	FlexColumnReverse
)

// MainAxisAlignment 主轴对齐
type MainAxisAlignment int

const (
	// MainStart 主轴起点对齐
	MainStart MainAxisAlignment = iota
	// MainEnd 主轴终点对齐
	MainEnd
	// Center 主轴居中对齐
	Center
	// SpaceBetween 两端对齐，间距平均分配
	SpaceBetween
	// SpaceAround 每个子元素两侧间距相等
	SpaceAround
	// SpaceEvenly 所有间距相等
	SpaceEvenly
)

// CrossAxisAlignment 交叉轴对齐
type CrossAxisAlignment int

const (
	// CrossStart 交叉轴起点对齐
	CrossStart CrossAxisAlignment = iota
	// CrossEnd 交叉轴终点对齐
	CrossEnd
	// CrossCenter 交叉轴居中对齐
	CrossCenter
	// Stretch 拉伸填满交叉轴
	Stretch
	// Baseline 基线对齐（仅用于 Row 布局）
	Baseline
)

// String returns the string representation
func (a CrossAxisAlignment) String() string {
	switch a {
	case CrossStart:
		return "CrossStart"
	case CrossEnd:
		return "CrossEnd"
	case CrossCenter:
		return "CrossCenter"
	case Stretch:
		return "Stretch"
	case Baseline:
		return "Baseline"
	default:
		return "Unknown"
	}
}

// FlexStyle 弹性布局样式
type FlexStyle struct {
	// Direction 弹性方向
	Direction FlexDirection

	// MainAxis 主轴对齐
	MainAxis MainAxisAlignment

	// CrossAxis 交叉轴对齐
	CrossAxis CrossAxisAlignment

	// Gap 主轴间距
	Gap int

	// CrossGap 交叉轴间距
	CrossGap int

	// Padding 内边距
	Padding Padding

	// Margin 外边距
	Margin Margin

	// FlexibleChildren 可伸缩子节点索引和配置
	FlexibleChildren map[int]*Flex
}

// Padding 内边距
type Padding struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Margin 外边距
// 外边距定义了节点与其兄弟节点或父容器之间的间距
type Margin struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Horizontal 返回水平方向总外边距
func (m Margin) Horizontal() int {
	return m.Left + m.Right
}

// Vertical 返回垂直方向总外边距
func (m Margin) Vertical() int {
	return m.Top + m.Bottom
}

// Horizontal 返回水平方向总内边距
func (p Padding) Horizontal() int {
	return p.Left + p.Right
}

// Vertical 返回垂直方向总内边距
func (p Padding) Vertical() int {
	return p.Top + p.Bottom
}

// Flex 弹性配置
type Flex struct {
	// Grow 放大比例（默认0，不放大）
	Grow int

	// Shrink 缩小比例（默认1，可以缩小）
	Shrink int

	// Basis 基础尺寸（默认auto，由内容决定）
	Basis int
}

// FlexStyleProvider 定义获取弹性布局样式的接口
// 实现 this 接口的节点将被 Engine 使用 FlexLayout 进行布局
type FlexStyleProvider interface {
	GetFlexStyle() *FlexStyle
}

// FlexChildProvider 定义获取子节点 flex 属性的接口
// 子节点可以实现此接口以告诉父容器它的 flex 值
type FlexChildProvider interface {
	GetFlex() int
}

// DefaultFlexStyle 默认弹性样式
func DefaultFlexStyle() *FlexStyle {
	return &FlexStyle{
		Direction:        FlexColumn,
		MainAxis:         MainStart,
		CrossAxis:        CrossStart,
		Gap:              0,
		CrossGap:         0,
		Padding:          Padding{},
		FlexibleChildren: make(map[int]*Flex),
	}
}

// FlexLayout 弹性布局节点
type FlexLayout struct {
	id       string
	children []Node
	style    *FlexStyle
	size     Size
	position Point
}

// NewFlexLayout 创建弹性布局
func NewFlexLayout(id string, children []Node) *FlexLayout {
	return &FlexLayout{
		id:       id,
		children: children,
		style:    DefaultFlexStyle(),
	}
}

// ID 返回节点ID
func (f *FlexLayout) ID() string {
	return f.id
}

// Type 返回节点类型
func (f *FlexLayout) Type() string {
	return "flex"
}

// Children 返回子节点
func (f *FlexLayout) Children() []Node {
	return f.children
}

// GetPosition 获取位置
func (f *FlexLayout) GetPosition() (int, int) {
	return f.position.X, f.position.Y
}

// SetPosition 设置位置
func (f *FlexLayout) SetPosition(x, y int) {
	f.position.X = x
	f.position.Y = y
}

// GetSize 获取尺寸
func (f *FlexLayout) GetSize() (int, int) {
	return f.size.Width, f.size.Height
}

// SetSize 设置尺寸
func (f *FlexLayout) SetSize(width, height int) {
	f.size.Width = width
	f.size.Height = height
}

// GetWidth 获取宽度
func (f *FlexLayout) GetWidth() int {
	return f.size.Width
}

// GetHeight 获取高度
func (f *FlexLayout) GetHeight() int {
	return f.size.Height
}

// SetDirection 设置弹性方向
func (f *FlexLayout) SetDirection(dir FlexDirection) {
	f.style.Direction = dir
}

// SetMainAxis 设置主轴对齐
func (f *FlexLayout) SetMainAxis(align MainAxisAlignment) {
	f.style.MainAxis = align
}

// SetCrossAxis 设置交叉轴对齐
func (f *FlexLayout) SetCrossAxis(align CrossAxisAlignment) {
	f.style.CrossAxis = align
}

// SetGap 设置主轴间距
func (f *FlexLayout) SetGap(gap int) {
	f.style.Gap = gap
}

// SetCrossGap 设置交叉轴间距
func (f *FlexLayout) SetCrossGap(gap int) {
	f.style.CrossGap = gap
}

// SetPadding 设置内边距
func (f *FlexLayout) SetPadding(left, right, top, bottom int) {
	f.style.Padding = Padding{
		Left:   left,
		Right:  right,
		Top:    top,
		Bottom: bottom,
	}
}

// GetFlexStyle 返回 FlexLayout 的样式
// 实现 FlexStyleProvider 接口，使 FlexLayout 可以被 Engine 正确布局
func (f *FlexLayout) GetFlexStyle() *FlexStyle {
	return f.style
}

// SetFlex 设置子节点的弹性配置
func (f *FlexLayout) SetFlex(index int, grow, shrink, basis int) {
	if f.style.FlexibleChildren == nil {
		f.style.FlexibleChildren = make(map[int]*Flex)
	}
	f.style.FlexibleChildren[index] = &Flex{
		Grow:   grow,
		Shrink: shrink,
		Basis:  basis,
	}
}

// Measure 测量节点尺寸
//
// FlexLayout 的 style.Padding 是内部配置，控制子节点区域
// 此方法返回包含 padding 的总尺寸
func (f *FlexLayout) Measure(constraints Constraints) Size {
	if len(f.children) == 0 {
		// 空容器返回 padding 占用
		paddingWidth := f.style.Padding.Horizontal()
		paddingHeight := f.style.Padding.Vertical()
		return Size{
			Width:  constraints.ConstrainWidth(paddingWidth),
			Height: constraints.ConstrainHeight(paddingHeight),
		}
	}

	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

	// Phase 1: 测量所有子节点
	childSizes := make([]Size, len(f.children))
	totalMainSize := 0
	maxCrossSize := 0
	inFlowCount := 0

	for i, child := range f.children {
		// 跳过 nil 子节点
		if child == nil {
			childSizes[i] = Size{Width: 0, Height: 0}
			continue
		}

		// 使用约束测量子节点
		if measurable, ok := child.(Measurable); ok {
			childSizes[i] = measurable.Measure(constraints)
		} else {
			// 尝试从 GetSize 获取尺寸
			w, h := child.GetSize()
			if w > 0 || h > 0 {
				childSizes[i] = Size{Width: w, Height: h}
			} else {
				// 最后使用约束的最大值作为默认
				childSizes[i] = Size{
					Width:  constraints.MaxWidth,
					Height: constraints.MaxHeight,
				}
			}
		}

		if isOutOfFlowFlexChild(child) {
			continue
		}

		inFlowCount++
		if isRow {
			// 横向布局：宽度累加，高度取最大
			if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
				// 可伸缩节点，使用 basis
				basis := flex.Basis
				if basis == 0 { // auto
					basis = childSizes[i].Width
				}
				totalMainSize += basis
				if totalMainSize > MaxInt {
					totalMainSize = MaxInt
				}
			} else {
				totalMainSize += childSizes[i].Width
				if totalMainSize > MaxInt {
					totalMainSize = MaxInt
				}
			}
			if childSizes[i].Height > maxCrossSize {
				maxCrossSize = childSizes[i].Height
			}
		} else {
			// 纵向布局：高度累加，宽度取最大
			if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
				basis := flex.Basis
				if basis == 0 { // auto
					basis = childSizes[i].Height
				}
				totalMainSize += basis
				if totalMainSize > MaxInt {
					totalMainSize = MaxInt
				}
			} else {
				totalMainSize += childSizes[i].Height
				if totalMainSize > MaxInt {
					totalMainSize = MaxInt
				}
			}
			if childSizes[i].Width > maxCrossSize {
				maxCrossSize = childSizes[i].Width
			}
		}
	}

	if inFlowCount == 0 {
		paddingWidth := f.style.Padding.Horizontal()
		paddingHeight := f.style.Padding.Vertical()
		return Size{
			Width:  constraints.ConstrainWidth(paddingWidth),
			Height: constraints.ConstrainHeight(paddingHeight),
		}
	}

	// 添加间距
	gapCount := inFlowCount - 1
	if gapCount > 0 {
		totalMainSize += f.style.Gap * gapCount
		if totalMainSize > MaxInt {
			totalMainSize = MaxInt
		}
	}

	// 计算包含 padding 的总尺寸
	var width, height int
	paddingWidth := f.style.Padding.Horizontal()
	paddingHeight := f.style.Padding.Vertical()

	if isRow {
		width = totalMainSize + paddingWidth
		height = maxCrossSize + paddingHeight
	} else {
		width = maxCrossSize + paddingWidth
		height = totalMainSize + paddingHeight
	}

	return Size{
		Width:  constraints.ConstrainWidth(width),
		Height: constraints.ConstrainHeight(height),
	}
}

// childConstraints 计算子节点约束
func (f *FlexLayout) childConstraints(constraints Constraints, index int) Constraints {
	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

	// 减去内边距
	availableMain := constraints.MaxWidth - f.style.Padding.Left - f.style.Padding.Right
	availableCross := constraints.MaxHeight - f.style.Padding.Top - f.style.Padding.Bottom

	// Ensure available space is non-negative
	if availableMain < 0 {
		availableMain = 0
	}
	if availableCross < 0 {
		availableCross = 0
	}
	if !isRow {
		availableMain, availableCross = availableCross, availableMain
	}

	if isRow {
		return NewConstraints(0, availableMain, 0, availableCross)
	}
	return NewConstraints(0, availableCross, 0, availableMain)
}

// LayoutChildren 布局子节点
func (f *FlexLayout) LayoutChildren(width, height int) []LayoutBox {
	if len(f.children) == 0 {
		return nil
	}

	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse
	isReverse := f.style.Direction == FlexRowReverse || f.style.Direction == FlexColumnReverse

	// 可用空间（减去内边距）
	availableWidth := width - f.style.Padding.Left - f.style.Padding.Right
	availableHeight := height - f.style.Padding.Top - f.style.Padding.Bottom

	// Phase 1: 测量所有子节点
	childSizes := make([]Size, len(f.children))
	childMarginContent := make([]int, len(f.children)) // 主轴方向的 margin 总和
	childMarginStart := make([]int, len(f.children))   // 主轴起始侧 margin
	childMarginCross := make([]int, len(f.children))   // 跨轴方向的 margin 总和
	childOutOfFlow := make([]bool, len(f.children))
	fixedTotal := 0    // 固定尺寸总和
	flexGrowTotal := 0 // flex-grow 总和
	flowIndices := make([]int, 0, len(f.children))

	for i, child := range f.children {
		// 跳过 nil children
		if child == nil {
			childSizes[i] = Size{Width: 0, Height: 0}
			continue
		}

		// 获取子节点的 margin
		marginSizeContent := 0
		marginSizeStart := 0
		marginSizeCross := 0
		if marginal, ok := child.(Marginal); ok {
			m := marginal.GetMargin()
			if isRow {
				marginSizeContent = m.Left + m.Right
				marginSizeStart = m.Left
				marginSizeCross = m.Top + m.Bottom
			} else {
				marginSizeContent = m.Top + m.Bottom
				marginSizeStart = m.Top
				marginSizeCross = m.Left + m.Right
			}
		}
		// 保存 margin 信息
		childMarginContent[i] = marginSizeContent
		childMarginStart[i] = marginSizeStart
		childMarginCross[i] = marginSizeCross

		constraints := Constraints{}
		if isRow {
			constraints = Constraints{
				MinWidth:  0,
				MaxWidth:  availableWidth,
				MinHeight: 0,
				MaxHeight: availableHeight,
			}
		} else {
			constraints = Constraints{
				MinWidth:  0,
				MaxWidth:  availableWidth,
				MinHeight: 0,
				MaxHeight: availableHeight,
			}
		}

		if measurable, ok := child.(Measurable); ok {
			childSizes[i] = measurable.Measure(constraints)
		} else {
			// 尝试从 GetSize 获取尺寸
			w, h := child.GetSize()
			if w > 0 || h > 0 {
				childSizes[i] = Size{Width: w, Height: h}
			} else {
				// 最后使用约束的最大值
				childSizes[i] = Size{
					Width:  constraints.MaxWidth,
					Height: constraints.MaxHeight,
				}
			}
		}

		childOutOfFlow[i] = isOutOfFlowFlexChild(child)
		if childOutOfFlow[i] {
			continue
		}
		flowIndices = append(flowIndices, i)

		if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
			flexGrowTotal += flex.Grow
			// 使用 basis 作为基础尺寸 + margin
			if flex.Basis > 0 {
				fixedTotal += flex.Basis + marginSizeContent
			} else {
				if isRow {
					fixedTotal += childSizes[i].Width + marginSizeContent
				} else {
					fixedTotal += childSizes[i].Height + marginSizeContent
				}
			}
		} else {
			// 固定尺寸节点，包含 margin
			if isRow {
				fixedTotal += childSizes[i].Width + marginSizeContent
			} else {
				fixedTotal += childSizes[i].Height + marginSizeContent
			}
		}
	}

	// Phase 2: 计算剩余空间
	gapCount := len(flowIndices) - 1
	totalGap := 0
	if gapCount > 0 {
		totalGap = f.style.Gap * gapCount
	}

	remainingSpace := 0
	if isRow {
		remainingSpace = availableWidth - fixedTotal - totalGap
	} else {
		remainingSpace = availableHeight - fixedTotal - totalGap
	}

	// Phase 3: 分配剩余空间给可伸缩节点
	finalSizes := make([]Size, len(f.children))
	for i := range f.children {
		if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
			// 按比例分配剩余空间
			// 注意: 剩余空间是 content space，不包含 margin
			extra := 0
			if flexGrowTotal > 0 {
				extra = (remainingSpace * flex.Grow) / flexGrowTotal
			}
			// finalSizes 包含 contentWidth + marginWidth
			if isRow {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width + childMarginContent[i] + extra,
					Height: childSizes[i].Height + childMarginCross[i],
				}
			} else {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width + childMarginCross[i],
					Height: childSizes[i].Height + childMarginContent[i] + extra,
				}
			}
		} else {
			// 固定尺寸节点，包含 margin
			if isRow {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width + childMarginContent[i],
					Height: childSizes[i].Height + childMarginCross[i],
				}
			} else {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width + childMarginCross[i],
					Height: childSizes[i].Height + childMarginContent[i],
				}
			}
		}
	}

	// Phase 4: 计算位置
	boxes := make([]LayoutBox, len(f.children))

	// 主轴起始位置
	mainPos := 0
	switch f.style.MainAxis {
	case MainStart:
		mainPos = 0
	case MainEnd:
		if isRow {
			mainPos = availableWidth - fixedTotal - totalGap
		} else {
			mainPos = availableHeight - fixedTotal - totalGap
		}
	case Center:
		if isRow {
			mainPos = (availableWidth - fixedTotal - totalGap) / 2
		} else {
			mainPos = (availableHeight - fixedTotal - totalGap) / 2
		}
	case SpaceBetween:
		mainPos = 0
	case SpaceAround, SpaceEvenly:
		// 需要额外计算间距
		mainPos = 0
	}

	flowCrossSize := f.getMaxCrossSizeForIndices(finalSizes, flowIndices)

	// 交叉轴起始位置
	crossPos := 0
	switch f.style.CrossAxis {
	case CrossStart:
		crossPos = 0
	case CrossEnd:
		if isRow {
			crossPos = availableHeight - flowCrossSize
		} else {
			crossPos = availableWidth - flowCrossSize
		}
	case CrossCenter:
		if isRow {
			crossPos = (availableHeight - flowCrossSize) / 2
		} else {
			crossPos = (availableWidth - flowCrossSize) / 2
		}
	case Stretch:
		crossPos = 0
	case Baseline:
		// Baseline alignment - will be handled per child
		crossPos = 0
	}

	// For baseline alignment, calculate max baseline
	maxBaseline := 0
	if f.style.CrossAxis == Baseline && isRow {
		for _, i := range flowIndices {
			child := f.children[i]
			if child == nil {
				continue
			}
			baseline := GetBaselineFromNode(child)
			// The position is based on baseline, need to account for content below baseline
			contentBelowBaseline := finalSizes[i].Height - baseline
			if contentBelowBaseline > maxBaseline {
				maxBaseline = contentBelowBaseline
			}
		}
	}

	// SpaceBetween/Around/Evenly 的额外间距
	extraGap := 0
	if (f.style.MainAxis == SpaceBetween || f.style.MainAxis == SpaceAround || f.style.MainAxis == SpaceEvenly) && len(flowIndices) > 1 {
		switch f.style.MainAxis {
		case SpaceBetween:
			extraGap = remainingSpace / gapCount
		case SpaceAround:
			extraGap = remainingSpace / len(flowIndices)
		case SpaceEvenly:
			extraGap = remainingSpace / (len(flowIndices) + 1)
		}
	}

	if f.style.MainAxis == SpaceEvenly {
		mainPos += extraGap
	} else if f.style.MainAxis == SpaceAround {
		mainPos += extraGap / 2
	}

	// 布局主流中的子节点
	visualIndices := flowIndices
	if isReverse {
		visualIndices = make([]int, len(flowIndices))
		for i := range flowIndices {
			visualIndices[i] = flowIndices[len(flowIndices)-1-i]
		}
	}

	for visualIndex, childIdx := range visualIndices {
		child := f.children[childIdx]
		if child == nil {
			continue
		}

		var x, y int
		if isRow {
			x = f.style.Padding.Left + mainPos
			y = f.style.Padding.Top + crossPos
		} else {
			x = f.style.Padding.Left + crossPos
			y = f.style.Padding.Top + mainPos
		}

		// 根据方向增加 mainPos
		if isRow {
			mainPos += finalSizes[childIdx].Width + f.style.Gap
		} else {
			mainPos += finalSizes[childIdx].Height + f.style.Gap
		}
		if extraGap > 0 && visualIndex < len(visualIndices)-1 {
			mainPos += extraGap
		}

		// Stretch 处理 - 使用正确的索引
		if f.style.CrossAxis == Stretch {
			if isRow {
				finalSizes[childIdx].Height = availableHeight
			} else {
				finalSizes[childIdx].Width = availableWidth
			}
		}

		// Baseline 处理 - 调整 Y 位置以对齐基线
		if f.style.CrossAxis == Baseline && isRow {
			baseline := GetBaselineFromNode(child)
			// 计算基线偏移: 最大基线下方内容 - 当前节点的基线下方内容
			contentBelowBaseline := finalSizes[childIdx].Height - baseline
			yOffset := maxBaseline - contentBelowBaseline
			y += yOffset
		}

		// childBox.X/Y 应该指向子节点的左上角（包括 margin 偏移）
		// 主轴方向：mainPos 已经包含了 start margin（因为 finalSizes 包含 margin）
		// 但是这里 mainPos 指向的是不含 margin 的位置，所以需要加上 start margin
		// 注意: finalSizes.ChildIdx 已经包含 marginTotal
		// mainPos += finalSizes 包含了 margin，所以 mainPos 本身就是下一个子节点的"内容位置"
		// childBox.X/Y 应该指向完整的盒子左上角，所以不需要再加 margin
		// ✅ 修正: childBox.X/Y 已经包含了 margin 空间（因为 mainPos 使用了 finalSizes）
		boxes[childIdx] = LayoutBox{
			ID:     child.ID(),
			X:      x,
			Y:      y,
			Width:  finalSizes[childIdx].Width,
			Height: finalSizes[childIdx].Height,
		}

		// 设置子节点位置和尺寸
		child.SetPosition(x, y)
		child.SetSize(finalSizes[childIdx].Width, finalSizes[childIdx].Height)
	}

	for i, child := range f.children {
		if child == nil || !childOutOfFlow[i] {
			continue
		}

		boxes[i] = LayoutBox{
			ID:     child.ID(),
			X:      f.style.Padding.Left,
			Y:      f.style.Padding.Top,
			Width:  finalSizes[i].Width,
			Height: finalSizes[i].Height,
		}
		child.SetPosition(boxes[i].X, boxes[i].Y)
		child.SetSize(finalSizes[i].Width, finalSizes[i].Height)
	}

	return boxes
}

func isOutOfFlowFlexChild(child Node) bool {
	if child == nil {
		return false
	}
	posProvider, ok := child.(PositionProvider)
	if !ok {
		return false
	}
	return isOutOfFlowPosition(posProvider.GetPositionType())
}

// getMaxCrossSize 获取最大交叉轴尺寸
func (f *FlexLayout) getMaxCrossSize(sizes []Size) int {
	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse
	maxSize := 0
	for _, size := range sizes {
		if isRow {
			if size.Height > maxSize {
				maxSize = size.Height
			}
		} else {
			if size.Width > maxSize {
				maxSize = size.Width
			}
		}
	}
	return maxSize
}

func (f *FlexLayout) getMaxCrossSizeForIndices(sizes []Size, indices []int) int {
	if len(indices) == 0 {
		return 0
	}
	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse
	maxSize := 0
	for _, index := range indices {
		if index < 0 || index >= len(sizes) {
			continue
		}
		if isRow {
			if sizes[index].Height > maxSize {
				maxSize = sizes[index].Height
			}
		} else {
			if sizes[index].Width > maxSize {
				maxSize = sizes[index].Width
			}
		}
	}
	return maxSize
}

// =============================================================================
// FlexShrink Calculation
// =============================================================================

// ShrinkInfo contains information needed for shrink calculation
type ShrinkInfo struct {
	// Index of the child
	Index int

	// Original size before shrink
	OriginalSize int

	// Shrink factor (from Flex.Shrink)
	ShrinkFactor int

	// Minimum size the child can shrink to
	MinSize int
}

// CalculateShrinkDistribution calculates how to distribute shrink among children
// when the container is smaller than the total natural size of children.
//
// Parameters:
//   - deficit: How much space needs to be removed (positive value)
//   - children: Shrink information for each shrinkable child
//
// Returns:
//   - Map of child index to amount to shrink (positive values)
func CalculateShrinkDistribution(deficit int, children []ShrinkInfo) map[int]int {
	result := make(map[int]int)

	if deficit <= 0 || len(children) == 0 {
		return result
	}

	// Calculate total shrink factor (weighted by original size)
	totalShrinkWeight := 0
	for _, child := range children {
		if child.ShrinkFactor > 0 {
			// Weight = shrink factor * original size (CSS flex-shrink behavior)
			totalShrinkWeight += child.ShrinkFactor * child.OriginalSize
		}
	}

	if totalShrinkWeight == 0 {
		return result
	}

	remainingDeficit := deficit

	// Distribute shrink proportionally
	for _, child := range children {
		if child.ShrinkFactor <= 0 {
			continue
		}

		// Calculate this child's share of the shrink
		weight := child.ShrinkFactor * child.OriginalSize
		share := (deficit * weight) / totalShrinkWeight

		// Don't shrink below minimum size
		maxShrink := child.OriginalSize - child.MinSize
		if share > maxShrink {
			share = maxShrink
		}

		if share > 0 {
			result[child.Index] = share
			remainingDeficit -= share
		}
	}

	// If there's still deficit, distribute it among children that can still shrink
	for remainingDeficit > 0 {
		distributed := false
		for _, child := range children {
			if child.ShrinkFactor <= 0 {
				continue
			}

			currentShrink := result[child.Index]
			maxShrink := child.OriginalSize - child.MinSize

			if currentShrink < maxShrink {
				// Can still shrink more
				extra := 1
				if currentShrink+extra > maxShrink {
					extra = maxShrink - currentShrink
				}
				if extra > 0 {
					result[child.Index] += extra
					remainingDeficit -= extra
					distributed = true
					if remainingDeficit <= 0 {
						break
					}
				}
			}
		}
		if !distributed {
			break
		}
	}

	return result
}

// ApplyShrinkToSizes applies shrink amounts to a slice of sizes
// isRow: true for horizontal layout (shrink width), false for vertical (shrink height)
func ApplyShrinkToSizes(sizes []Size, shrinkAmounts map[int]int, isRow bool) {
	for idx, shrink := range shrinkAmounts {
		if idx < len(sizes) {
			if isRow {
				sizes[idx].Width -= shrink
				if sizes[idx].Width < 0 {
					sizes[idx].Width = 0
				}
			} else {
				sizes[idx].Height -= shrink
				if sizes[idx].Height < 0 {
					sizes[idx].Height = 0
				}
			}
		}
	}
}

// GetShrinkableChildren extracts shrink information from flex children
func GetShrinkableChildren(children []Node, flexConfig map[int]*Flex, sizes []Size, isRow bool) []ShrinkInfo {
	var shrinkable []ShrinkInfo

	for i, child := range children {
		if child == nil {
			continue
		}

		flex, ok := flexConfig[i]
		if !ok || flex.Shrink <= 0 {
			continue
		}

		originalSize := 0
		if isRow {
			originalSize = sizes[i].Width
		} else {
			originalSize = sizes[i].Height
		}

		shrinkable = append(shrinkable, ShrinkInfo{
			Index:        i,
			OriginalSize: originalSize,
			ShrinkFactor: flex.Shrink,
			MinSize:      0, // Default minimum
		})
	}

	return shrinkable
}

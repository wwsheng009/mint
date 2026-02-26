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
		Direction:  FlexColumn,
		MainAxis:   MainStart,
		CrossAxis:  CrossStart,
		Gap:        0,
		CrossGap:   0,
		Padding:    Padding{},
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
func (f *FlexLayout) Measure(constraints Constraints) Size {
	if len(f.children) == 0 {
		width := constraints.ConstrainWidth(f.style.Padding.Left + f.style.Padding.Right)
		height := constraints.ConstrainHeight(f.style.Padding.Top + f.style.Padding.Bottom)
		return Size{Width: width, Height: height}
	}

	isRow := f.style.Direction == FlexRow || f.style.Direction == FlexRowReverse

	// Phase 1: 测量所有子节点
	childSizes := make([]Size, len(f.children))
	totalMainSize := 0
	maxCrossSize := 0

	for i, child := range f.children {
		// 跳过 nil 子节点
		if child == nil {
			childSizes[i] = Size{Width: 0, Height: 0}
			continue
		}

		childConstraints := f.childConstraints(constraints, i)
		if measurable, ok := child.(Measurable); ok {
			childSizes[i] = measurable.Measure(childConstraints)
		} else {
			// 尝试从 GetSize 获取尺寸
			w, h := child.GetSize()
			if w > 0 || h > 0 {
				childSizes[i] = Size{Width: w, Height: h}
			} else {
				// 最后使用约束的最大值作为默认
				childSizes[i] = Size{
					Width:  childConstraints.MaxWidth,
					Height: childConstraints.MaxHeight,
				}
			}
		}

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

	// 添加间距
	gapCount := len(f.children) - 1
	if gapCount > 0 {
		totalMainSize += f.style.Gap * gapCount
		if totalMainSize > MaxInt {
			totalMainSize = MaxInt
		}
	}

	// 计算总尺寸
	var width, height int
	if isRow {
		width = f.style.Padding.Left + totalMainSize + f.style.Padding.Right
		height = f.style.Padding.Top + maxCrossSize + f.style.Padding.Bottom
	} else {
		width = f.style.Padding.Left + maxCrossSize + f.style.Padding.Right
		height = f.style.Padding.Top + totalMainSize + f.style.Padding.Bottom
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
	fixedTotal := 0    // 固定尺寸总和
	flexGrowTotal := 0 // flex-grow 总和

	for i, child := range f.children {
		// 跳过 nil children
		if child == nil {
			childSizes[i] = Size{Width: 0, Height: 0}
			continue
		}

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

		if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
			flexGrowTotal += flex.Grow
			// 使用 basis 作为基础尺寸
			if flex.Basis > 0 {
				fixedTotal += flex.Basis
			} else {
				if isRow {
					fixedTotal += childSizes[i].Width
				} else {
					fixedTotal += childSizes[i].Height
				}
			}
		} else {
			// 固定尺寸节点
			if isRow {
				fixedTotal += childSizes[i].Width
			} else {
				fixedTotal += childSizes[i].Height
			}
		}
	}

	// Phase 2: 计算剩余空间
	gapCount := len(f.children) - 1
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
	flexIndex := 0
	for i := range f.children {
		if flex, ok := f.style.FlexibleChildren[i]; ok && flex.Grow > 0 {
			// 按比例分配剩余空间
			extra := 0
			if flexGrowTotal > 0 {
				extra = (remainingSpace * flex.Grow) / flexGrowTotal
			}
			if isRow {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width + extra,
					Height: childSizes[i].Height,
				}
			} else {
				finalSizes[i] = Size{
					Width:  childSizes[i].Width,
					Height: childSizes[i].Height + extra,
				}
			}
			flexIndex++
		} else {
			finalSizes[i] = childSizes[i]
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

	// 交叉轴起始位置
	crossPos := 0
	switch f.style.CrossAxis {
	case CrossStart:
		crossPos = 0
	case CrossEnd:
		if isRow {
			crossPos = availableHeight - f.getMaxCrossSize(finalSizes)
		} else {
			crossPos = availableWidth - f.getMaxCrossSize(finalSizes)
		}
	case CrossCenter:
		if isRow {
			crossPos = (availableHeight - f.getMaxCrossSize(finalSizes)) / 2
		} else {
			crossPos = (availableWidth - f.getMaxCrossSize(finalSizes)) / 2
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
		for i, child := range f.children {
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
	if (f.style.MainAxis == SpaceBetween || f.style.MainAxis == SpaceAround || f.style.MainAxis == SpaceEvenly) && len(f.children) > 1 {
		switch f.style.MainAxis {
		case SpaceBetween:
			extraGap = remainingSpace / gapCount
		case SpaceAround:
			extraGap = remainingSpace / len(f.children)
		case SpaceEvenly:
			extraGap = remainingSpace / (len(f.children) + 1)
		}
	}

	if f.style.MainAxis == SpaceEvenly {
		mainPos += extraGap
	} else if f.style.MainAxis == SpaceAround {
		mainPos += extraGap / 2
	}

	// 布局每个子节点
	for i, child := range f.children {
		// Skip nil children
		if child == nil {
			continue
		}

		var x, y int
		var childIdx int // 实际的子节点索引
		if isReverse {
			childIdx = len(f.children) - 1 - i
		} else {
			childIdx = i
		}

		if isReverse {
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
			if extraGap > 0 && i < len(f.children)-1 {
				mainPos += extraGap
			}
		} else {
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
			if extraGap > 0 && i < len(f.children)-1 {
				mainPos += extraGap
			}
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

		boxes[i] = LayoutBox{
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

	return boxes
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

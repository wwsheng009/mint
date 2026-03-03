package layout

// NewEngine 创建新的布局引擎
func NewEngine() *Engine {
	return &Engine{
		dirty: NewDirtyTracker(),
		stats: LayoutStats{},
		cache: &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize: 1000,
		},
		flexCache:      NewFlexCache(),
		hitMap:         NewHitMap(),
		layerManager:   NewLayerManager(),
		overlayManager: NewOverlayManager(),
	}
}

// Invalidate 使整个布局树失效
func (e *Engine) Invalidate() {
	e.dirty.Clear()
}

// InvalidateNode 使单个节点失效
func (e *Engine) InvalidateNode(id string) {
	e.dirty.MarkLayoutDirty(id)
}

// Layout 执行布局计算
// 输入根节点和约束，返回布局结果
func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult {
	if root == nil {
		return &LayoutResult{}
	}

	// ✨ 保存 viewport 约束（根节点的原始约束）
	// 这样 Fixed 定位节点可以使用完整的 viewport 尺寸进行定位计算
	e.viewportConstraints = constraints

	// 检查缓存
	if e.cache != nil {
		if cached := e.cache.Get(root, constraints); cached != nil {
			e.stats.CacheHits++
			return cached
		}
		e.stats.CacheMisses++
	}

	result := &LayoutResult{
		Boxes: make([]LayoutBox, 0),
		Dirty: true,
	}

	// 递归布局节点
	box := e.layoutNode(root, constraints, 0, 0)
	result.Root = box
	result.Boxes = e.collectBoxes(box)

	// 构建命中映射表
	if e.hitMap != nil {
		result.HitMap = e.hitMap
		e.hitMap.BuildFromLayoutBox(box)
	}

	// 存入缓存
	if e.cache != nil {
		e.cache.Put(root, constraints, result)
	}

	return result
}

// LayoutIncremental 执行增量布局计算
// 使用脏标记跳过干净的节点
func (e *Engine) LayoutIncremental(root Node, constraints Constraints) *LayoutResult {
	if root == nil {
		return &LayoutResult{}
	}

	// 检查缓存
	if e.cache != nil {
		if cached := e.cache.Get(root, constraints); cached != nil {
			e.stats.CacheHits++
			return cached
		}
		e.stats.CacheMisses++
	}

	result := &LayoutResult{
		Boxes: make([]LayoutBox, 0),
		Dirty: true,
	}

	// 使用增量布局
	box := e.layoutNodeIncremental(root, constraints, 0, 0)
	result.Root = box
	result.Boxes = e.collectBoxes(box)

	// 构建命中映射表
	if e.hitMap != nil {
		result.HitMap = e.hitMap
		e.hitMap.BuildFromLayoutBox(box)
	}

	// 存入缓存
	if e.cache != nil {
		e.cache.Put(root, constraints, result)
	}

	// 清除脏标记
	e.clearDirtyMarkers(root)

	return result
}


// layoutNode 递归布局单个节点
func (e *Engine) layoutNode(node Node, constraints Constraints, x, y int) *LayoutBox {
	return e.layoutNodeWithDepth(node, constraints, x, y, 0, make(map[string]bool))
}

// layoutNodeWithDepth 递归布局单个节点，带深度限制和循环检测
func (e *Engine) layoutNodeWithDepth(node Node, constraints Constraints, x, y int, depth int, visited map[string]bool) *LayoutBox {
	if node == nil {
		return nil
	}

	// 深度限制检查
	if depth > MaxLayoutDepth {
		// 达到最大深度，返回最小尺寸的盒子
		return &LayoutBox{
			ID:      node.ID(),
			X:       x,
			Y:       y,
			AbsX:    x,  // ✨ Phase 1.2: 保存全局坐标
			AbsY:    y,
			Width:   0,
			Height:  0,
			Children: make([]*LayoutBox, 0),
		}
	}

	// 循环检测
	nodeID := node.ID()
	if nodeID != "" {
		if visited[nodeID] {
			// 检测到循环，返回空盒子避免无限递归
			return &LayoutBox{
				ID:      node.ID() + "_cycle",
				X:       x,
				Y:       y,
				AbsX:    x,  // ✨ Phase 1.2: 保存全局坐标
				AbsY:    y,
				Width:   0,
				Height:  0,
				Children: make([]*LayoutBox, 0),
			}
		}
		visited[nodeID] = true
		defer delete(visited, nodeID) // 离开时清除标记
	}

	// 获取节点尺寸
	width, height := node.GetSize()

	// 如果节点实现了 Measurable 接口，测量其尺寸
	if measurable, ok := node.(Measurable); ok {
		size := measurable.Measure(constraints)
		width, height = size.Width, size.Height
	}

	// Get Layer and ZIndex from node if it implements Layered interface
	layer := GetLayerFromNode(node)
	shouldCenter := false // ✨ 初始化居中标志（Phase 1.1 + Phase 2.3）
	zIndex := GetZIndexFromNode(node)

	// Get Border from node if it implements Bordered interface
	nodeBorder := GetBorderFromNode(node)

	// ✨ Phase 2.3: PositionFixed 处理
	// 如果节点使用了 fixed 定位，以 Root 为参考系，不受父布局影响
	position := PositionRelative
	anchor := AnchorTopLeft

	posProvider, ok := node.(PositionProvider)
	if ok {
		position = posProvider.GetPositionType()
		anchor = posProvider.GetAnchor()
	}


	// Fixed 定位：使用 viewport 约束重新计算坐标
	if position == PositionFixed && width > 0 && height > 0 {
		// ✨ 使用保存的 viewport 约束而不是传入的约束
		// Modal 等组件需要以完整的 viewport 尺寸作为参考系进行定位
		// 而不是受父布局限制后的约束
		rootW := e.viewportConstraints.MaxWidth
		rootH := e.viewportConstraints.MaxHeight

		// 根据 Anchor 计算固定定位坐标
		switch anchor {
		case AnchorTopLeft:
			x, y = 0, 0
		case AnchorTop:
			x = (rootW - width) / 2
			y = 0
		case AnchorTopRight:
			x = rootW - width
			y = 0
		case AnchorLeft:
			x = 0
			y = (rootH - height) / 2
		case AnchorCenter:
			x = (rootW - width) / 2
			y = (rootH - height) / 2
			shouldCenter = true // ✨ 居中定位时设置标志
		case AnchorRight:
			x = rootW - width
			y = (rootH - height) / 2
		case AnchorBottomLeft:
			x = 0
			y = rootH - height
		case AnchorBottom:
			x = (rootW - width) / 2
			y = rootH - height
		case AnchorBottomRight:
			x = rootW - width
			y = rootH - height
		default:
			x, y = 0, 0
		}
	}

	// ✨ Phase 1.3: 设置全局坐标
	// x, y 已经是传入的全局坐标（由父节点累积）
	absX, absY := x, y

	// Get PropsID from node if it implements PropsIDProvider interface
	propsID := ""
	if propsIDProvider, ok := node.(PropsIDProvider); ok {
		propsID = propsIDProvider.GetPropsID()
	}

	box := &LayoutBox{
		ID:       node.ID(),
		Tag:      node.Type(),  // ✨ 设置 Tag 用于调试和类型识别
		PropsID:  propsID,
		X:        x,
		Y:        y,
		AbsX:     absX,  // ✨ Phase 1.2: 保存全局坐标
		AbsY:     absY,
		Width:    width,
		Height:   height,
		Layer:    layer,
		ZIndex:   zIndex,
		ShouldCenter: shouldCenter,  // ✨ Phase 1.1: 保存居中标记
		Border:   nodeBorder,
		Children: make([]*LayoutBox, 0),
	}

	// 设置节点位置和尺寸
	node.SetPosition(x, y)
	node.SetSize(width, height)

	// 计算内容区域偏移（用于布局子节点）
	// 优先使用 BoxModelProvider，否则使用 Bordered 接口
	contentOffsetX, contentOffsetY := 0, 0
	if boxModelProvider, ok := node.(BoxModelProvider); ok {
		boxModel := boxModelProvider.GetBoxModel()
		contentOffsetX = boxModel.ContentOffsetX()
		contentOffsetY = boxModel.ContentOffsetY()
	} else if nodeBorder.HasBorder() {
		contentOffsetX, contentOffsetY = nodeBorder.ContentOffset()
	}
	// 兼容旧的 borderOffsetX 变量名
	borderOffsetX := contentOffsetX
	borderOffsetY := contentOffsetY

	// 检查节点是否实现了 FlexStyleProvider 接口
	// 如果是，使用 FlexLayout 进行子节点布局
	if flexProvider, ok := node.(FlexStyleProvider); ok {
		flexStyle := flexProvider.GetFlexStyle()
		// 检查是否有非 nil 的子节点
		hasValidChildren := false
		for _, child := range node.Children() {
			if child != nil {
				hasValidChildren = true
				break
			}
		}
		if flexStyle != nil && hasValidChildren {
			// 使用 FlexLayout 进行布局
			flex := NewFlexLayout(node.ID(), node.Children())
			flex.SetDirection(flexStyle.Direction)
			flex.SetMainAxis(flexStyle.MainAxis)
			flex.SetCrossAxis(flexStyle.CrossAxis)
			flex.SetGap(flexStyle.Gap)
			flex.SetPadding(flexStyle.Padding.Left, flexStyle.Padding.Right, flexStyle.Padding.Top, flexStyle.Padding.Bottom)

			// 设置子节点的 flex 属性
			children := node.Children()
			for i, child := range children {
				if flexChild, ok := child.(FlexChildProvider); ok {
					childFlex := flexChild.GetFlex()
					if childFlex > 0 {
						flex.SetFlex(i, childFlex, 0, 0)
					}
				}
			}

			// 计算子节点可用的内部空间（减去 padding/border）
			// 优先使用 BoxModelProvider，否则使用 Bordered 接口
			var innerWidth, innerHeight int
			if boxModelProvider, ok := node.(BoxModelProvider); ok {
				boxModel := boxModelProvider.GetBoxModel()
				innerWidth = boxModel.InnerWidth(width)
				innerHeight = boxModel.InnerHeight(height)
			} else {
				innerWidth = width - nodeBorder.HorizontalPadding()
				innerHeight = height - nodeBorder.VerticalPadding()
				if innerWidth < 0 {
					innerWidth = 0
				}
				if innerHeight < 0 {
					innerHeight = 0
				}
			}

			// FlexLayout 的 padding 是内部配置，用于控制子节点布局区域
			// 这与 BoxModel 的 padding（外部边距）不同，需要手动扣除
			// TODO: 未来考虑将 FlexLayout 的 padding 迁移到 BoxModel 语义
			innerWidth = max(0, innerWidth-flexStyle.Padding.Horizontal())
			innerHeight = max(0, innerHeight-flexStyle.Padding.Vertical())

			// 布局子节点
			childBoxes := flex.LayoutChildren(innerWidth, innerHeight)

			// ✨ 重新计算子节点位置以考虑主轴方向上的 margin
			// FlexLayout 不考虑 margin，所以我们需要手动调整主轴位置
			// 思路：childBox.Y/X 已经包含了 gap，但需要在主轴方向上额外添加前面的兄弟节点的 margin
			isFlexRow := flexStyle.Direction == FlexRow || flexStyle.Direction == FlexRowReverse
			mainAxisMarginOffset := 0 // 累积前面所有兄弟节点的主轴边距

			for i, childBox := range childBoxes {
				// 递归布局子节点的子节点
				child := node.Children()[i]
				if child != nil {
					// 获取子节点的 margin（如果实现了 Marginal 接口）
					marginTop, marginBottom, marginLeft, marginRight := 0, 0, 0, 0
					if marginal, ok := child.(Marginal); ok {
						m := marginal.GetMargin()
						marginTop = m.Top
						marginBottom = m.Bottom
						marginLeft = m.Left
						marginRight = m.Right
					}

					// 主轴方向：childBox.X/Y 是不含 margin 的位置
					// 需要加上：前面所有兄弟节点的累积 margin + 当前节点的起始侧 margin
					var childX, childY int
					if isFlexRow {
						// Row: X 是主轴，Y 是跨轴
						childX = x + childBox.X + borderOffsetX + mainAxisMarginOffset + marginLeft
						childY = y + childBox.Y + borderOffsetY + marginTop  // ✅ 添加跨轴垂直 margin
						// 为下一个节点累积：currentRightMargin + nextLeftMargin
						mainAxisMarginOffset += marginLeft + marginRight
					} else {
						// Column: Y 是主轴，X 是跨轴
						childY = y + childBox.Y + borderOffsetY + mainAxisMarginOffset + marginTop
						childX = x + childBox.X + borderOffsetX + marginLeft  // ✅ 添加跨轴水平 margin
						// 为下一个节点累积：currentBottomMargin + nextTopMargin
						mainAxisMarginOffset += marginTop + marginBottom
					}

					// ✨ FIX: 为子节点创建正确的约束，基于 Flex 分配的尺寸并扣除 margin
					// 这样嵌套布局（如 VStack 内嵌 HStack）可以使用正确的约束
					childConstraints := Constraints{
						MinWidth:  max(0, childBox.Width-marginLeft-marginRight),
						MaxWidth:  max(0, childBox.Width-marginLeft-marginRight),
						MinHeight: max(0, childBox.Height-marginTop-marginBottom),
						MaxHeight: max(0, childBox.Height-marginTop-marginBottom),
					}

					subBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)
					if subBox != nil {
						// 🔴 BUG FIX: Fixed 定位的子节点不能被父容器覆盖位置
						// Modal 等 Fixed 定位节点已经使用 viewportConstraints 计算了正确位置
						// 只对非 Fixed 子节点使用 FlexLayout 计算的位置
						childPosition := PositionRelative
						if posProvider, ok := child.(PositionProvider); ok {
							childPosition = posProvider.GetPositionType()
						}
						
						if childPosition != PositionFixed {
							// 使用 FlexLayout 计算的位置和尺寸
							subBox.X = childX
							subBox.Y = childY
						}
						// Fixed 定位节点：subBox.X 和 subBox.Y 保持不变（已在 layoutNodeWithDepth 中计算）
						
						box.Children = append(box.Children, subBox)
					}
				}
			}
			return box
		}
	}

	// 检查节点是否实现了 GridStyleProvider 接口
	// 如果是，使用 GridLayout 进行子节点布局
	if gridProvider, ok := node.(GridStyleProvider); ok {
		gridStyle := gridProvider.GetGridStyle()
		if gridStyle != nil && (len(gridStyle.Cells) > 0 || len(node.Children()) > 0) {
			// 使用 GridLayout 进行布局
			grid := NewGridLayout(node.ID(), gridStyle)
			grid.SetChildren(node.Children())

			// 计算子节点可用的内部空间（减去 padding/border）
			// 优先使用 BoxModelProvider，否则使用 Bordered 接口
			var innerWidth, innerHeight int
			if boxModelProvider, ok := node.(BoxModelProvider); ok {
				boxModel := boxModelProvider.GetBoxModel()
				innerWidth = boxModel.InnerWidth(width)
				innerHeight = boxModel.InnerHeight(height)
			} else {
				innerWidth = width - nodeBorder.HorizontalPadding()
				innerHeight = height - nodeBorder.VerticalPadding()
				if innerWidth < 0 {
					innerWidth = 0
				}
				if innerHeight < 0 {
					innerHeight = 0
				}
			}

			// 布局子节点
			childBoxes := grid.LayoutChildren(innerWidth, innerHeight)
			for i, childBox := range childBoxes {
				// 递归布局子节点的子节点
				// 需要找到对应的 child 节点
				var child Node
				if len(gridStyle.Cells) > 0 && i < len(gridStyle.Cells) {
					child = gridStyle.Cells[i].Child
				} else if i < len(node.Children()) {
					child = node.Children()[i]
				}
				if child != nil {
					// 获取子节点的 margin（如果实现了 Marginal 接口）
					marginTop, marginBottom, marginLeft, marginRight := 0, 0, 0, 0
					if marginal, ok := child.(Marginal); ok {
						m := marginal.GetMargin()
						marginTop = m.Top
						marginBottom = m.Bottom
						marginLeft = m.Left
						marginRight = m.Right
					}

					childX := x + childBox.X + borderOffsetX + marginLeft
					childY := y + childBox.Y + borderOffsetY + marginTop

					// ✨ FIX: 为子节点创建正确的约束，基于分配的尺寸并扣除 margin
					childConstraints := Constraints{
						MinWidth:  max(0, childBox.Width-marginLeft-marginRight),
						MaxWidth:  max(0, childBox.Width-marginLeft-marginRight),
						MinHeight: max(0, childBox.Height-marginTop-marginBottom),
						MaxHeight: max(0, childBox.Height-marginTop-marginBottom),
					}

					subBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)
					if subBox != nil {
						// 🔴 BUG FIX: Fixed 定位的子节点不能被父容器覆盖位置
						childPosition := PositionRelative
						if posProvider, ok := child.(PositionProvider); ok {
							childPosition = posProvider.GetPositionType()
						}
						
						if childPosition != PositionFixed {
							subBox.X = childX
							subBox.Y = childY
						}
						
						box.Children = append(box.Children, subBox)
					}
				}
			}
			return box
		}
	}

	// 检查节点是否实现了 WrapStyleProvider 接口
	// 如果是，使用 WrapLayout 进行子节点布局（换行布局）
	if wrapProvider, ok := node.(WrapStyleProvider); ok {
		wrapStyle := wrapProvider.GetWrapStyle()
		if wrapStyle != nil && len(node.Children()) > 0 {
			// 使用 WrapLayout 进行布局
			wrap := NewWrapLayout(node.ID(), wrapStyle)
			wrap.SetChildren(node.Children())

			// 布局子节点
			childBoxes := wrap.LayoutChildren(width, height)
			for i, childBox := range childBoxes {
				child := node.Children()[i]
				if child != nil {
					// 获取子节点的 margin（如果实现了 Marginal 接口）
					marginTop, marginBottom, marginLeft, marginRight := 0, 0, 0, 0
					if marginal, ok := child.(Marginal); ok {
						m := marginal.GetMargin()
						marginTop = m.Top
						marginBottom = m.Bottom
						marginLeft = m.Left
						marginRight = m.Right
					}

					childX := x + childBox.X + borderOffsetX + marginLeft
					childY := y + childBox.Y + borderOffsetY + marginTop

					// ✨ FIX: 为子节点创建正确的约束，基于分配的尺寸并扣除 margin
					childConstraints := Constraints{
						MinWidth:  max(0, childBox.Width-marginLeft-marginRight),
						MaxWidth:  max(0, childBox.Width-marginLeft-marginRight),
						MinHeight: max(0, childBox.Height-marginTop-marginBottom),
						MaxHeight: max(0, childBox.Height-marginTop-marginBottom),
					}

					subBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, depth+1, visited)
					if subBox != nil {
						// 🔴 BUG FIX: Fixed 定位的子节点不能被父容器覆盖位置
						childPosition := PositionRelative
						if posProvider, ok := child.(PositionProvider); ok {
							childPosition = posProvider.GetPositionType()
						}
						
						if childPosition != PositionFixed {
							subBox.X = childX
							subBox.Y = childY
						}
						
						box.Children = append(box.Children, subBox)
					}
				}
			}
			return box
		}
	}

	// 检查节点是否实现了 AbsoluteStyleProvider 接口
	// 如果是，使用绝对定位进行子节点布局
	if absProvider, ok := node.(AbsoluteStyleProvider); ok {
		absStyle := absProvider.GetAbsoluteStyle()
		if absStyle != nil {
			// 绝对定位容器：子元素相对于容器定位
			// 使用 absolute 节点的尺寸作为容器尺寸
			// 如果尺寸为 0，使用约束的最大值
			containerWidth := width
			containerHeight := height
			if containerWidth <= 0 {
				containerWidth = constraints.MaxWidth
			}
			if containerHeight <= 0 {
				containerHeight = constraints.MaxHeight
			}
			for _, child := range node.Children() {
				// 获取子节点的 margin（如果实现了 Marginal 接口）
				// 注意：绝对定位只应用 margin 的位置偏移，不影响约束
				marginTop, marginLeft := 0, 0
				if marginal, ok := child.(Marginal); ok {
					m := marginal.GetMargin()
					marginTop = m.Top
					marginLeft = m.Left
				}

				// 获取子元素尺寸
				childWidth, childHeight := child.GetSize()

				// 如果子元素实现了 Measurable，测量其尺寸
				// 注意：绝对定位的约束不受 margin 影响
				if measurable, ok := child.(Measurable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  containerWidth,
						MinHeight: 0,
						MaxHeight: containerHeight,
					}
					size := measurable.Measure(childConstraints)
					childWidth = size.Width
					childHeight = size.Height
				}

				// 使用 AbsoluteStyle 计算子元素位置
				childX, childY := absStyle.CalculatePosition(containerWidth, containerHeight, childWidth, childHeight)

				// 应用 margin 偏移
				childX += marginLeft
				childY += marginTop

				// 递归布局子节点
				subBox := e.layoutNodeWithDepth(child, constraints, x+childX+borderOffsetX, y+childY+borderOffsetY, depth+1, visited)
				if subBox != nil {
					// 🔴 BUG FIX: Fixed 定位的子节点不能被父容器覆盖位置
					childPosition := PositionRelative
					if posProvider, ok := child.(PositionProvider); ok {
						childPosition = posProvider.GetPositionType()
					}

					if childPosition != PositionFixed {
						subBox.X = x + childX + borderOffsetX
						subBox.Y = y + childY + borderOffsetY
					}
					box.Children = append(box.Children, subBox)
				}
			}
			return box
		}
	}

	// 默认布局：递归布局子节点（垂直方向），考虑内容区域偏移
	childX := x + contentOffsetX
	childY := y + contentOffsetY
	// 为子节点创建新的约束，使用节点的实际尺寸减去内容偏移
	// 这样 absolute 子节点可以使用正确的内容区域尺寸
	childConstraints := constraints
	if width > 0 && height > 0 {
		// 计算内容区域尺寸（使用 BoxModel 比 2*offset 更准确）
		var contentWidth, contentHeight int
		if boxModelProvider, ok := node.(BoxModelProvider); ok {
			boxModel := boxModelProvider.GetBoxModel()
			contentWidth = boxModel.InnerWidth(width)
			contentHeight = boxModel.InnerHeight(height)
		} else {
			contentWidth = width - 2*contentOffsetX
			contentHeight = height - 2*contentOffsetY
		}
		if contentWidth > 0 && contentHeight > 0 {
			childConstraints = Constraints{
				MinWidth:  0,
				MaxWidth:  min(constraints.MaxWidth, contentWidth),
				MinHeight: 0,
				MaxHeight: min(constraints.MaxHeight, contentHeight),
			}
		}
	}
	for _, child := range node.Children() {
		// 获取子节点的 margin（如果实现了 Marginal 接口）
		marginTop, marginBottom, marginLeft, marginRight := 0, 0, 0, 0
		if marginal, ok := child.(Marginal); ok {
			m := marginal.GetMargin()
			marginTop = m.Top
			marginBottom = m.Bottom
			marginLeft = m.Left
			marginRight = m.Right
		}

		// 应用 margin 偏移
		actualChildX := childX + marginLeft
		actualChildY := childY + marginTop

		// 调整约束，考虑 margin
		adjustedConstraints := childConstraints
		if childConstraints.MaxWidth > 0 {
			adjustedConstraints.MaxWidth = max(0, childConstraints.MaxWidth-marginLeft-marginRight)
		}

		childBox := e.layoutNodeWithDepth(child, adjustedConstraints, actualChildX, actualChildY, depth+1, visited)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
			// 移动下一个子节点的位置，包含当前子节点的高度和垂直 margin
			childY += childBox.Height + marginTop + marginBottom
		}
	}

	return box
}

// collectBoxes 收集所有布局盒子
func (e *Engine) collectBoxes(root *LayoutBox) []LayoutBox {
	if root == nil {
		return nil
	}

	boxes := make([]LayoutBox, 0)
	e.collectBoxesRecursive(root, &boxes)
	return boxes
}

// collectBoxesRecursive 递归收集布局盒子
func (e *Engine) collectBoxesRecursive(box *LayoutBox, boxes *[]LayoutBox) {
	*boxes = append(*boxes, *box)
	for _, child := range box.Children {
		e.collectBoxesRecursive(child, boxes)
	}
}

// layoutNodeIncremental 递归布局单个节点（使用脏标记）
func (e *Engine) layoutNodeIncremental(node Node, constraints Constraints, x, y int) *LayoutBox {
	return e.layoutNodeIncrementalWithDepth(node, constraints, x, y, 0, make(map[string]bool))
}

// layoutNodeIncrementalWithDepth 递归布局单个节点，带深度限制和循环检测
func (e *Engine) layoutNodeIncrementalWithDepth(node Node, constraints Constraints, x, y int, depth int, visited map[string]bool) *LayoutBox {
	if node == nil {
		return nil
	}

	// 深度限制检查
	if depth > MaxLayoutDepth {
		return &LayoutBox{
			ID:      node.ID(),
			X:       x,
			Y:       y,
			AbsX:    x,  // ✨ Phase 1.2: 保存全局坐标
			AbsY:    y,
			Width:   0,
			Height:  0,
			Children: make([]*LayoutBox, 0),
		}
	}

	// 循环检测
	nodeID := node.ID()
	if nodeID != "" {
		if visited[nodeID] {
			return &LayoutBox{
				ID:      node.ID() + "_cycle",
				X:       x,
				Y:       y,
				AbsX:    x,  // ✨ Phase 1.2: 保存全局坐标
				AbsY:    y,
				Width:   0,
				Height:  0,
				Children: make([]*LayoutBox, 0),
			}
		}
		visited[nodeID] = true
		defer delete(visited, nodeID)
	}

	// 检查节点是否是脏的
	if !e.dirty.IsLayoutDirty(node.ID()) {
		// 🔴 BUG FIX: 节点是干净的，但 Fixed 定位不能直接返回 curX, curY
		// 即使节点本身是干净的，Fixed 定位也需要使用 viewportConstraints 重新计算
		
		width, height := node.GetSize()
		curX, curY := node.GetPosition()
		
		// 🐛 获取定位属性（PositionFixed 需要）
		position := PositionRelative
		anchor := AnchorTopLeft
		if posProvider, ok := node.(PositionProvider); ok {
			position = posProvider.GetPositionType()
			anchor = posProvider.GetAnchor()
		}
		
		// ✨ Phase 2.3: Fixed 定位处理（Clean 路径也需要）
		// Fixed 定位必须重新计算，不能只使用 curX, curY
		if position == PositionFixed && width > 0 && height > 0 {
			// 使用保存的 viewport 约束
			rootW := e.viewportConstraints.MaxWidth
			rootH := e.viewportConstraints.MaxHeight
			
			// 根据 Anchor 计算固定定位坐标
			switch anchor {
			case AnchorTopLeft:
				curX, curY = 0, 0
			case AnchorTop:
				curX = (rootW - width) / 2
				curY = 0
			case AnchorTopRight:
				curX = rootW - width
				curY = 0
			case AnchorLeft:
				curX = 0
				curY = (rootH - height) / 2
			case AnchorCenter:
				curX = (rootW - width) / 2
				curY = (rootH - height) / 2
			case AnchorRight:
				curX = rootW - width
				curY = (rootH - height) / 2
			case AnchorBottomLeft:
				curX = 0
				curY = rootH - height
			case AnchorBottom:
				curX = (rootW - width) / 2
				curY = rootH - height
			case AnchorBottomRight:
				curX = rootW - width
				curY = rootH - height
			default:
				curX, curY = 0, 0
			}
		}

		// ✨ Phase 1.3: 设置全局坐标
		absX, absY := curX, curY

		box := &LayoutBox{
			ID:       node.ID(),
			X:        curX,
			Y:        curY,
			AbsX:     absX,  // ✨ Phase 1.2: 保存全局坐标
			AbsY:     absY,
			Width:    width,
			Height:   height,
			ShouldCenter: (anchor == AnchorCenter && position == PositionFixed),  // ✨ 修复 Dirty 状态
			Children: make([]*LayoutBox, 0),
		}

		// 递归处理子节点（仍然检查脏标记）
		childX := curX
		childY := curY
		for _, child := range node.Children() {
			childBox := e.layoutNodeIncrementalWithDepth(child, constraints, childX, childY, depth+1, visited)
			if childBox != nil {
				box.Children = append(box.Children, childBox)
				childY += childBox.Height
			}
		}

		return box
	}

	// 节点是脏的，需要重新布局
	width, height := node.GetSize()

	// 如果节点实现了 Measurable 接口，测量其尺寸
	if measurable, ok := node.(Measurable); ok {
		size := measurable.Measure(constraints)
		width, height = size.Width, size.Height
	}

	// Get Layer from node
	layer := GetLayerFromNode(node)
	shouldCenter := false // ✨ 初始化居中标志（Phase 1.1 + Phase 2.3）

	// ✨ Phase 2.3: PositionFixed 处理 (incremental path)
	// 如果节点使用了 fixed 定位，以 Root 为参考系，不受父布局影响
	position := PositionRelative
	anchor := AnchorTopLeft

	if posProvider, ok := node.(PositionProvider); ok {
		position = posProvider.GetPositionType()
		anchor = posProvider.GetAnchor()
	}

	// Fixed 定位：使用 viewport 约束重新计算坐标
	if position == PositionFixed && width > 0 && height > 0 {
		// ✨ 使用保存的 viewport 约束而不是传入的约束
		// Modal 等组件需要以完整的 viewport 尺寸作为参考系进行定位
		// 而不是受父布局限制后的约束
		rootW := e.viewportConstraints.MaxWidth
		rootH := e.viewportConstraints.MaxHeight

		// 根据 Anchor 计算固定定位坐标
		switch anchor {
		case AnchorTopLeft:
			x, y = 0, 0
		case AnchorTop:
			x = (rootW - width) / 2
			y = 0
		case AnchorTopRight:
			x = rootW - width
			y = 0
		case AnchorLeft:
			x = 0
			y = (rootH - height) / 2
		case AnchorCenter:
			x = (rootW - width) / 2
			y = (rootH - height) / 2
			shouldCenter = true // ✨ 居中定位时设置标志
		case AnchorRight:
			x = rootW - width
			y = (rootH - height) / 2
		case AnchorBottomLeft:
			x = 0
			y = rootH - height
		case AnchorBottom:
			x = (rootW - width) / 2
			y = rootH - height
		case AnchorBottomRight:
			x = rootW - width
			y = rootH - height
		default:
			x, y = 0, 0
		}
	}

	// ✨ Phase 1.1: Modal 居中逻辑 (incremental path)
	// 注意：不适用于 PositionFixed 定位（Phase 2.3 已处理）
	if layer == LayerModal && width > 0 && height > 0 && position != PositionFixed {
		// ✨ Phase 1.4: 优先检查 ModalCenteringProvider 接口
		// Modal 组件通过此接口显式控制是否居中
		if centeringProvider, ok := node.(ModalCenteringProvider); ok {
			shouldCenter = centeringProvider.ShouldCenter()
		} else {
			// 如果没有实现 ModalCenteringProvider，检查 AbsoluteStyleProvider
			// 保持向后兼容性
			if absProvider, ok := node.(AbsoluteStyleProvider); ok {
				absStyle := absProvider.GetAbsoluteStyle()
				// 如果未设置明确的 left/top/right/bottom，则需要居中
				if absStyle != nil {
					shouldCenter = absStyle.ShouldCenter()
				} else {
					shouldCenter = true
				}
			} else {
				// 默认 Modal 居中（向后兼容）
				shouldCenter = true
			}
		}

		// 计算居中坐标
		if shouldCenter {
			x = (constraints.MaxWidth - width) / 2
			y = (constraints.MaxHeight - height) / 2
		}
	}

	// ✨ Phase 1.3: 设置全局坐标
	absX, absY := x, y

	box := &LayoutBox{
		ID:      node.ID(),
		X:       x,
		Y:       y,
		AbsX:    absX,  // ✨ Phase 1.2: 保存全局坐标
		AbsY:    absY,
		Width:   width,
		Height:  height,
		ShouldCenter: shouldCenter,  // ✨ Phase 1.1: 保存居中标记
		Children: make([]*LayoutBox, 0),
	}

	// 设置节点位置和尺寸
	node.SetPosition(x, y)
	node.SetSize(width, height)

	// 递归布局子节点
	childX := x
	childY := y
	for _, child := range node.Children() {
		childBox := e.layoutNodeIncrementalWithDepth(child, constraints, childX, childY, depth+1, visited)
		if childBox != nil {
			box.Children = append(box.Children, childBox)
			childY += childBox.Height
		}
	}

	return box
}

// clearDirtyMarkers 清除节点树的脏标记
func (e *Engine) clearDirtyMarkers(node Node) {
	if node == nil {
		return
	}

	// 清除当前节点的脏标记
	e.dirty.ClearKey(node.ID())

	// 递归清除子节点的脏标记
	for _, child := range node.Children() {
		e.clearDirtyMarkers(child)
	}
}


// GetStats 获取布局统计
func (e *Engine) GetStats() LayoutStats {
	return e.stats
}

// GetOverlayManager returns the overlay manager for handling portal nodes
// Phase 3.3: Portal 跨树挂载支持
func (e *Engine) GetOverlayManager() *OverlayManager {
	return e.overlayManager
}

// Measure 测量节点的尺寸
//
// 测量流程：
// 1. 获取节点的 BoxModel（如果实现了 BoxModelProvider）
// 2. 扣除 Padding 和 Border，创建内容区域的约束
// 3. 测量内容尺寸
// 4. 加回 Padding 和 Border，返回总尺寸
//
// 注意：返回的尺寸包含 Padding 和 Border，但不包含 Margin
// Margin 仅在布局阶段使用，用于计算节点之间的间距
func (e *Engine) Measure(node Node, constraints Constraints) Size {
	if node == nil {
		return Size{}
	}

	// Step 1: 获取 BoxModel 扣除 padding/border
	var boxModel BoxModel

	// 特殊处理：FlexLayout 自己处理内部 padding，Engine 只处理 border
	// 这是因为 FlexLayout 的 style.Padding 是控制子节点位置的内部配置
	// 而不是标准的 box model padding
	if _, ok := node.(*FlexLayout); ok {
		// FlexLayout 只扣除 border
		if provider, ok := node.(Bordered); ok {
			boxModel.Border = provider.GetBorder()
		}
		boxModel.Padding = Padding{} // padding 由 FlexLayout 自己处理
	} else if provider, ok := node.(BoxModelProvider); ok {
		boxModel = provider.GetBoxModel()
	}

	// Step 2: 计算内部约束（扣除 padding/border）
	horizPadding := boxModel.HorizontalPadding()
	vertPadding := boxModel.VerticalPadding()

	minInnerWidth := constraints.MinWidth - horizPadding
	maxInnerWidth := constraints.MaxWidth - horizPadding
	minInnerHeight := constraints.MinHeight - vertPadding
	maxInnerHeight := constraints.MaxHeight - vertPadding

	// 防止负值
	if maxInnerWidth < 0 {
		maxInnerWidth = 0
	}
	if maxInnerHeight < 0 {
		maxInnerHeight = 0
	}
	if minInnerWidth < 0 {
		minInnerWidth = 0
	}
	if minInnerHeight < 0 {
		minInnerHeight = 0
	}
	if maxInnerWidth < minInnerWidth {
		minInnerWidth = maxInnerWidth
	}
	if maxInnerHeight < minInnerHeight {
		minInnerHeight = maxInnerHeight
	}

	innerConstraints := NewConstraints(minInnerWidth, maxInnerWidth, minInnerHeight, maxInnerHeight)

	// Step 3: 测量内容
	var contentSize Size
	if measurable, ok := node.(Measurable); ok {
		contentSize = measurable.Measure(innerConstraints)
	} else {
		// 对于非可测量节点，使用其当前尺寸
		w, h := node.GetSize()
		contentSize.Width = innerConstraints.ConstrainWidth(w)
		contentSize.Height = innerConstraints.ConstrainHeight(h)
	}

	// Step 4: 返回总尺寸（包含 padding/border）
	totalWidth := contentSize.Width + horizPadding
	totalHeight := contentSize.Height + vertPadding

	// 确保不超过约束
	totalWidth = constraints.ConstrainWidth(totalWidth)
	totalHeight = constraints.ConstrainHeight(totalHeight)

	return Size{Width: totalWidth, Height: totalHeight}
}
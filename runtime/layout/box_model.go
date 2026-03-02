package layout

// BoxModel 统一表示节点的盒模型属性（margin, padding, border）
// 该结构体用于统一处理节点的空间占用信息
type BoxModel struct {
	// Margin 外边距（节点与兄弟节点之间的间距）
	Margin Margin

	// Padding 内边距（内容与边框之间的间距）
	Padding Padding

	// Border 边框（视觉边界）
	Border Border
}

// BoxModelProvider 提供盒模型信息的接口
// 实现此接口的节点将自动由布局引擎处理 padding/border 的约束传播
type BoxModelProvider interface {
	Node
	GetBoxModel() BoxModel
}

// HorizontalPadding 返回水平方向占用的总空间
// 包括左边框 + 左padding + 右padding + 右边框
func (b BoxModel) HorizontalPadding() int {
	return b.Border.HorizontalPadding() +
		b.Padding.Left +
		b.Padding.Right
}

// VerticalPadding 返回垂直方向占用的总空间
// 包括上边框 + 上padding + 下padding + 下边框
func (b BoxModel) VerticalPadding() int {
	return b.Border.VerticalPadding() +
		b.Padding.Top +
		b.Padding.Bottom
}

// ContentOffsetX 返回内容区域的 X 偏移
// 相对于容器左边缘的距离
func (b BoxModel) ContentOffsetX() int {
	return b.Padding.Left + b.Border.HorizontalPadding()/2
}

// ContentOffsetY 返回内容区域的 Y 偏移
// 相对于容器上边缘的距离
func (b BoxModel) ContentOffsetY() int {
	return b.Padding.Top + b.Border.VerticalPadding()/2
}

// TotalWidth 计算总宽度（包含 padding 和 border）
func (b BoxModel) TotalWidth(contentWidth int) int {
	return contentWidth + b.HorizontalPadding()
}

// TotalHeight 计算总高度（包含 padding 和 border）
func (b BoxModel) TotalHeight(contentHeight int) int {
	return contentHeight + b.VerticalPadding()
}

// InnerWidth 计算内部可用宽度
// 给定总宽度，返回内容区域可用宽度
func (b BoxModel) InnerWidth(totalWidth int) int {
	innerW := totalWidth - b.HorizontalPadding()
	if innerW < 0 {
		innerW = 0
	}
	return innerW
}

// InnerHeight 计算内部可用高度
// 给定总高度，返回内容区域可用高度
func (b BoxModel) InnerHeight(totalHeight int) int {
	innerH := totalHeight - b.VerticalPadding()
	if innerH < 0 {
		innerH = 0
	}
	return innerH
}

// IsEmpty 检查盒模型是否为空（所有属性均为零值）
func (b BoxModel) IsEmpty() bool {
	return b.Margin.Horizontal() == 0 && b.Margin.Vertical() == 0 &&
		b.Padding.Left == 0 && b.Padding.Right == 0 &&
		b.Padding.Top == 0 && b.Padding.Bottom == 0 &&
		!b.Border.HasBorder()
}

// Clone 创建 BoxModel 的副本
func (b BoxModel) Clone() BoxModel {
	return BoxModel{
		Margin:  b.Margin,
		Padding: b.Padding,
		Border:  b.Border,
	}
}

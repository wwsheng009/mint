package layout

// ContentBox 返回内容区域（去除 margin 后的区域）
func (mb *MarginBox) ContentBox() *LayoutBox {
	if mb.Box == nil {
		return nil
	}
	return &LayoutBox{
		ID:      mb.Box.ID,
		X:       mb.Box.X + mb.Margin.Left,
		Y:       mb.Box.Y + mb.Margin.Top,
		Width:   mb.Box.Width - mb.Margin.Horizontal(),
		Height:  mb.Box.Height - mb.Margin.Vertical(),
	}
}

// BorderBox 返回边框盒（包含 margin 的完整区域）
func (mb *MarginBox) BorderBox() *LayoutBox {
	return mb.Box
}

package layout

// NewConstraints 创建约束，并确保 Min <= Max 且值非负
func NewConstraints(minWidth, maxWidth, minHeight, maxHeight int) Constraints {
	if minWidth < 0 {
		minWidth = 0
	}
	if minHeight < 0 {
		minHeight = 0
	}
	if maxWidth < 0 {
		maxWidth = 0
	}
	if maxHeight < 0 {
		maxHeight = 0
	}
	if maxWidth < minWidth {
		minWidth = maxWidth
	}
	if maxHeight < minHeight {
		minHeight = maxHeight
	}
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// Tight 创建紧约束（固定尺寸）
func TightConstraints(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Loose 创建松约束（只有最小值）
func LooseConstraints(minWidth, minHeight int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  MaxInt,
		MinHeight: minHeight,
		MaxHeight: MaxInt,
	}
}

// Unbounded 创建无界约束
func UnboundedConstraints() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  MaxInt,
		MinHeight: 0,
		MaxHeight: MaxInt,
	}
}

// Width 创建宽度约束
func (c Constraints) Width(minWidth, maxWidth int) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
}

// Height 创建高度约束
func (c Constraints) Height(minHeight, maxHeight int) Constraints {
	return Constraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// IsTight 检查是否为紧约束
func (c Constraints) IsTight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// IsBounded 检查是否有界
func (c Constraints) IsBounded() bool {
	return c.MaxWidth < MaxInt || c.MaxHeight < MaxInt
}

// Constrain 约束尺寸到范围内
func (c Constraints) Constrain(width, height int) (int, int) {
	if width < c.MinWidth {
		width = c.MinWidth
	}
	if width > c.MaxWidth {
		width = c.MaxWidth
	}
	if height < c.MinHeight {
		height = c.MinHeight
	}
	if height > c.MaxHeight {
		height = c.MaxHeight
	}
	return width, height
}

// ConstrainWidth 约束宽度
// MaxWidth=0 或 MaxWidth=MaxInt 表示无界（自动宽度）
func (c Constraints) ConstrainWidth(width int) int {
	if width < c.MinWidth {
		return c.MinWidth
	}
	// ✨ FIX: MaxWidth=0 或 MaxWidth=MaxInt 表示无界，不进行上限约束
	if c.MaxWidth > 0 && c.MaxWidth < MaxInt && width > c.MaxWidth {
		return c.MaxWidth
	}
	return width
}

// ConstrainHeight 约束高度
// MaxHeight=0 或 MaxHeight=MaxInt 表示无界（自动高度）
func (c Constraints) ConstrainHeight(height int) int {
	if height < c.MinHeight {
		return c.MinHeight
	}
	// ✨ FIX: MaxHeight=0 或 MaxHeight=MaxInt 表示无界，不进行上限约束
	if c.MaxHeight > 0 && c.MaxHeight < MaxInt && height > c.MaxHeight {
		return c.MaxHeight
	}
	return height
}

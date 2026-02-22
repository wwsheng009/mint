## Qwen Added Memories
- Mint 布局盒模型分隔线修复完成：分隔线长度基于子元素 contentWidth 而非 boxWidth，确保分隔线右端 ┤ 与最宽子元素右边框 │ 对齐且无溢出，左右对称 `│  ├─...─┤  │`。已通过 test_separator_fix.go 测试验证。
- Mint 布局盒模型分隔线精确对齐修复：分隔线长度 = maxContentWidth（所有子元素的内容宽度-2）。分隔线位置：左端 `├` = childX+1，右端 `┤` = childX+maxContentWidth。确保 `┤` 位于最宽子元素内容区域内，不触及右边框。模式：`│  ├─...─┤  │`(左右各 1 空格)。已通过 tree_test.go 15项测试。
- Mint 布局盒模型分隔线已移除：取消子元素分隔线（├─...─┤），子元素直接垂直堆叠，简化了布局可视化的复杂度。

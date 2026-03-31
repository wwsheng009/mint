# Chart Canvas

计划中的低层字符画布。

目标：

- 统一 Braille / Block / ASCII 三种渲染模式
- 提供逻辑像素到终端 cell 的映射
- 屏蔽不同图表组件对底层字符细节的直接耦合

后续建议实现内容：

- 逻辑像素缓冲
- cell 聚合
- glyph 选择
- 颜色合成与前景样式输出


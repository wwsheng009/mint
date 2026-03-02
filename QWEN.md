## Qwen Added Memories
- Portal设计核心原则：Portal = Fiber 不动，Layout 重建。两阶段Layout：主树忽略Portal并收集，Overlay阶段独立计算（Root坐标系）。Portal本质是Fiber层语义（PortalRoot）+ Layout层执行策略（坐标系重定向）
- Panel 组件已成功迁移到原生边框方案（不使用 border.New()），所有 Go 代码中不再有 border.New() 的实际调用（仅 border 包内注释中有示例）。Stack/Grid/Wrap 已添加 borderColor 完整支持。
- ui.Bordered() 已添加 Deprecated 标记，推荐使用 stack.Bordered() 或 stack.New().SingleBorder() 作为替代。
- ui.Bordered() 迁移进度：已完成 Inspector 内部代码全部迁移（5个文件），开始迁移 Demo 代码（已完成 demo1_full_featured/main.go），剩余 demo2 等文件待处理。Stack 已添加 Bordered() / BorderedWithLabel() 等便捷构造函数。
- ui.Bordered() 迁移进度更新：已完成 Inspector 内部代码（5文件），demo1_full_featured（4处），demo2_runtime_internals/main.go（5处），demo2/inspector_standalone（5处）。剩余：demo2/inspector_overlay（5处），demo2/inspector_demo（约10处），以及 examples/sandbox/dump_buffer/ 大量测试文件。
- ui.Bordered() 迁移进度更新：已完成 Inspector 内部代码（5个文件约14处），demo1_full_featured（4处），demo2_runtime_internals 全部子目录（main.go: 5处，inspector_standalone: 5处，inspector_overlay: 5处，inspector_demo: 6处，inspector_demo/simple: 5处）。剩余：examples/sandbox/dump_buffer/ 约54处（测试文件），examples/sandbox/demo/modal_center_test.go 2处。
- ui.Bordered() 迁移全部完成！总计迁移约85+处调用：
- Inspector 内部代码（5个文件，约14处）
- demo1_full_featured（4处）
- demo2_runtime_internals 全部子目录（main.go 5处，inspector_standalone 5处，inspector_overlay 5处，inspector_demo 6处，inspector_demo/simple 5处）
- examples/sandbox/modal_center_test.go（2处）
- examples/sandbox/dump_buffer/（51处：border_showcase_test.go 21处，border_output_test.go 25处，bordered_test.go 5处）
- examples/component_fixtures/fixtures.go（5处）
- examples/ant_design_demo/main.go 和 main_test.go（5处）
- examples/fiber_demos/demo1_full_featured/main.go（4处）

同时已完成：
- Panel 组件移除对 ui/components/border 包的依赖
- ui.Bordered() 已标记 Deprecated，推荐使用 stack.Bordered() 或 stack.New().SingleBorder()

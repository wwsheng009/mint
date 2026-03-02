## Qwen Added Memories
- Portal设计核心原则：Portal = Fiber 不动，Layout 重建。两阶段Layout：主树忽略Portal并收集，Overlay阶段独立计算（Root坐标系）。Portal本质是Fiber层语义（PortalRoot）+ Layout层执行策略（坐标系重定向）

# Control

交互行为基础包，给按钮、输入框、选择器等组件复用统一的状态和行为模型。

## 已支持

- `InteractionState`
- `FocusableBehavior`
- `HoverableBehavior`
- `PressableBehavior`
- `DisableableBehavior`
- `ActivatableBehavior`
- `BehaviorList`
- `StayPressedIntent`

## 说明

这个包主要面向组件实现层，不是常规业务 UI 直接调用的入口。它的职责是把 `focus / hover / press / disable / active` 这些交互状态统一起来，减少各个组件各自维护一套状态机。

如果你在新增组件，优先复用这里的 behavior，而不是再写一份新的交互状态逻辑。

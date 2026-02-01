// 05_injection_strategy/demo_test.go
// 事件注入策略演示测试

package main

import (
	"errors"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/mock"
	"github.com/wwsheng009/mint/ui"
)

// TestInjectAllowed 演示允许注入策略
func TestInjectAllowed(t *testing.T) {
	t.Log("=== 允许注入策略演示 ===")

	// 创建使用允许注入策略的沙箱
	sb := mock.New(40, 18)
	sb.Initialize(nil)
	defer sb.Close()

	// 确认策略为允许注入
	t.Logf("默认策略: %v", sb.Injector().Strategy())

	// 注入应该成功
	err := sb.InjectKey('a')
	if err != nil {
		t.Errorf("注入失败: %v", err)
	} else {
		t.Log("✅ 注入成功")
	}

	err = sb.InjectSpecialKey(platform.KeyEnter)
	if err != nil {
		t.Errorf("注入失败: %v", err)
	} else {
		t.Log("✅ 特殊键注入成功")
	}
}

// TestInjectProhibited 演示禁止注入策略
func TestInjectProhibited(t *testing.T) {
	t.Log("=== 禁止注入策略演示 ===")

	// 创建沙箱并设置禁止注入策略
	sb := mock.New(40, 18)
	sb.Initialize(nil)
	defer sb.Close()

	// 设置禁止注入策略
	sb.Injector().SetStrategy(sandbox.InjectProhibited)

	// 注入应该被拒绝
	err := sb.InjectKey('a')
	if err != nil {
		t.Logf("✅ 注入被拒绝: %v", err)
	} else {
		t.Error("注入应该被拒绝")
	}

	// 但事件仍然会被录制
	recorder := sb.Recorder()
	if recorder != nil && recorder.Len() > 0 {
		t.Log("✅ 事件被录制")
	}
}

// TestInjectRecorded 演示仅录制策略
func TestInjectRecorded(t *testing.T) {
	t.Log("=== 仅录制策略演示 ===")

	sb := mock.New(40, 18)
	sb.Initialize(nil)
	defer sb.Close()

	// 设置仅录制策略
	sb.Injector().SetStrategy(sandbox.InjectRecorded)

	// 注入会被录制但不会发送到处理器
	recorder := sb.Recorder()
	initialCount := recorder.Len()

	sb.InjectKey('a')
	sb.InjectSpecialKey(platform.KeyEnter)

	time.Sleep(50 * time.Millisecond)

	finalCount := recorder.Len()
	t.Logf("录制事件数: %d", finalCount-initialCount)

	if finalCount > initialCount {
		t.Log("✅ 事件已录制")
	}
}

// TestStrategySwitch 演示动态切换策略
func TestStrategySwitch(t *testing.T) {
	t.Log("=== 动态切换策略演示 ===")

	sb := mock.New(40, 18)
	sb.Initialize(nil)
	defer sb.Close()

	injector := sb.Injector()

	// 开始：允许注入
	t.Log("\n--- 策略: Allowed ---")
	injector.SetStrategy(sandbox.InjectAllowed)
	err := sb.InjectKey('a')
	if err == nil {
		t.Log("✅ 注入成功")
	}

	// 切换：禁止注入
	t.Log("\n--- 策略: Prohibited ---")
	injector.SetStrategy(sandbox.InjectProhibited)
	err = sb.InjectKey('b')
	if err != nil {
		t.Logf("✅ 注入被拒绝: %v", err)
	}

	// 切换回允许注入
	t.Log("\n--- 策略: Allowed (恢复) ---")
	injector.SetStrategy(sandbox.InjectAllowed)
	err = sb.InjectKey('c')
	if err == nil {
		t.Log("✅ 注入成功")
	}
}

// TestStrategyWithApp 演示策略与应用交互
func TestStrategyWithApp(t *testing.T) {
	t.Log("=== 策略与应用交互演示 ===")

	// 创建测试应用
	testApp, err := ui.RunTestWithSandbox(StrategyApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 使用 Allowed 策略，事件应该正常处理
	t.Log("\n--- 使用 Allowed 策略 ---")
	sb.Helper().
		Tab().       // 切换到 + 按钮
		Press(platform.KeyEnter).  // 点击
		Process()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	if contains(rendered, "Count: 1") {
		t.Log("✅ 事件处理成功，计数器增加")
	}
}

// TestInjectErrorHandling 演示注入错误处理
func TestInjectErrorHandling(t *testing.T) {
	t.Log("=== 注入错误处理演示 ===")

	sb := mock.New(40, 18)
	sb.Initialize(nil)
	defer sb.Close()

	// 设置禁止注入策略
	sb.Injector().SetStrategy(sandbox.InjectProhibited)

	// 尝试各种注入操作
	operations := []struct {
		name string
		fn   func() error
	}{
		{"InjectKey", func() error {
			return sb.InjectKey('a')
		}},
		{"InjectSpecialKey", func() error {
			return sb.InjectSpecialKey(platform.KeyEnter)
		}},
		{"InjectString", func() error {
			return sb.InjectString("test")
		}},
		{"InjectMouse", func() error {
			return sb.InjectMouse(10, 10, platform.MouseLeft, platform.MousePress)
		}},
	}

	t.Log("\n--- 测试各种注入操作 ---")
	for _, op := range operations {
		err := op.fn()
		if err != nil {
			if errors.Is(err, sandbox.ErrInjectionNotAllowed) {
				t.Logf("%s: ✅ 正确返回 ErrInjectionNotAllowed", op.name)
			} else {
				t.Logf("%s: 返回错误 %v", op.name, err)
			}
		} else {
			t.Errorf("%s: 应该返回错误", op.name)
		}
	}
}

// TestRecordingWithDifferentStrategies 演示不同策略下的录制
func TestRecordingWithDifferentStrategies(t *testing.T) {
	t.Log("=== 不同策略下的录制演示 ===")

	strategies := []struct {
		name     string
		strategy sandbox.InjectionStrategy
	}{
		{"Allowed", sandbox.InjectAllowed},
		{"Prohibited", sandbox.InjectProhibited},
		{"Recorded", sandbox.InjectRecorded},
	}

	for _, tc := range strategies {
		t.Run(tc.name, func(t *testing.T) {
			sb := mock.New(40, 18)
			sb.Initialize(nil)
			defer sb.Close()

			sb.Injector().SetStrategy(tc.strategy)

			// 创建新的录制器
			recorder := sandbox.NewEventRecorder(100)
			sb.SetRecorder(recorder)

			// 执行一些注入操作
			sb.InjectKey('a')
			sb.InjectKey('b')
			sb.InjectSpecialKey(platform.KeyEnter)

			// 检查录制结果
			recorded := recorder.Len()
			t.Logf("%s 策略: 录制了 %d 个事件", tc.name, recorded)

			if recorded > 0 {
				t.Log("✅ 事件已录制")
			}
		})
	}
}

// TestStrategyInProduction 演示生产环境策略设置
func TestStrategyInProduction(t *testing.T) {
	t.Log("=== 生产环境策略配置演示 ===")

	// 模拟生产环境配置
	t.Log("\n--- 生产环境配置 ---")
	prodConfig := &sandbox.Config{
		Width:  80,
		Height: 24,
		Event: sandbox.EventConfig{
			Strategy:     sandbox.InjectProhibited,
			QueueMaxSize: 10000,
			RecordEnabled: true, // 生产环境可以录制用于调试
		},
	}

	sb := mock.NewWithConfig(prodConfig)
	sb.Initialize(nil)
	defer sb.Close()

	// 验证策略
	if sb.Injector().Strategy() == sandbox.InjectProhibited {
		t.Log("✅ 生产环境策略设置正确: Prohibited")
	}

	// 验证注入被拒绝
	err := sb.InjectKey('a')
	if err != nil {
		t.Logf("✅ 注入被正确拒绝: %v", err)
	}

	// 但录制应该工作
	recorder := sb.Recorder()
	if recorder != nil {
		t.Logf("✅ 录制器已配置，容量: %d", recorder.Len())
	}
}

// 辅助函数
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

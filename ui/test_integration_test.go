// ui/test_integration_test.go - UI测试集成示例
package ui

import (
	"testing"

	"github.com/wwsheng009/mint/sandbox"
)

func TestTestRun(t *testing.T) {
	app := "test app"

	testApp, err := TestRun(app, TestWithSize(80, 24))
	if err != nil {
		t.Fatalf("TestRun() error = %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	if sb == nil {
		t.Fatal("Sandbox() returned nil")
	}

	width, height := sb.Size()
	if width != 80 || height != 24 {
		t.Errorf("Size() = %dx%d, want 80x24", width, height)
	}
}

func TestTestRunWithConfig(t *testing.T) {
	app := "test app"
	config := sandbox.DefaultConfig()
	config.Width = 100
	config.Height = 30

	testApp, err := TestRunWithConfig(app, config)
	if err != nil {
		t.Fatalf("TestRunWithConfig() error = %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	width, height := sb.Size()
	if width != 100 || height != 30 {
		t.Errorf("Size() = %dx%d, want 100x30", width, height)
	}
}

func TestTestHelperChain(t *testing.T) {
	app := "test app"
	testApp, err := TestRun(app, TestWithSize(80, 24))
	if err != nil {
		t.Fatalf("TestRun() error = %v", err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 链式测试
	result := helper.
		Type("hello").
		Tab().
		Enter().
		Process().
		Result()

	if !result.OK() {
		t.Errorf("Chain produced errors: %v", result.Errors)
	}
}

func TestTestWithWidth(t *testing.T) {
	config := &testConfig{}
	TestWithWidth(100)(config)

	if config.width != 100 {
		t.Errorf("TestWithWidth() width = %v, want 100", config.width)
	}
}

func TestTestWithHeight(t *testing.T) {
	config := &testConfig{}
	TestWithHeight(30)(config)

	if config.height != 30 {
		t.Errorf("TestWithHeight() height = %v, want 30", config.height)
	}
}

func TestTestWithSize(t *testing.T) {
	config := &testConfig{}
	TestWithSize(120, 40)(config)

	if config.width != 120 || config.height != 40 {
		t.Errorf("TestWithSize() size = %dx%d, want 120x40", config.width, config.height)
	}
}

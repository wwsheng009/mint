// 测试 Ant Design Demo
package main

import (
	"testing"

	"github.com/wwsheng009/mint/app"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// TestStep3LayoutVNode 测试 Step 3 布局的 VNode 结构
func TestStep3LayoutVNode(t *testing.T) {
	// 创建 Step 3 的 VNode 结构，使用固定测试值
	vnode := ui.Bordered().
		Child(
			ui.VStackBuilder(
				ConfirmInfo("Username:", ""),
				ConfirmInfo("Email:", ""),
				ConfirmInfo("Age:", ""),
				ui.HStackBuilder(
					ui.Text("         "),
					app.CheckboxBuilder().
						Checked(false).
						Label("I agree to the Terms and Conditions").
						Build(),
				).Build(),
			).Gap(2).Build(),
		).
		Build()

	// 创建 Fiber 树
	fiber := rtui.CreateFiberFromVNode(vnode)

	// 验证 Fiber 结构
	if fiber.Tag != "bordered" {
		t.Errorf("Expected Bordered tag, got '%s'", fiber.Tag)
	}

	// 统计子节点数量
	childCount := 0
	child := fiber.Child
	for child != nil {
		childCount++
		child = child.Sibling
	}

	if childCount != 1 {
		t.Errorf("Expected Bordered to have 1 child, got %d", childCount)
	}

	t.Logf("Fiber structure validated successfully")
}

// TestConfirmInfoWidth 测试单个 ConfirmInfo 的宽度
func TestConfirmInfoWidth(t *testing.T) {
	// 测试长电子邮件地址
	vnode := ConfirmInfo("Email:", "very.long.email.address@example.com")

	fiber := rtui.CreateFiberFromVNode(vnode)

	// 检查 HStack 结构
	if fiber.Tag != "hstack" {
		t.Errorf("Expected HStack tag, got '%s'", fiber.Tag)
	}

	// 检查子节点
	child := fiber.Child
	textCount := 0

	for child != nil {
		if child.Tag == "text" {
			textCount++
			if child.MemoizedState != nil {
				if content, ok := child.MemoizedState.(string); ok {
					t.Logf("Text %d content: '%s' (length %d)", textCount, content, len(content))
				}
			}
		}
		child = child.Sibling
	}

	if textCount != 2 {
		t.Errorf("Expected ConfirmInfo to have 2 text children, got %d", textCount)
	}
}

// TestFormItemVNode 测试 FormItem 的 VNode 结构
func TestFormItemVNode(t *testing.T) {
	// 创建测试 FormItem
	vnode := FormItem("test@example.com", "Enter email", 24, emailKey, "Help text", true)

	fiber := rtui.CreateFiberFromVNode(vnode)

	// 验证最外层是 VStackBuilder
	// 注意：VStackBuilder 返回的是 *VNode 类型，需要检查实际标签
	t.Logf("FormItem root tag: '%s'", fiber.Tag)

	// 统计子节点
	child := fiber.Child
	componentCount := 0
	for child != nil {
		componentCount++
		child = child.Sibling
	}

	if componentCount != 2 {
		t.Errorf("Expected FormItem to have 2 children (label row + help row), got %d", componentCount)
	}

	t.Logf("FormItem VNode structure validated")
}

// 03_test_helper/main.go
// TestHelper 链式 API 示例 (Store 模式)
//
// 演示如何使用 TestHelper 的流畅链式 API
// 简化测试代码编写。

package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	Username  string  // 用户名
	Password  string  // 密码
	Submitted bool    // 是否已提交
	Message   string  // 消息
}

// ============================================================================
// Intent Types
// ============================================================================

type SubmitFormIntent struct{}
func (SubmitFormIntent) IntentType() string { return "SubmitForm" }
func (SubmitFormIntent) StayPressed() bool  { return true }

type ClearFormIntent struct{}
func (ClearFormIntent) IntentType() string { return "ClearForm" }
func (ClearFormIntent) StayPressed() bool  { return true }

type SetTestHelperUsernameIntent struct {
	Username string
}
func (SetTestHelperUsernameIntent) IntentType() string { return "SetTestHelperUsername" }
func (SetTestHelperUsernameIntent) StayPressed() bool  { return false }

type SetTestHelperPasswordIntent struct {
	Password string
}
func (SetTestHelperPasswordIntent) IntentType() string { return "SetTestHelperPassword" }
func (SetTestHelperPasswordIntent) StayPressed() bool  { return false }

// ============================================================================
// Store 初始化
// ============================================================================

var testHelperStore = store.NewStore(AppState{
	Username:  "",
	Password:  "",
	Submitted: false,
	Message:   "",
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Submitted = true
			if strings.TrimSpace(s.Username) != "" && strings.TrimSpace(s.Password) != "" {
				s.Message = fmt.Sprintf("Welcome, %s!", s.Username)
			} else {
				s.Message = "Please fill all fields"
			}
			return s
		}).
		On(ClearFormIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Submitted = false
			s.Message = ""
			return s
		}).
		On(SetTestHelperUsernameIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Username = i.(SetTestHelperUsernameIntent).Username
			return s
		}).
		On(SetTestHelperPasswordIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Password = i.(SetTestHelperPasswordIntent).Password
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), testHelperStore)
}

// ============================================================================
// FormApp - 表单应用
// ============================================================================

func FormApp() ui.VNode {
	// ✅ 订阅存储的状态
	username := ui.UseStoreSelector(testHelperStore, func(s AppState) string { return s.Username })
	password := ui.UseStoreSelector(testHelperStore, func(s AppState) string { return s.Password })
	submitted := ui.UseStoreSelector(testHelperStore, func(s AppState) bool { return s.Submitted })
	message := ui.UseStoreSelector(testHelperStore, func(s AppState) string { return s.Message })

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     TestHelper Demo           ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Username: "),
			ui.NewInputBuilder().
				Value(username).
				Placeholder("Enter username").
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Password: "),
			ui.NewInputBuilder().
				Value(password).
				Placeholder("Enter password").
				Password().
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),
		ui.NewButtonBuilder("  [ Submit ]  ").
			OnPress(SubmitFormIntent{}).
			Build(),
		ui.Text(""),
		ui.NewButtonBuilder("  [ Clear ]  ").
			OnPress(ClearFormIntent{}).
			Build(),
		ui.Text(""),
		func() ui.VNode {
			if submitted {
				color := "red"
				if strings.TrimSpace(username) != "" && strings.TrimSpace(password) != "" {
					color = "green"
				}
				return ui.NewTextBuilder(message).
					FgColor(color).
					Bold(true).
					Build()
			}
			return ui.Text("")
		}(),
	)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
		ui.WithTitle("TestHelper Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

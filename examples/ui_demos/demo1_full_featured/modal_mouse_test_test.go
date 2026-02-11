// Modal Mouse Click Test
// 使用 TestableApp 模拟点击 modal 按钮，收集 HitTest 数据
package main

import (
	"os"
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	runtimeevent "github.com/wwsheng009/mint/runtime/event"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestModalMouseClick 测试 modal 按钮的 HitTest
func TestModalMouseClick(t *testing.T) {
	// 启用详细日志
	os.Setenv("TUI_DEBUG_LOG", "modal_mouse_test.log")

	// 初始化 logger
	log.UILogger.Debug("=== Modal Mouse Click Test Started ===")

	// 确保主题加载
	_ = theme.SetTheme("nord")

	// 创建测试应用 - modal 自动打开
	testApp, err := ui.RunTest(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Modal Mouse Test"),
	)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Close()

	log.Info("TEST", "TestableApp created and initialized")

	// 强制进行一次渲染以确保 HitMap 被设置
	fwApp := testApp.GetFrameworkApp()
	fwApp.ForceRenderNow()

	// 获取 HitMap
	hitMap := fwApp.GetHitMap()

	if hitMap == nil {
		// 如果还是 nil，打印调试信息
		log.Info("TEST", "ERROR: HitMap is nil after ForceRenderNow")
		log.Info("TEST", "Framework App State: %v", fwApp.GetState())
		t.Fatal("ERROR: HitMap is nil after render")
	}

	log.Info("TEST", "=== HitMap Info ===")
	log.Info("TEST", "HitMap Size: %d entries", hitMap.Size())

	// 收集所有 modal 按钮的 Bounds
	log.Info("TEST", "=== Collecting Modal Button Bounds ===")
	var modalButtons []ButtonInfo
	allEntries := hitMap.AllEntries()
	for _, entry := range allEntries {
		if entry.NodeID == "button" {
			info := ButtonInfo{
				ID:     entry.NodeID,
				X:      entry.Bounds.X,
				Y:      entry.Bounds.Y,
				Width:  entry.Bounds.Width,
				Height: entry.Bounds.Height,
			}
			modalButtons = append(modalButtons, info)
			log.Info("TEST", "Found button: Bounds=(%d,%d,%dx%d)",
				info.X, info.Y, info.Width, info.Height)

			// 打印按钮的中心点（用于点击测试）
			centerX := info.X + info.Width/2
			centerY := info.Y + info.Height/2
			log.Info("TEST", "  Center point for click: (%d, %d)", centerX, centerY)
		}
	}

	if len(modalButtons) == 0 {
		log.Info("TEST", "WARNING: No modal buttons found!")
	} else {
		log.Info("TEST", "Found %d modal button(s)", len(modalButtons))
	}

	// 对每个按钮执行点击测试
	log.Info("TEST", "")
	log.Info("TEST", "=== Performing Click Tests ===")

	for i, btn := range modalButtons {
		log.Info("TEST", "")
		log.Info("TEST", "--- Test Button %d ---", i+1)

		// 测试点击按钮的左上角
		testClick(t, testApp, hitMap, btn.X, btn.Y, "Top-Left")

		// 测试点击按钮的中心
		centerX := btn.X + btn.Width/2
		centerY := btn.Y + btn.Height/2
		testClick(t, testApp, hitMap, centerX, centerY, "Center")

		// 测试点击按钮的右下角
		bottomRightX := btn.X + btn.Width - 1
		bottomRightY := btn.Y + btn.Height - 1
		testClick(t, testApp, hitMap, bottomRightX, bottomRightY, "Bottom-Right")

		// 测试点击按钮外部（左侧，应该不命中）
		testClick(t, testApp, hitMap, btn.X-1, btn.Y, "Left-Outside (should miss)")

		// 测试点击按钮外部（右侧，应该不命中）
		testClick(t, testApp, hitMap, btn.X+btn.Width, btn.Y, "Right-Outside (should miss)")
	}

	// 收集并打印所有 HitMap 条目用于完整分析
	log.Info("TEST", "")
	log.Info("TEST", "=== Complete HitMap Dump ===")
	allEntries = hitMap.AllEntries()
	for i, entry := range allEntries {
		log.Info("TEST", "[%d] ID='%s' Bounds=(%d,%d,%dx%d) ZOrder=%d",
			i, entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height, entry.ZOrder)
	}

	// 打印 buffer 内容用于验证视觉位置
	log.Info("TEST", "")
	log.Info("TEST", "=== Buffer Content (Modal Region) ===")
	buffer := testApp.GetBuffer()

	// 找出 modal 所在的区域
	modalY := 0
	modalHeight := 0
	for _, entry := range allEntries {
		if entry.NodeID == "bordered" && entry.Bounds.Height > 5 {
			// 这应该是 modal 的 bordered 容器
			modalY = entry.Bounds.Y
			modalHeight = entry.Bounds.Height
			break
		}
	}

	// 打印 modal 区域的内容
	endY := modalY + modalHeight
	if endY > buffer.Height {
		endY = buffer.Height
	}

	for y := modalY; y < endY; y++ {
		line := ""
		for x := 0; x < buffer.Width && x < 80; x++ {
			cell := buffer.Cells[y][x]
			if cell.Cluster != "" {
				// 获取第一个字符
				for _, r := range cell.Cluster {
					line += string(r)
					break
				}
			} else {
				line += " "
			}
		}
		log.Info("TEST", "Buffer[%d]: %q", y, line)
	}

	log.Info("TEST", "")
	log.Info("TEST", "=== Test Complete ===")
	log.Info("TEST", "Log file: modal_mouse_test.log")
}

func testClick(t *testing.T, testApp *ui.TestableApp, hitMap *runtimeevent.HitMap, x, y int, label string) {
	log := logger.Get()
	log.Info("TEST", "Click at (%d,%d) - %s", x, y, label)

	// 执行 HitTest
	entry := hitMap.HitTest(x, y)

	if entry != nil {
		log.Info("TEST", "  -> HIT: ID='%s' Bounds=(%d,%d,%dx%d)",
			entry.NodeID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height)

		// 模拟鼠标点击事件
		err := testApp.InjectMouse(x, y, platform.MouseLeft, platform.MousePress)
		if err != nil {
			log.Info("TEST", "  -> Inject ERROR: %v", err)
		} else {
			log.Info("TEST", "  -> Inject SUCCESS")
		}

		// 检查是否有多个重叠的条目
		allEntries := hitMap.FindAllAt(x, y)
		if len(allEntries) > 1 {
			log.Info("TEST", "  -> WARNING: %d entries at this position!", len(allEntries))
			for j, e := range allEntries {
				log.Info("TEST", "     [%d] ID='%s' Bounds=(%d,%d,%dx%d) ZOrder=%d",
					j, e.NodeID, e.Bounds.X, e.Bounds.Y,
					e.Bounds.Width, e.Bounds.Height, e.ZOrder)
			}
		}
	} else {
		log.Info("TEST", "  -> MISS: No hit at (%d,%d)", x, y)
	}
}

type ButtonInfo struct {
	ID     string
	X      int
	Y      int
	Width  int
	Height int
}

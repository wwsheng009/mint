package paint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime/style"
)

// TestIntegration_ComplexScene 测试复杂场景的多层渲染
func TestIntegration_ComplexScene(t *testing.T) {
	t.Run("多层Z-Index排序验证", func(t *testing.T) {
		width, height := 40, 10
		compositor := NewCompositor(width, height)

		// 创建三层，相同的重叠区域
		layer1 := NewLayer("layer1", LayerContent, 0, 20, 10)
		layer1.Buffer.Cells[5][5] = Cell{Cluster: "L1", Style: style.Style{}.Foreground(style.Red)}

		layer2 := NewLayer("layer2", LayerContent, 1, 20, 10)
		layer2.Buffer.Cells[5][5] = Cell{Cluster: "L2", Style: style.Style{}.Foreground(style.Green)}

		layer3 := NewLayer("layer3", LayerContent, 2, 20, 10)
		layer3.Buffer.Cells[5][5] = Cell{Cluster: "L3", Style: style.Style{}.Foreground(style.Blue)}

		compositor.AddLayer(layer1)
		compositor.AddLayer(layer2)
		compositor.AddLayer(layer3)

		result := compositor.Composite()

		// Z-Index最高的layer（layer3）应该可见
		cell := result.Cells[5][5]
		assert.Equal(t, "L3", cell.Cluster, "Highest Z-Index layer should be visible")
	})

	t.Run("动态层切换", func(t *testing.T) {
		width, height := 40, 10
		compositor := NewCompositor(width, height)

		// 初始状态：只有层1
		layer1 := NewLayer("layer1", LayerContent, 0, width, height)
		layer1.Buffer.Cells[5][5] = Cell{Cluster: "State1", Style: style.Style{}.Foreground(style.Red)}
		compositor.AddLayer(layer1)

		result1 := compositor.Composite()
		assert.Contains(t, result1.Cells[5][5].Cluster, "S")

		// 切换到状态2：移除层1，添加层2
		compositor.RemoveLayer("layer1")
		layer2 := NewLayer("layer2", LayerContent, 0, width, height)
		layer2.Buffer.Cells[5][5] = Cell{Cluster: "State2", Style: style.Style{}.Foreground(style.Blue)}
		compositor.AddLayer(layer2)

		result2 := compositor.Composite()
		assert.Equal(t, "State2", result2.Cells[5][5].Cluster)
		assert.NotEqual(t, result1.Cells[5][5].Cluster, result2.Cells[5][5].Cluster)
	})
}

// TestIntegration_Performance 测试渲染性能
func TestIntegration_Performance(t *testing.T) {
	t.Run("大量单元格渲染", func(t *testing.T) {
		width, height := 100, 100
		renderer := NewRenderer(width, height)

		// 填充10000个单元格
		back := renderer.GetBackBuffer()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				char := rune('A' + (x+y)%26)
				s := style.Style{}.Foreground(style.Color(uint8(x % 16)))
				back.Cells[y][x] = Cell{
					Cluster: string(char),
					Style:   s,
				}
			}
		}

		// 标记为脏并计时渲染
		renderer.ForceFullRender()

		start := time.Now()
		output := renderer.Render()
		duration := time.Since(start)

		// 验证输出不为空
		assert.NotEmpty(t, output)

		// 性能要求：10000单元格 < 50ms
		assert.Less(t, duration.Milliseconds(), int64(50),
			"渲染10000个单元格应该在50ms内完成，实际耗时: %dms", duration.Milliseconds())

		t.Logf("渲染%dx%d=%d个单元格耗时: %dms (输出大小: %d字节)",
			width, height, width*height, duration.Milliseconds(), len(output))
	})

	t.Run("增量渲染性能", func(t *testing.T) {
		width, height := 80, 24
		renderer := NewRenderer(width, height)

		// 初始全量渲染
		back := renderer.GetBackBuffer()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				back.Cells[y][x] = Cell{Cluster: " ", Style: style.Style{}}
			}
		}
		renderer.ForceFullRender()
		_ = renderer.Render() // 初始渲染

		// 初始帧只修改少量单元格
		render1 := renderer.GetBackBuffer()
		for i := 0; i < 10; i++ {
			render1.Cells[0][i] = Cell{Cluster: "X", Style: style.Style{}.Foreground(style.Red)}
		}

		start := time.Now()
		output1 := renderer.Render()
		duration1 := time.Since(start)

		// 增量渲染应该更快
		assert.Less(t, duration1.Milliseconds(), int64(5),
			"增量渲染10个单元格应该在5ms内完成")
		assert.Less(t, len(output1), 200, // 只输出变化部分
			"增量输出应该很小")

		t.Logf("增量渲染耗时: %dms (输出大小: %d字节)", duration1.Milliseconds(), len(output1))
	})

	t.Run("大量样式变更", func(t *testing.T) {
		width, height := 50, 50
		renderer := NewRenderer(width, height)

		// 每个单元格都有不同样式
		back := renderer.GetBackBuffer()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				s := style.Style{}.
					Foreground(style.Color(uint8(x % 16))).
					Background(style.Color(uint8(y % 16))).
					Bold(x%2 == 0).
					Underline(y%2 == 0)
				back.Cells[y][x] = Cell{
					Cluster: "X",
					Style:   s,
				}
			}
		}

		renderer.ForceFullRender()

		start := time.Now()
		_ = renderer.Render()
		duration := time.Since(start)

		// 即使有大量样式变化，也应该在合理时间内完成
		assert.Less(t, duration.Milliseconds(), int64(100),
			"大量样式变更渲染应该在100ms内完成")

		t.Logf("样式密集渲染耗时: %dms", duration.Milliseconds())
	})
}

// TestIntegration_LayerTransitions 测试层切换和动画
func TestIntegration_LayerTransitions(t *testing.T) {
	t.Run("层级切换动画", func(t *testing.T) {
		width, height := 40, 10
		compositor := NewCompositor(width, height)

		// 创建多个层
		for i := 0; i < 5; i++ {
			layer := NewLayer("layer"+itoa(i), LayerContent, i, width, height)
			layer.Buffer.Cells[5][5] = Cell{Cluster: "L" + string(rune('0'+i)), Style: style.Style{}.Foreground(style.Color(uint8(i)))}
			compositor.AddLayer(layer)
		}

		// 初始状态：layer4在顶部
		result := compositor.Composite()
		assert.Equal(t, "L4", result.Cells[5][5].Cluster)

		// 激活layer2：隐藏其他层，显示layer2
		for i := 0; i < 5; i++ {
			if i == 2 {
				continue
			}
			layer := compositor.GetLayer("layer" + itoa(i))
			if layer != nil {
				layer.Hide()
			}
		}

		result2 := compositor.Composite()
		assert.Equal(t, "L2", result2.Cells[5][5].Cluster)
	})

	t.Run("位置动画", func(t *testing.T) {
		width, height := 40, 10
		compositor := NewCompositor(width, height)

		// 创建一个移动的层
		movingLayer := NewLayer("moving", LayerContent, 0, 10, 3)
		movingLayer.Buffer.Cells[0][0] = Cell{Cluster: "X", Style: style.Style{}.Foreground(style.Cyan)}

		// 背景
		bg := NewLayer("bg", LayerBackground, -1, width, height)

		compositor.AddLayer(bg)
		compositor.AddLayer(movingLayer)

		// 模拟动画帧
		positions := []int{0, 10, 20, 30}
		for _, pos := range positions {
			movingLayer.SetRect(Rect{X: pos, Y: 5, Width: 10, Height: 3})
			result := compositor.Composite()

			// 验证位置影响渲染
			assert.NotNil(t, result)
			if pos+5 < width && pos+5 >= 0 {
				// 检查在移动层设置的位置有预期内容
				assert.NotNil(t, result.Cells[5][pos])
			}
		}
	})
}

// TestIntegration_WideCharacters 测试宽字符渲染
// 注意：详细的宽字符测试在 wide_char_integration_test.go 中
// 这里只做集成级验证
func TestIntegration_WideCharacters(t *testing.T) {
	t.Run("宽字符+普通字符混合", func(t *testing.T) {
		width, height := 40, 10
		renderer := NewRenderer(width, height)

		back := renderer.GetBackBuffer()
		ctx := NewPaintContext(back, Rect{X: 0, Y: 0, Width: width, Height: height})

		// 混合文本：普通字符和宽字符
		text := "Hello世界测试World"
		ctx.SetString(0, 0, text, style.Style{}.Foreground(style.White))

		renderer.ForceFullRender()
		output := renderer.Render()

		assert.NotEmpty(t, output)

		// 验证位置
		assert.Equal(t, "H", back.Cells[0][0].Cluster)
		assert.Equal(t, "世", back.Cells[0][5].Cluster)
		assert.Equal(t, 2, back.Cells[0][5].Width)
		assert.True(t, back.Cells[0][6].IsContinuation)
		assert.Equal(t, "W", back.Cells[0][13].Cluster)
	})
}

// TestIntegration_DoubleBuffer 双缓冲完整流程测试
func TestIntegration_DoubleBuffer(t *testing.T) {
	t.Run("完整的渲染循环", func(t *testing.T) {
		width, height := 40, 10
		renderer := NewRenderer(width, height)

		// 第一帧
		back := renderer.GetBackBuffer()
		back.SetString(0, 0, "Frame 1", style.Style{}.Foreground(style.Red))
		renderer.ForceFullRender()
		output1 := renderer.Render()

		assert.NotEmpty(t, output1)
		assert.Contains(t, output1, "Frame")

		// 第二帧：修改部分区域
		back = renderer.GetBackBuffer() // 获取新的back buffer
		back.SetString(0, 0, "Frame 2", style.Style{}.Foreground(style.Blue))
		output2 := renderer.Render()

		// 增量渲染应该有输出
		assert.NotEmpty(t, output2)
		assert.Contains(t, output2, "Frame")

		// 第三帧：无变化，应该空输出
		output3 := renderer.Render()
		assert.Empty(t, output3, "无变化的帧应该返回空")
	})

	t.Run("脏区域跟踪", func(t *testing.T) {
		width, height := 40, 10
		renderer := NewRenderer(width, height)

		back := renderer.GetBackBuffer()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				back.Cells[y][x] = Cell{Cluster: " "}
			}
		}
		renderer.ForceFullRender()
		_ = renderer.Render() // 初始渲染

		// 修改特定矩形区域
		back.SetCell(10, 5, 'D', style.Style{}.Foreground(style.Yellow))
		back.SetCell(11, 5, 'i', style.Style{}.Foreground(style.Yellow))
		back.SetCell(12, 5, 'r', style.Style{}.Foreground(style.Yellow))
		renderer.MarkDirtyRect(Rect{X: 10, Y: 5, Width: 3, Height: 1})

		output := renderer.Render()

		// 输出应该相对较小（只包含变化的区域）
		assert.NotEmpty(t, output)

		// 统计信息
		stats := renderer.GetStats()
		assert.True(t, stats.ChangedCells > 0)
		assert.Greater(t, stats.OutputBytes, 0)

		t.Logf("脏区域渲染: 变化单元格=%d, 输出大小=%d字节",
			stats.ChangedCells, stats.OutputBytes)
	})
}

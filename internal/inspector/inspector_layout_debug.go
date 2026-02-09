package inspector

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"os"
)

// DumpInspectorLayout 将 Inspector 自身的 VNode 树导出到文件
func (si *StandaloneInspector) DumpInspectorLayout(filePath string) error {
	si.mu.RLock()
	defer si.mu.RUnlock()

	// 1. 构建 Inspector 的 UI 树
	inspectorRoot := si.buildOverlayContent()
	if inspectorRoot == nil {
		return fmt.Errorf("inspector root is nil")
	}

	// 2. 使用真实的 Engine 执行测量
	// 这会触发布局分配并将结果存储在节点的 RenderInfo 中
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  si.overlayWidth,
		MinHeight: 0,
		MaxHeight: si.overlayHeight,
	}

	// 执行测量 (这会递归测量所有子节点)
	// 我们需要将 VNode 转换为 compute.VNode
	engine.Layout(inspectorRoot, constraints)

	// 3. 使用 Analyzer 捕获并格式化
	analyzer := NewLayoutAnalyzer()
	snapshot := analyzer.Capture(inspectorRoot, 0)
	treeStr := analyzer.FormatTree(snapshot)

	return os.WriteFile(filePath, []byte(treeStr), 0644)
}

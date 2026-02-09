package layer

import (
	"fmt"
	"os"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFragmentType 检查 Fragment 的实际类型
func TestFragmentType(t *testing.T) {
	appContent := rtui.NewElement("app")
	inspectorOverlay := rtui.NewElement("inspector")
	inspectorOverlay.SetLayer(rtui.LayerInspector)

	root := rtui.Fragment(appContent, inspectorOverlay)

	fmt.Fprintf(os.Stderr, "root type: %T\n", root)

	if fragment, ok := root.(*rtui.FragmentVNode); ok {
		fmt.Fprintf(os.Stderr, "✅ root is FragmentVNode\n")
		fmt.Fprintf(os.Stderr, "fragment.Children() count: %d\n", len(fragment.Children()))
	} else {
		fmt.Fprintf(os.Stderr, "❌ root is NOT FragmentVNode!\n")
	}

	// 检查 NewFragment 返回的类型
	newFragment := rtui.NewFragment()
	fmt.Fprintf(os.Stderr, "NewFragment() type: %T\n", newFragment)
}

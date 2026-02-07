package layout

import (
	"testing"

	"github.com/wwsheng009/mint/components/basic"
	"github.com/wwsheng009/mint/runtime/style"
	ui "github.com/wwsheng009/mint/runtime/ui"
)

func TestWrap_BasicWrapping(t *testing.T) {
	btn1 := basic.NewText("Button1")
	btn2 := basic.NewText("Button2")
	btn3 := basic.NewText("Button3")

	// Screen width only 20, should wrap to multiple rows
	wrapped := NewWrapBuilder(btn1, btn2, btn3).
		Gap(1).
		ScreenWidth(20).
		Build()

	// Verify result is not nil
	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	// The result should be a VStack containing multiple HStacks
	// We can't easily test the structure without exposing internal details,
	// but we can verify it builds without panic
	t.Logf("Wrap built successfully: %T", wrapped)
}

func TestWrap_NoWrappingNeeded(t *testing.T) {
	items := []ui.VNode{
		basic.NewText("A"),
		basic.NewText("B"),
		basic.NewText("C"),
	}

	// Screen width 100, should not need to wrap
	wrapped := NewWrapBuilder(items...).
		Gap(1).
		ScreenWidth(100).
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	t.Logf("Wrap built successfully: %T", wrapped)
}

func TestWrap_EmptyChildren(t *testing.T) {
	wrapped := NewWrapBuilder().
		Gap(1).
		ScreenWidth(80).
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	t.Logf("Wrap with empty children built successfully: %T", wrapped)
}

func TestWrap_AlignCenter(t *testing.T) {
	items := createTestButtons(5)

	wrapped := NewWrapBuilder(items...).
		Gap(1).
		Align(ui.AlignCenter).
		ScreenWidth(40).
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	t.Logf("Wrap with AlignCenter built successfully: %T", wrapped)
}

func TestWrap_RowGap(t *testing.T) {
	items := createTestButtons(8)

	wrapped := NewWrapBuilder(items...).
		Gap(1).
		RowGap(2).
		ScreenWidth(40).
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	t.Logf("Wrap with RowGap built successfully: %T", wrapped)
}

func TestWrap_SingleItem(t *testing.T) {
	item := basic.NewText("Single")

	wrapped := NewWrapBuilder(item).
		Gap(1).
		ScreenWidth(80).
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	t.Logf("Wrap with single item built successfully: %T", wrapped)
}

func TestWrap_StyleAndKey(t *testing.T) {
	items := createTestButtons(3)

	s := style.Style{}.Bold(true)

	wrapped := NewWrapBuilder(items...).
		Gap(1).
		ScreenWidth(80).
		Style(s).
		Key("test-wrap").
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	resultStyle := wrapped.Style()
	if !resultStyle.IsBold() {
		t.Errorf("Expected Bold to be true")
	}

	if wrapped.Key() != "test-wrap" {
		t.Errorf("Expected key to be 'test-wrap', got %s", wrapped.Key())
	}

	t.Logf("Wrap with style and key built successfully")
}

func TestWrap_FillWidth(t *testing.T) {
	items := createTestButtons(5)

	wrapped := NewWrapBuilder(items...).
		Gap(1).
		ScreenWidth(40).
		FillWidth().
		Build()

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}

	// Verify stretchCross is set (via LayoutNode interface)
	if layoutNode, ok := wrapped.(interface{ StretchCross() bool }); ok {
		if !layoutNode.StretchCross() {
			t.Errorf("Expected StretchCross to be true")
		}
	} else {
		t.Logf("Warning: Cannot check StretchCross, wrapped is not layout.LayoutNode")
	}

	t.Logf("Wrap with FillWidth built successfully")
}

// BenchmarkWrap_Building benchmarks the Wrap building process
func BenchmarkWrap_Building(b *testing.B) {
	items := createTestButtons(20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapped := NewWrapBuilder(items...).
			Gap(1).
			ScreenWidth(80).
			Build()
		_ = wrapped
	}
}

// createTestButtons creates test button VNodes for testing
func createTestButtons(count int) []ui.VNode {
	items := make([]ui.VNode, count)
	for i := 0; i < count; i++ {
		items[i] = basic.NewText("Btn")
	}
	return items
}

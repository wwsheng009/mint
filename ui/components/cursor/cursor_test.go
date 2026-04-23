package cursor

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestVNodeBuilderDefaults(t *testing.T) {
	vnode := NewBuilder().BuildTyped()

	if vnode.Tag() != "cursor" {
		t.Fatalf("Tag = %q, want %q", vnode.Tag(), "cursor")
	}
	if vnode.Config().BlinkInterval != NormalBlinkInterval {
		t.Fatalf("BlinkInterval = %v, want %v", vnode.Config().BlinkInterval, NormalBlinkInterval)
	}
}

func TestInstancePaintBlock(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"config": Config{
			Shape:         ShapeBlock,
			Blink:         false,
			BlinkInterval: NormalBlinkInterval,
		},
	})

	cmds := inst.Paint(4, 2)
	if len(cmds) != 1 {
		t.Fatalf("Paint returned %d commands, want 1", len(cmds))
	}
	if cmds[0].X != 4 || cmds[0].Y != 2 {
		t.Fatalf("Cursor position = (%d,%d), want (4,2)", cmds[0].X, cmds[0].Y)
	}
	if !cmds[0].Style.IsReverse() {
		t.Fatal("Block cursor should use reverse style")
	}
}

func TestInstanceTickBlink(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"config": Config{
			Blink:         true,
			BlinkInterval: 5 * time.Millisecond,
		},
	})

	if !inst.WantsTick() {
		t.Fatal("Blinking cursor should want ticks")
	}

	time.Sleep(6 * time.Millisecond)
	if !inst.Tick(time.Now()) {
		t.Fatal("Tick should toggle blink state")
	}

	cmds := inst.Paint(0, 0)
	if len(cmds) != 0 {
		t.Fatalf("Hidden blink phase should paint 0 commands, got %d", len(cmds))
	}
}

func TestBuilderSpeedPresets(t *testing.T) {
	vnode := NewBuilder().FastBlink().BuildTyped()
	if vnode.Config().BlinkInterval != FastBlinkInterval {
		t.Fatalf("FastBlink interval = %v, want %v", vnode.Config().BlinkInterval, FastBlinkInterval)
	}

	vnode = NewBuilder().SlowBlink().BuildTyped()
	if vnode.Config().BlinkInterval != SlowBlinkInterval {
		t.Fatalf("SlowBlink interval = %v, want %v", vnode.Config().BlinkInterval, SlowBlinkInterval)
	}
}

func TestInstanceMeasure(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width != 1 || size.Height != 1 {
		t.Fatalf("Measure = %+v, want {Width:1 Height:1}", size)
	}
}

func TestNormalizeConfig_ZeroUsesDefaults(t *testing.T) {
	cfg := NormalizeConfig(Config{})
	if !cfg.Blink {
		t.Fatal("zero config should default Blink to true")
	}
	if cfg.BlinkInterval != NormalBlinkInterval {
		t.Fatalf("BlinkInterval = %v, want %v", cfg.BlinkInterval, NormalBlinkInterval)
	}
}

func TestNormalizeConfig_PartialKeepsBlinkDefault(t *testing.T) {
	cfg := NormalizeConfig(Config{
		Shape: ShapeBar,
	})
	if !cfg.Blink {
		t.Fatal("partial config should keep default Blink=true")
	}
}

func TestNormalizeConfig_ExplicitBlinkFalse(t *testing.T) {
	cfg := NormalizeConfig(Config{
		Blink:         false,
		BlinkInterval: NormalBlinkInterval,
	})
	if cfg.Blink {
		t.Fatal("Blink should remain false when explicitly configured")
	}
}

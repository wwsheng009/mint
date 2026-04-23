package paint

import (
	"sync"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/style"
)

func TestAsyncRenderer_SubmitAndFlushNow(t *testing.T) {
	var mu sync.Mutex
	var outputs []string

	ar := NewAsyncRenderer(10, 2, AsyncRendererOptions{
		FrameInterval: time.Second,
		Output: func(s string) {
			mu.Lock()
			outputs = append(outputs, s)
			mu.Unlock()
		},
	})
	ar.Start()
	defer ar.Stop()

	frame := NewBuffer(10, 2)
	frame.SetString(0, 0, "HELLO", style.NewStyle())
	ar.SubmitFrame(frame, nil, true)
	ar.FlushNow()

	mu.Lock()
	defer mu.Unlock()
	if len(outputs) == 0 {
		t.Fatal("expected async renderer to produce output")
	}
}

func TestAsyncRenderer_PartialUpdateWithHint(t *testing.T) {
	var mu sync.Mutex
	var outputs []string

	ar := NewAsyncRenderer(8, 1, AsyncRendererOptions{
		FrameInterval: time.Second,
		Output: func(s string) {
			mu.Lock()
			outputs = append(outputs, s)
			mu.Unlock()
		},
	})
	ar.Start()
	defer ar.Stop()

	st := style.NewStyle()
	frame1 := NewBuffer(8, 1)
	frame1.SetStringAligned(0, 0, "ABCDEFGH", st, 8)
	ar.SubmitFrame(frame1, nil, true)
	ar.FlushNow()

	frame2 := NewBuffer(8, 1)
	frame2.SetStringAligned(0, 0, "ABCDEFGH", st, 8)
	// No content change, only hint region should force output.
	ar.SubmitFrame(frame2, []Rect{{X: 2, Y: 0, Width: 2, Height: 1}}, false)
	ar.FlushNow()

	mu.Lock()
	defer mu.Unlock()
	if len(outputs) < 2 {
		t.Fatalf("expected 2 outputs, got %d", len(outputs))
	}
}

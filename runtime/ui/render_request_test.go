package ui

import "testing"

func TestRequestGlobalRenderInvokesRegisteredCallback(t *testing.T) {
	SetGlobalRenderScheduler(nil)
	t.Cleanup(func() {
		SetGlobalRenderScheduler(nil)
	})

	called := 0
	SetGlobalRenderScheduler(func() {
		called++
	})

	RequestGlobalRender()

	if called != 1 {
		t.Fatalf("RequestGlobalRender() calls = %d, want 1", called)
	}
}

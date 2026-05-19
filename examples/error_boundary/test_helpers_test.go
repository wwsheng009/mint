package main

import "testing"

func resetErrorBoundaryTestState(t *testing.T) {
	t.Helper()
	resetErrorBoundaryGlobals()
	t.Cleanup(resetErrorBoundaryGlobals)
}

func resetErrorBoundaryGlobals() {
	shouldPanic = false
	panicThreshold = 5
	panicInEffect = false
	shouldPanicOnRender = false
}

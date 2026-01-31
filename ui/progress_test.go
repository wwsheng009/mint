package ui

import "testing"

// TestNewProgress tests progress bar creation
func TestNewProgress(t *testing.T) {
	progress := NewProgress()

	if progress == nil {
		t.Fatal("NewProgress() returned nil")
	}

	if progress.Value() != 0 {
		t.Errorf("Initial value = %v, want 0", progress.Value())
	}

	if progress.Max() != 100 {
		t.Errorf("Initial max = %v, want 100", progress.Max())
	}

	if progress.Width() != 30 {
		t.Errorf("Initial width = %v, want 30", progress.Width())
	}
}

// TestProgressBuilder tests progress bar builder
func TestProgressBuilder(t *testing.T) {
	progress := ProgressBuilder().
		Value(50).
		Max(200).
		Width(40).
		ShowPercent(true).
		Label("Loading").
		Build()

	progressVNode, ok := progress.(*ProgressVNode)
	if !ok {
		t.Fatal("ProgressBuilder() did not return *ProgressVNode")
	}

	if progressVNode.Value() != 50 {
		t.Errorf("Value = %v, want 50", progressVNode.Value())
	}

	if progressVNode.Max() != 200 {
		t.Errorf("Max = %v, want 200", progressVNode.Max())
	}

	if progressVNode.Width() != 40 {
		t.Errorf("Width = %v, want 40", progressVNode.Width())
	}

	if progressVNode.Label() != "Loading" {
		t.Errorf("Label = %v, want 'Loading'", progressVNode.Label())
	}
}

// TestProgressSetValue tests setting progress value
func TestProgressSetValue(t *testing.T) {
	progress := NewProgress()

	progress.SetValue(75)
	if progress.Value() != 75 {
		t.Errorf("Value = %v, want 75", progress.Value())
	}

	// Test clamping at max
	progress.SetValue(150)
	if progress.Value() != 100 {
		t.Errorf("Value should be clamped to max (100), got %v", progress.Value())
	}

	// Test clamping at 0
	progress.SetValue(-10)
	if progress.Value() != 0 {
		t.Errorf("Value should be clamped to 0, got %v", progress.Value())
	}
}

// TestProgressPercent tests percentage calculation
func TestProgressPercent(t *testing.T) {
	progress := NewProgress()

	progress.SetValue(0)
	if progress.Percent() != 0 {
		t.Errorf("Percent = %v, want 0", progress.Percent())
	}

	progress.SetValue(50)
	if progress.Percent() != 50 {
		t.Errorf("Percent = %v, want 50", progress.Percent())
	}

	progress.SetValue(100)
	if progress.Percent() != 100 {
		t.Errorf("Percent = %v, want 100", progress.Percent())
	}

	// Test with different max
	progress.SetMax(200)
	progress.SetValue(100)
	if progress.Percent() != 50 {
		t.Errorf("Percent = %v, want 50 (100/200)", progress.Percent())
	}
}

// TestProgressSetMax tests setting max value
func TestProgressSetMax(t *testing.T) {
	progress := NewProgress()

	progress.SetMax(500)
	if progress.Max() != 500 {
		t.Errorf("Max = %v, want 500", progress.Max())
	}

	progress.SetMax(0)
	progress.SetValue(100)
	// Should handle zero max gracefully
	if progress.Percent() != 0 {
		t.Errorf("Percent with zero max should be 0, got %v", progress.Percent())
	}
}

// TestProgressWidth tests width setting
func TestProgressWidth(t *testing.T) {
	progress := NewProgress()

	progress.SetWidth(50)
	if progress.Width() != 50 {
		t.Errorf("Width = %v, want 50", progress.Width())
	}

	// Test minimum width
	progress.SetWidth(5)
	if progress.Width() != 10 {
		t.Errorf("Width should be clamped to minimum 10, got %v", progress.Width())
	}
}

// TestProgressShowPercent tests show percent flag
func TestProgressShowPercent(t *testing.T) {
	progress := NewProgress()

	if !progress.ShowPercent() {
		t.Error("ShowPercent should be true by default")
	}

	progress.SetShowPercent(false)
	if progress.ShowPercent() {
		t.Error("ShowPercent should be false after SetShowPercent(false)")
	}
}

// TestProgressLabel tests label setting
func TestProgressLabel(t *testing.T) {
	progress := NewProgress()

	progress.SetLabel("Downloading...")
	if progress.Label() != "Downloading..." {
		t.Errorf("Label = %v, want 'Downloading...'", progress.Label())
	}
}

// TestNewSpinner tests spinner creation
func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner()

	if spinner == nil {
		t.Fatal("NewSpinner() returned nil")
	}

	if spinner.Message() != "Loading..." {
		t.Errorf("Initial message = %v, want 'Loading...'", spinner.Message())
	}

	if len(spinner.Frames()) == 0 {
		t.Error("Frames should not be empty")
	}
}

// TestSpinnerBuilder tests spinner builder
func TestSpinnerBuilder(t *testing.T) {
	spinner := SpinnerBuilder().
		Message("Please wait...").
		Build()

	spinnerVNode, ok := spinner.(*SpinnerVNode)
	if !ok {
		t.Fatal("SpinnerBuilder() did not return *SpinnerVNode")
	}

	if spinnerVNode.Message() != "Please wait..." {
		t.Errorf("Message = %v, want 'Please wait...'", spinnerVNode.Message())
	}
}

// TestSpinnerSetMessage tests setting spinner message
func TestSpinnerSetMessage(t *testing.T) {
	spinner := NewSpinner()

	spinner.SetMessage("Processing...")
	if spinner.Message() != "Processing..." {
		t.Errorf("Message = %v, want 'Processing...'", spinner.Message())
	}
}

// TestSpinnerFrames tests custom frames
func TestSpinnerFrames(t *testing.T) {
	spinner := NewSpinner()
	customFrames := []string{"|", "/", "-", "\\"}

	spinner.SetFrames(customFrames)
	frames := spinner.Frames()
	if len(frames) != 4 {
		t.Errorf("Frames length = %v, want 4", len(frames))
	}
}

// TestSpinnerCurrentFrame tests current frame
func TestSpinnerCurrentFrame(t *testing.T) {
	spinner := NewSpinner()

	frame := spinner.CurrentFrame()
	if frame == "" {
		t.Error("CurrentFrame should not be empty")
	}
}

// BenchmarkProgressBuilder benchmarks progress bar builder
func BenchmarkProgressBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProgressBuilder().
			Value(50).
			Width(30).
			Build()
	}
}

// BenchmarkSpinnerBuilder benchmarks spinner builder
func BenchmarkSpinnerBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SpinnerBuilder().
			Message("Loading...").
			Build()
	}
}

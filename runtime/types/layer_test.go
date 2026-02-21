package types

import (
	"testing"
)

func TestLayer_String(t *testing.T) {
	tests := []struct {
		layer    Layer
		expected string
	}{
		{LayerBase, "base"},
		{LayerOverlay, "overlay"},
		{LayerModal, "modal"},
		{LayerTooltip, "tooltip"},
		{LayerInspector, "inspector"},
		{Layer(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.layer.String(); got != tt.expected {
				t.Errorf("Layer(%d).String() = %q, want %q", tt.layer, got, tt.expected)
			}
		})
	}
}

func TestLayer_ZIndex(t *testing.T) {
	if LayerBase.ZIndex() != 0 {
		t.Errorf("LayerBase.ZIndex() = %d, want 0", LayerBase.ZIndex())
	}
	if LayerOverlay.ZIndex() != 1 {
		t.Errorf("LayerOverlay.ZIndex() = %d, want 1", LayerOverlay.ZIndex())
	}
	if LayerModal.ZIndex() != 2 {
		t.Errorf("LayerModal.ZIndex() = %d, want 2", LayerModal.ZIndex())
	}
	if LayerTooltip.ZIndex() != 3 {
		t.Errorf("LayerTooltip.ZIndex() = %d, want 3", LayerTooltip.ZIndex())
	}
	if LayerInspector.ZIndex() != 4 {
		t.Errorf("LayerInspector.ZIndex() = %d, want 4", LayerInspector.ZIndex())
	}
}

func TestLayer_IsValid(t *testing.T) {
	if !LayerBase.IsValid() {
		t.Error("LayerBase.IsValid() should be true")
	}
	if !LayerModal.IsValid() {
		t.Error("LayerModal.IsValid() should be true")
	}
	if !LayerInspector.IsValid() {
		t.Error("LayerInspector.IsValid() should be true")
	}
	if Layer(99).IsValid() {
		t.Error("Layer(99).IsValid() should be false")
	}
	if Layer(-1).IsValid() {
		t.Error("Layer(-1).IsValid() should be false")
	}
}

func TestLayer_IsModal(t *testing.T) {
	if LayerBase.IsModal() {
		t.Error("LayerBase.IsModal() should be false")
	}
	if !LayerModal.IsModal() {
		t.Error("LayerModal.IsModal() should be true")
	}
	if LayerInspector.IsModal() {
		t.Error("LayerInspector.IsModal() should be false")
	}
}

func TestLayer_IsOverlay(t *testing.T) {
	if LayerBase.IsOverlay() {
		t.Error("LayerBase.IsOverlay() should be false")
	}
	if !LayerOverlay.IsOverlay() {
		t.Error("LayerOverlay.IsOverlay() should be true")
	}
	if !LayerModal.IsOverlay() {
		t.Error("LayerModal.IsOverlay() should be true")
	}
	if !LayerTooltip.IsOverlay() {
		t.Error("LayerTooltip.IsOverlay() should be true")
	}
	if !LayerInspector.IsOverlay() {
		t.Error("LayerInspector.IsOverlay() should be true")
	}
}

func TestLayer_Ordering(t *testing.T) {
	// 验证层级顺序: Base < Overlay < Modal < Tooltip < Inspector
	if LayerBase >= LayerOverlay {
		t.Error("LayerBase should be less than LayerOverlay")
	}
	if LayerOverlay >= LayerModal {
		t.Error("LayerOverlay should be less than LayerModal")
	}
	if LayerModal >= LayerTooltip {
		t.Error("LayerModal should be less than LayerTooltip")
	}
	if LayerTooltip >= LayerInspector {
		t.Error("LayerTooltip should be less than LayerInspector")
	}
}

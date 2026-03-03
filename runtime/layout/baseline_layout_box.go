package layout

// GetEffectiveBaseline returns the baseline, or estimated if not available
func (b *BaselineLayoutBox) GetEffectiveBaseline() int {
	if b.HasBaselineInfo {
		return b.Baseline
	}
	// Default baseline is 2/3 of height (typical for text)
	if b.Height > 0 {
		return b.Height * 2 / 3
	}
	return 0
}
package selection

import "testing"

func TestSelectionModesRemainStable(t *testing.T) {
	tests := []struct {
		name string
		got  SelectionMode
		want SelectionMode
	}{
		{name: "none is zero value", got: SelectionNone, want: 0},
		{name: "single follows none", got: SelectionSingle, want: 1},
		{name: "multiple follows single", got: SelectionMultiple, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSelectionModesAreDistinct(t *testing.T) {
	seen := map[SelectionMode]bool{}
	for _, mode := range []SelectionMode{SelectionNone, SelectionSingle, SelectionMultiple} {
		if seen[mode] {
			t.Fatalf("duplicate selection mode value detected: %d", mode)
		}
		seen[mode] = true
	}
}

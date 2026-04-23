package downsample

import "testing"

func TestResampleNearestFloat64(t *testing.T) {
	got := ResampleNearestFloat64([]float64{1, 2, 3, 4, 5}, 3)
	want := []float64{1, 3, 5}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestResampleNearestFloat64SingleWidth(t *testing.T) {
	got := ResampleNearestFloat64([]float64{3, 6, 9}, 1)
	if len(got) != 1 || got[0] != 9 {
		t.Fatalf("ResampleNearestFloat64(..., 1) = %#v, want []float64{9}", got)
	}
}

func TestMinMaxFloat64(t *testing.T) {
	minVal, maxVal := MinMaxFloat64([]float64{5, 1, 9, 2})
	if minVal != 1 || maxVal != 9 {
		t.Fatalf("MinMaxFloat64() = (%v, %v), want (1, 9)", minVal, maxVal)
	}
}

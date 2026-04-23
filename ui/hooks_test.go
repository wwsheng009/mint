package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// deepEqual Tests
// =============================================================================

func TestDeepEqual_BasicTypes(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"int: equal", 42, 42, true},
		{"int: not equal", 42, 43, false},
		{"int64: equal", int64(42), int64(42), true},
		{"string: equal", "hello", "hello", true},
		{"string: not equal", "hello", "world", false},
		{"bool: equal true", true, true, true},
		{"bool: equal false", false, false, true},
		{"bool: not equal", true, false, false},
		{"float64: equal", 3.14, 3.14, true},
		{"float64: not equal", 3.14, 2.71, false},
		{"nil: both nil", nil, nil, true},
		{"nil: one nil", nil, 42, false},
		{"nil: other nil", 42, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeepEqual_Slices(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"slice: equal", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"slice: different length", []int{1, 2}, []int{1, 2, 3}, false},
		{"slice: different values", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"slice: empty", []int{}, []int{}, true},
		{"slice: nil vs empty", ([]int)(nil), []int{}, false},
		{"slice: string", []string{"a", "b"}, []string{"a", "b"}, true},
		{"slice: struct", []struct{ X int }{{1}, {2}}, []struct{ X int }{{1}, {2}}, true},
		{"slice: nested", [][]int{{1, 2}, {3}}, [][]int{{1, 2}, {3}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeepEqual_Maps(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"map: equal", map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 2}, true},
		{"map: different size", map[string]int{"a": 1}, map[string]int{"a": 1, "b": 2}, false},
		{"map: different values", map[string]int{"a": 1}, map[string]int{"a": 2}, false},
		{"map: empty", map[string]int{}, map[string]int{}, true},
		{"map: nil vs empty", (map[string]int)(nil), map[string]int{}, false},
		{"map: int keys", map[int]string{1: "a", 2: "b"}, map[int]string{1: "a", 2: "b"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeepEqual_Structs(t *testing.T) {
	type SimpleStruct struct {
		Name string
		Age  int
	}

	type NestedStruct struct {
		Inner SimpleStruct
	}

	type UnexportedStruct struct {
		name string // unexported field should be skipped
		Age  int
	}

	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"struct: equal", SimpleStruct{Name: "Alice", Age: 25}, SimpleStruct{Name: "Alice", Age: 25}, true},
		{"struct: different field", SimpleStruct{Name: "Alice", Age: 25}, SimpleStruct{Name: "Bob", Age: 25}, false},
		{"struct: nested", NestedStruct{Inner: SimpleStruct{Name: "Alice", Age: 25}}, NestedStruct{Inner: SimpleStruct{Name: "Alice", Age: 25}}, true},
		{"struct: pointer values", &SimpleStruct{Name: "Alice"}, &SimpleStruct{Name: "Alice"}, true}, // deep comparison - same value
		{"struct: unexported", UnexportedStruct{name: "Alice", Age: 25}, UnexportedStruct{name: "Bob", Age: 25}, true}, // unexported ignored
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeepEqual_Pointers(t *testing.T) {
	x := 42
	y := 42
	z := 43

	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"pointer: same", &x, &x, true},
		{"pointer: different values", &x, &z, false},
		{"pointer: same value different address", &x, &y, true}, // deep comparison - same value
		{"pointer: nil", (*int)(nil), (*int)(nil), true},
		{"pointer: nil vs non-nil", (*int)(nil), &x, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeepEqual_EdgeCases(t *testing.T) {
	arr := [3]int{1, 2, 3}
	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}

	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"array vs slice", arr, slice1, false}, // different types
		{"same slice content different addresses", slice1, slice2, true}, // deep comparison
		{"func: same", func() {}, func() {}, false}, // functions not equal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

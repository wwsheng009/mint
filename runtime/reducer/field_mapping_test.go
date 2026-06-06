package reducer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
)

type TestState struct {
	Username string
	Email    string
	Age      int
	Price    float64
	Agreed   bool
}

func TestBindStringField(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	var lastState TestState

	fb.BindStringField("username", func(s TestState, val string) TestState {
		lastState = s
		s.Username = val
		return s
	})

	reducer := fb.Build()

	// Test string field update
	newState := reducer(lastState, intent.FieldChangeIntent{Field: "username", Value: "john"})
	if newState.Username != "john" {
		t.Errorf("expected Username=john, got %s", newState.Username)
	}

	// Test with different field (should not update)
	newState2 := reducer(newState, intent.FieldChangeIntent{Field: "email", Value: "test@example.com"})
	if newState2.Username != "john" {
		t.Errorf("expected Username to remain john, got %s", newState2.Username)
	}
}

func TestBindIntField(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	var lastState TestState

	fb.BindIntField("age", func(s TestState, val int) TestState {
		lastState = s
		s.Age = val
		return s
	})

	reducer := fb.Build()

	// Test int field update
	newState := reducer(lastState, intent.FieldChangeIntent{Field: "age", Value: "25"})
	if newState.Age != 25 {
		t.Errorf("expected Age=25, got %d", newState.Age)
	}

	// Test invalid int (should keep original)
	newState2 := reducer(newState, intent.FieldChangeIntent{Field: "age", Value: "invalid"})
	if newState2.Age != 25 {
		t.Errorf("expected Age to remain 25, got %d", newState2.Age)
	}

	// Test another valid int
	newState3 := reducer(newState2, intent.FieldChangeIntent{Field: "age", Value: "30"})
	if newState3.Age != 30 {
		t.Errorf("expected Age=30, got %d", newState3.Age)
	}
}

func TestBindBoolField(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	var lastState TestState

	fb.BindBoolField("agreed", func(s TestState, val bool) TestState {
		lastState = s
		s.Agreed = val
		return s
	})

	reducer := fb.Build()

	// Test bool field update (true)
	newState := reducer(lastState, intent.FieldChangeIntent{Field: "agreed", Value: "true"})
	if newState.Agreed != true {
		t.Errorf("expected Agreed=true, got %v", newState.Agreed)
	}

	// Test bool field update (false)
	newState2 := reducer(newState, intent.FieldChangeIntent{Field: "agreed", Value: "false"})
	if newState2.Agreed != false {
		t.Errorf("expected Agreed=false, got %v", newState2.Agreed)
	}

	// Test case insensitive
	newState3 := reducer(newState2, intent.FieldChangeIntent{Field: "agreed", Value: "TRUE"})
	if newState3.Agreed != true {
		t.Errorf("expected Agreed=true (case insensitive), got %v", newState3.Agreed)
	}
}

func TestBindFloatField(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	var lastState TestState

	fb.BindFloatField("price", func(s TestState, val float64) TestState {
		lastState = s
		s.Price = val
		return s
	})

	reducer := fb.Build()

	// Test float field update
	newState := reducer(lastState, intent.FieldChangeIntent{Field: "price", Value: "19.99"})
	if newState.Price != 19.99 {
		t.Errorf("expected Price=19.99, got %.2f", newState.Price)
	}

	// Test invalid float (should keep original)
	newState2 := reducer(newState, intent.FieldChangeIntent{Field: "price", Value: "invalid"})
	if newState2.Price != 19.99 {
		t.Errorf("expected Price to remain 19.99, got %.2f", newState2.Price)
	}

	// Test another valid float
	newState3 := reducer(newState2, intent.FieldChangeIntent{Field: "price", Value: "29.99"})
	if newState3.Price != 29.99 {
		t.Errorf("expected Price=29.99, got %.2f", newState3.Price)
	}
}

func TestBindFieldMap(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	initialState := TestState{}

	fieldMap := FieldMap[TestState]{
		"username": func(s TestState, val string) TestState {
			s.Username = val
			return s
		},
		"email": func(s TestState, val string) TestState {
			s.Email = val
			return s
		},
		"age": func(s TestState, val string) TestState {
			var age int
			if val != "" {
				n, err := strconv.Atoi(val)
				if err == nil {
					age = n
				}
			}
			s.Age = age
			return s
		},
	}

	reducer := fb.BindFieldMap(fieldMap).Build()

	// Test multiple field updates
	state1 := reducer(initialState, intent.FieldChangeIntent{Field: "username", Value: "john"})
	if state1.Username != "john" {
		t.Errorf("expected Username=john, got %s", state1.Username)
	}

	state2 := reducer(state1, intent.FieldChangeIntent{Field: "email", Value: "john@example.com"})
	if state2.Email != "john@example.com" {
		t.Errorf("expected Email=john@example.com, got %s", state2.Email)
	}
	if state2.Username != "john" {
		t.Errorf("expected Username to remain john, got %s", state2.Username)
	}

	state3 := reducer(state2, intent.FieldChangeIntent{Field: "age", Value: "25"})
	if state3.Age != 25 {
		t.Errorf("expected Age=25, got %d", state3.Age)
	}
}

func TestUpdateStringFieldIfChangedSkipsSameValue(t *testing.T) {
	state := TestState{Username: "john", Email: "old@example.com"}
	calls := 0

	same := UpdateStringFieldIfChanged(state, state.Username, "john", func(s TestState, val string) TestState {
		calls++
		s.Username = val
		s.Email = "changed@example.com"
		return s
	})
	if calls != 0 || same.Username != "john" || same.Email != "old@example.com" {
		t.Fatalf("same value update = %+v calls=%d, want unchanged no-op", same, calls)
	}

	changed := UpdateStringFieldIfChanged(state, state.Username, "jane", func(s TestState, val string) TestState {
		calls++
		s.Username = val
		s.Email = "changed@example.com"
		return s
	})
	if calls != 1 || changed.Username != "jane" || changed.Email != "changed@example.com" {
		t.Fatalf("changed value update = %+v calls=%d, want updater applied once", changed, calls)
	}

	nilUpdate := UpdateStringFieldIfChanged(state, state.Username, "jane", nil)
	if nilUpdate != state {
		t.Fatalf("nil updater = %+v, want unchanged state", nilUpdate)
	}
}

func TestBindFieldMapWithEntries(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	initialState := TestState{}

	entries := map[string]*FieldEntry[TestState]{
		"username": {
			Updater: func(s TestState, val string) TestState {
				s.Username = val
				return s
			},
			Validator: func(field, val string) bool {
				return len(val) >= 3
			},
			Required: true,
		},
		"email": {
			Updater: func(s TestState, val string) TestState {
				s.Email = val
				return s
			},
			Validator: func(field, val string) bool {
				return val != "" && val != "invalid"
			},
			Transform: func(val string) string {
				return val // Could be more complex
			},
		},
	}

	reducer := fb.BindFieldMapWithEntries(entries).Build()

	// Test required field with valid value
	state1 := reducer(initialState, intent.FieldChangeIntent{Field: "username", Value: "john"})
	if state1.Username != "john" {
		t.Errorf("expected Username=john, got %s", state1.Username)
	}

	// Test required field with invalid value (too short)
	state2 := reducer(state1, intent.FieldChangeIntent{Field: "username", Value: "ab"})
	if state2.Username != "john" {
		t.Errorf("expected Username to remain john (invalid value), got %s", state2.Username)
	}

	// Test required field with empty value
	state3 := reducer(state2, intent.FieldChangeIntent{Field: "username", Value: ""})
	if state3.Username != "john" {
		t.Errorf("expected Username to remain john (empty value), got %s", state3.Username)
	}

	// Test validation
	state4 := reducer(state3, intent.FieldChangeIntent{Field: "email", Value: "invalid"})
	if state4.Email != "" {
		t.Errorf("expected Email to be empty (invalid value), got %s", state4.Email)
	}

	state5 := reducer(state4, intent.FieldChangeIntent{Field: "email", Value: "test@example.com"})
	if state5.Email != "test@example.com" {
		t.Errorf("expected Email=test@example.com, got %s", state5.Email)
	}
}

func TestBindFieldGeneric(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	var lastState TestState

	fb.BindFieldGeneric("username", func(s TestState, val string) TestState {
		lastState = s
		s.Username = val
		return s
	})

	reducer := fb.Build()

	// Test generic field binding
	newState := reducer(lastState, intent.FieldChangeIntent{Field: "username", Value: "jane"})
	if newState.Username != "jane" {
		t.Errorf("expected Username=jane, got %s", newState.Username)
	}

	// Test with wrong field
	newState2 := reducer(newState, intent.FieldChangeIntent{Field: "password", Value: "secret"})
	if newState2.Username != "jane" {
		t.Errorf("expected Username to remain jane, got %s", newState2.Username)
	}
}

func TestBindFieldPerformance(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	initialState := TestState{}

	// Create a large field map
	fieldMap := FieldMap[TestState]{}
	for i := 0; i < 100; i++ {
		fieldName := fmt.Sprintf("field_%d", i)
		fieldMap[fieldName] = func(s TestState, val string) TestState {
			return s
		}
	}

	reducer := fb.BindFieldMap(fieldMap).Build()

	// Benchmark: update fields multiple times
	state := initialState
	for i := 0; i < 1000; i++ {
		fieldName := fmt.Sprintf("field_%d", i%100)
		state = reducer(state, intent.FieldChangeIntent{Field: fieldName, Value: "value"})
	}

	// If this completes without hanging or panicking, the test passes
	t.Log("Performance test passed")
}

func TestGetBuilder(t *testing.T) {
	builder := NewBuilder[TestState]()
	fb := BindField(builder)

	returnedBuilder := fb.GetBuilder()
	if returnedBuilder != builder {
		t.Error("GetBuilder did not return the same builder")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test ParseBool
	if !ParseBool("true") {
		t.Error("ParseBool(true) should return true")
	}
	if !ParseBool("TRUE") {
		t.Error("ParseBool(TRUE) should return true (case insensitive)")
	}
	if ParseBool("false") {
		t.Error("ParseBool(false) should return false")
	}

	// Test ParseInt
	val, err := ParseInt("42")
	if err != nil || val != 42 {
		t.Errorf("ParseInt(42) failed: %v, %d", err, val)
	}

	// Test ParseFloat
	fval, err := ParseFloat("3.14")
	if err != nil || fval != 3.14 {
		t.Errorf("ParseFloat(3.14) failed: %v, %f", err, fval)
	}

	// Test FormatBool
	if FormatBool(true) != "true" {
		t.Errorf("FormatBool(true) returned %s", FormatBool(true))
	}

	// Test FormatInt
	if FormatInt(42) != "42" {
		t.Errorf("FormatInt(42) returned %s", FormatInt(42))
	}
}

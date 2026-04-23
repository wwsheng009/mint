package intent

// GlobalIntent marks intents that should be handled by the global Intent Runtime.
//
// An intent that implements this interface:
// - IsGlobal() = true: Send to global Intent Runtime (for State/Reducer/FieldMap)
// - IsGlobal() = false: Bubble locally through Parent() (for component communication)
//
// Default Behavior:
// - Intents that DO NOT implement this interface are treated as global intents
// - This ensures backward compatibility with existing code
//
// Example (Global Intent - for State/Reducer):
//
//	type SubmitIntent struct{}
//	func (SubmitIntent) IntentType() string { return "Submit" }
//	func (SubmitIntent) IsGlobal() bool { return true }  // explicit, optional
//
// Example (Local Intent - for component communication):
//
//	type ChildSelectIntent struct {
//	    SelectedValue string
//	}
//	func (ChildSelectIntent) IntentType() string { return "ChildSelect" }
//	func (ChildSelectIntent) IsGlobal() bool { return false }  // must implement
//
// Usage in Components:
// 1. Global Intents (FieldChange, Submit, etc.): Send to global Runtime
//    → Handled by Reducer/FieldMap handlers
//    → Update State and trigger UI update
//
// 2. Local Intents (SelectChange, ButtonPress, etc.): Bubble to parent
//    → Handled by parent's HandleIntent()
//    → Component internal logic and event delegation
//
// This interface enables optional Intent Bubble functionality without breaking
// existing functionality.
type GlobalIntent interface {
	Intent
	IsGlobal() bool
}

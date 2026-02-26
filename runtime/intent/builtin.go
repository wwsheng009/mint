package intent

// =============================================================================
// Built-in Intent Types
// =============================================================================

// --- Navigation Intents ---

// NavigateIntent requests navigation to a new location.
type NavigateIntent struct {
	// Path is the target path.
	Path string

	// Params are optional navigation parameters.
	Params map[string]interface{}
}

// IntentType implements Intent.
func (NavigateIntent) IntentType() string {
	return "Navigate"
}

// Priority implements PriorityAware.
func (NavigateIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// --- State Intents ---

// SetStateIntent requests a state update.
type SetStateIntent struct {
	// Key is the state key.
	Key string

	// Value is the new value.
	Value interface{}
}

// IntentType implements Intent.
func (SetStateIntent) IntentType() string {
	return "SetState"
}

// Priority implements PriorityAware.
func (SetStateIntent) Priority() ActionPriority {
	return PriorityNormal
}

// ToggleIntent toggles a boolean state.
type ToggleIntent struct {
	// Key is the state key.
	Key string
}

// IntentType implements Intent.
func (ToggleIntent) IntentType() string {
	return "Toggle"
}

// Priority implements PriorityAware.
func (ToggleIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// IncrementIntent increments a numeric state by a delta.
// This is useful for counter-like operations without closures.
type IncrementIntent struct {
	// Key is the state key.
	Key string

	// Delta is the amount to add (can be negative for decrement).
	Delta int
}

// IntentType implements Intent.
func (IncrementIntent) IntentType() string {
	return "Increment"
}

// Priority implements PriorityAware.
func (IncrementIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// --- UI Intents ---

// OpenModalIntent requests opening a modal.
type OpenModalIntent struct {
	// ModalID is the identifier of the modal to open.
	ModalID string

	// Data is optional data to pass to the modal.
	Data interface{}
}

// IntentType implements Intent.
func (OpenModalIntent) IntentType() string {
	return "OpenModal"
}

// Priority implements PriorityAware.
func (OpenModalIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// CloseModalIntent requests closing a modal.
type CloseModalIntent struct {
	// ModalID is the identifier of the modal to close.
	ModalID string

	// Result is optional result data from the modal.
	Result interface{}
}

// IntentType implements Intent.
func (CloseModalIntent) IntentType() string {
	return "CloseModal"
}

// Priority implements PriorityAware.
func (CloseModalIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// ShowTooltipIntent requests showing a tooltip.
type ShowTooltipIntent struct {
	// Content is the tooltip content.
	Content string

	// TargetID is the target element ID.
	TargetID string
}

// IntentType implements Intent.
func (ShowTooltipIntent) IntentType() string {
	return "ShowTooltip"
}

// Priority implements PriorityAware.
func (ShowTooltipIntent) Priority() ActionPriority {
	return PriorityNormal
}

// HideTooltipIntent requests hiding a tooltip.
type HideTooltipIntent struct{}

// IntentType implements Intent.
func (HideTooltipIntent) IntentType() string {
	return "HideTooltip"
}

// --- Focus Intents ---

// FocusIntent requests focusing an element.
type FocusIntent struct {
	// TargetID is the element to focus.
	TargetID string
}

// IntentType implements Intent.
func (FocusIntent) IntentType() string {
	return "Focus"
}

// Priority implements PriorityAware.
func (FocusIntent) Priority() ActionPriority {
	return PriorityImmediate
}

// BlurIntent requests removing focus from an element.
type BlurIntent struct {
	// TargetID is the element to blur.
	TargetID string
}

// IntentType implements Intent.
func (BlurIntent) IntentType() string {
	return "Blur"
}

// Priority implements PriorityAware.
func (BlurIntent) Priority() ActionPriority {
	return PriorityImmediate
}

// --- Data Intents (Transition) ---

// LoadDataIntent requests loading data asynchronously.
// This is a Transition intent that can be interrupted.
type LoadDataIntent struct {
	// URL is the data source URL.
	URL string

	// Key is where to store the loaded data.
	Key string
}

// IntentType implements Intent.
func (LoadDataIntent) IntentType() string {
	return "LoadData"
}

// IsTransition implements TransitionIntent.
func (LoadDataIntent) IsTransition() bool {
	return true
}

// RefreshIntent requests refreshing data.
type RefreshIntent struct {
	// Keys are the data keys to refresh.
	Keys []string
}

// IntentType implements Intent.
func (RefreshIntent) IntentType() string {
	return "Refresh"
}

// IsTransition implements TransitionIntent.
func (RefreshIntent) IsTransition() bool {
	return true
}

// --- Form Intents ---

// SubmitFormIntent requests form submission.
type SubmitFormIntent struct {
	// FormID is the form identifier.
	FormID string

	// Data is the form data.
	Data map[string]interface{}
}

// IntentType implements Intent.
func (SubmitFormIntent) IntentType() string {
	return "SubmitForm"
}

// Priority implements PriorityAware.
func (SubmitFormIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// ValidateFormIntent requests form validation.
type ValidateFormIntent struct {
	// FormID is the form identifier.
	FormID string
}

// IntentType implements Intent.
func (ValidateFormIntent) IntentType() string {
	return "ValidateForm"
}

// Priority implements PriorityAware.
func (ValidateFormIntent) Priority() ActionPriority {
	return PriorityNormal
}

// --- Action Intents ---

// ClickIntent represents a click action.
type ClickIntent struct {
	// TargetID is the clicked element.
	TargetID string
}

// IntentType implements Intent.
func (ClickIntent) IntentType() string {
	return "Click"
}

// Priority implements PriorityAware.
func (ClickIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// PressIntent represents a press action (Enter key, etc).
type PressIntent struct {
	// TargetID is the pressed element.
	TargetID string
}

// IntentType implements Intent.
func (PressIntent) IntentType() string {
	return "Press"
}

// Priority implements PriorityAware.
func (PressIntent) Priority() ActionPriority {
	return PriorityUserBlocking
}

// =============================================================================
// Intent Constructors
// =============================================================================

// Navigate creates a NavigateIntent.
func Navigate(path string) NavigateIntent {
	return NavigateIntent{Path: path}
}

// NavigateWithParams creates a NavigateIntent with params.
func NavigateWithParams(path string, params map[string]interface{}) NavigateIntent {
	return NavigateIntent{Path: path, Params: params}
}

// SetState creates a SetStateIntent.
func SetState(key string, value interface{}) SetStateIntent {
	return SetStateIntent{Key: key, Value: value}
}

// Toggle creates a ToggleIntent.
func Toggle(key string) ToggleIntent {
	return ToggleIntent{Key: key}
}

// Increment creates an IncrementIntent.
func Increment(key string, delta int) IncrementIntent {
	return IncrementIntent{Key: key, Delta: delta}
}

// Decrement creates an IncrementIntent with negative delta.
func Decrement(key string, delta int) IncrementIntent {
	return IncrementIntent{Key: key, Delta: -delta}
}

// OpenModal creates an OpenModalIntent.
func OpenModal(modalID string) OpenModalIntent {
	return OpenModalIntent{ModalID: modalID}
}

// OpenModalWithData creates an OpenModalIntent with data.
func OpenModalWithData(modalID string, data interface{}) OpenModalIntent {
	return OpenModalIntent{ModalID: modalID, Data: data}
}

// CloseModal creates a CloseModalIntent.
func CloseModal(modalID string) CloseModalIntent {
	return CloseModalIntent{ModalID: modalID}
}

// CloseModalWithResult creates a CloseModalIntent with result.
func CloseModalWithResult(modalID string, result interface{}) CloseModalIntent {
	return CloseModalIntent{ModalID: modalID, Result: result}
}

// Focus creates a FocusIntent.
func Focus(targetID string) FocusIntent {
	return FocusIntent{TargetID: targetID}
}

// Blur creates a BlurIntent.
func Blur(targetID string) BlurIntent {
	return BlurIntent{TargetID: targetID}
}

// LoadData creates a LoadDataIntent.
func LoadData(url, key string) LoadDataIntent {
	return LoadDataIntent{URL: url, Key: key}
}

// SubmitForm creates a SubmitFormIntent.
func SubmitForm(formID string, data map[string]interface{}) SubmitFormIntent {
	return SubmitFormIntent{FormID: formID, Data: data}
}

// Click creates a ClickIntent.
func Click(targetID string) ClickIntent {
	return ClickIntent{TargetID: targetID}
}

// Press creates a PressIntent.
func Press(targetID string) PressIntent {
	return PressIntent{TargetID: targetID}
}

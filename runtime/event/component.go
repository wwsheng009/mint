package event

// Component is the interface for event handling components.
//
// Components implement this interface to handle events in a capture-target-bubble
// propagation model. The HandleEvent method is called during the target phase.
//
// Returns true if the event was handled (stops further propagation),
// false to continue propagation.
type Component interface {
	// HandleEvent processes an event.
	//
	// Implementations can:
	// - Return true to stop event propagation
	// - Return false to allow the event to continue bubbling
	// - Modify the event state (e.g., call PreventDefault())
	HandleEvent(ev Event) bool
}

// EventComponent is the interface for components that handle events.
// This serves as a named marker instead of using anonymous interfaces in type assertions.
type EventComponent interface {
	Component
}

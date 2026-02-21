package panel

// Panel is a composition-based component that delegates to Border + Stack.
// No separate Instance is needed - the Border component handles all runtime state.
//
// The VNode.CreateInstance() method delegates to the composed Border's CreateInstance().

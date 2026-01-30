// Package testing provides testing utilities for DevTools.
package testing

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// Fixture provides a test environment with automatic cleanup.
type Fixture struct {
	Runtime  *MockRuntime
	DevTools *MockDevTools
	t        testing.TB
}

// Setup creates a new test fixture.
// The fixture will be automatically cleaned up when the test completes.
func Setup(t *testing.T) *Fixture {
	t.Helper()

	rt := NewMockRuntime()
	dt := NewMockDevTools() // Already enabled

	return &Fixture{
		Runtime:  rt,
		DevTools: dt,
		t:        t,
	}
}

// Teardown cleans up the test fixture.
func (f *Fixture) Teardown() {
	f.t.Helper()

	f.DevTools.Disable()
	f.DevTools.DevTools.Shutdown()
}

// Cleanup is called automatically via t.Cleanup.
func (f *Fixture) Cleanup() {
	f.t.Helper()
	f.Teardown()
}

// NewFixtureWithConfig creates a fixture with custom configuration.
func NewFixtureWithConfig(t *testing.T, config FixtureConfig) *Fixture {
	t.Helper()

	fixture := Setup(t)
	fixture.Cleanup()

	return fixture
}

// FixtureConfig holds configuration for test fixtures.
type FixtureConfig struct {
	EnableDevTools     bool
	EnableMutationTap  bool
	AutoCleanup        bool
	Timeout            time.Duration
}

// DefaultFixtureConfig returns default fixture configuration.
func DefaultFixtureConfig() FixtureConfig {
	return FixtureConfig{
		EnableDevTools:    true,
		EnableMutationTap: false,
		AutoCleanup:       true,
		Timeout:           30 * time.Second,
	}
}

// ScenarioBuilder helps build test scenarios.
type ScenarioBuilder struct {
	name        string
	description string
	steps       []ScenarioStep
}

// ScenarioStep represents a single step in a test scenario.
type ScenarioStep struct {
	Name       string
	Actions    []TestAction
	Assertions []TestAssertion
	WaitFor    time.Duration
}

// TestAction is an action that can be executed in a test.
type TestAction interface {
	Execute(*Fixture) error
}

// TestAssertion is an assertion that can be verified in a test.
type TestAssertion interface {
	Verify(*Fixture) error
}

// NewScenarioBuilder creates a new scenario builder.
func NewScenarioBuilder(name, description string) *ScenarioBuilder {
	return &ScenarioBuilder{
		name:        name,
		description: description,
		steps:       make([]ScenarioStep, 0),
	}
}

// AddStep adds a step to the scenario.
func (b *ScenarioBuilder) AddStep(name string, actions []TestAction, assertions []TestAssertion) *ScenarioBuilder {
	b.steps = append(b.steps, ScenarioStep{
		Name:       name,
		Actions:    actions,
		Assertions: assertions,
	})
	return b
}

// AddStepWithWait adds a step with a wait duration.
func (b *ScenarioBuilder) AddStepWithWait(name string, wait time.Duration, actions []TestAction, assertions []TestAssertion) *ScenarioBuilder {
	b.steps = append(b.steps, ScenarioStep{
		Name:       name,
		Actions:    actions,
		Assertions: assertions,
		WaitFor:    wait,
	})
	return b
}

// Build builds the scenario.
func (b *ScenarioBuilder) Build() *Scenario {
	return &Scenario{
		Name:        b.name,
		Description: b.description,
		Steps:       b.steps,
	}
}

// Scenario represents a test scenario.
type Scenario struct {
	Name        string
	Description string
	Steps       []ScenarioStep
}

// Run executes the scenario with the given fixture.
func (s *Scenario) Run(t *testing.T, fixture *Fixture) error {
	t.Helper()

	for _, step := range s.Steps {
		t.Run(step.Name, func(t *testing.T) {
			// Execute actions
			for _, action := range step.Actions {
				if err := action.Execute(fixture); err != nil {
					t.Errorf("Action failed: %v", err)
				}
			}

			// Wait if specified
			if step.WaitFor > 0 {
				time.Sleep(step.WaitFor)
			}

			// Verify assertions
			for _, assertion := range step.Assertions {
				if err := assertion.Verify(fixture); err != nil {
					t.Errorf("Assertion failed: %v", err)
				}
			}
		})
	}

	return nil
}

// Helper function to create mock layout result
func CreateMockLayoutResult(nodeCount int) *devtools.LayoutResultAdapter {
	return devtools.MockLayoutResult(nodeCount)
}

// Helper function to create mock layout result with dynamic nodes
func CreateMockDynamicLayoutResult(nodeCount int, frame int) *devtools.LayoutResultAdapter {
	return devtools.MockLayoutResultWithDynamicNodes(nodeCount, frame)
}

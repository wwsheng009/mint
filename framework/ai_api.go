package framework

import (
	"errors"
	"time"

	aiservice "github.com/wwsheng009/mint/internal/ai/service"
	"github.com/wwsheng009/mint/runtime/state"
)

type AIComponentInfo = aiservice.ComponentInfo
type AIStateQuery = aiservice.StateQuery
type AIDirection = aiservice.Direction
type AIWaitCondition = aiservice.WaitCondition
type AIBatchOperation = aiservice.BatchOperation
type AIBatchResult = aiservice.BatchResult

const (
	AIDirectionNext  = aiservice.DirectionNext
	AIDirectionPrev  = aiservice.DirectionPrev
	AIDirectionFirst = aiservice.DirectionFirst
	AIDirectionLast  = aiservice.DirectionLast
	AIDirectionUp    = aiservice.DirectionUp
	AIDirectionDown  = aiservice.DirectionDown
	AIDirectionLeft  = aiservice.DirectionLeft
	AIDirectionRight = aiservice.DirectionRight
)

// AIController exposes host-level AI inspection helpers.
// It is intentionally small for now and will grow alongside internal/ai features.
type AIController struct {
	app *App
}

// AIController returns the current AI controller facade if AI support is enabled.
func (a *App) AIController() *AIController {
	if a == nil || a.aiService == nil {
		return nil
	}
	return &AIController{app: a}
}

// Inspect returns the latest semantic UI snapshot captured after render.
func (c *AIController) Inspect() (*state.Snapshot, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.Inspect()
}

func (c *AIController) Find(selector string) ([]AIComponentInfo, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.Find(selector)
}

func (c *AIController) Query(query AIStateQuery) (map[string]interface{}, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.QueryState(query)
}

func (c *AIController) WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.WaitUntil(condition, timeout)
}

func (c *AIController) WaitFor(condition AIWaitCondition, timeout time.Duration) (*state.Snapshot, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.WaitFor(condition, timeout)
}

func (c *AIController) GetValue(locator string) (interface{}, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.GetValue(locator)
}

func (c *AIController) SetValue(locator string, value interface{}) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.SetValue(locator, value)
}

func (c *AIController) SetProp(locator, key string, value interface{}) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.SetProp(locator, key, value)
}

func (c *AIController) Click(locator string) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.Click(locator)
}

func (c *AIController) Input(locator, text string) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.Input(locator, text)
}

func (c *AIController) Dispatch(locator, actionType string, payload interface{}) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.Dispatch(locator, actionType, payload)
}

func (c *AIController) Navigate(direction AIDirection) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.Navigate(direction)
}

func (c *AIController) GetTree(kind string) (interface{}, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.GetTree(kind)
}

func (c *AIController) GetNode(locator string) (interface{}, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.GetNode(locator)
}

func (c *AIController) GetFormData(locator string) (map[string]interface{}, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.GetFormData(locator)
}

func (c *AIController) SetFormField(locator, field string, value interface{}) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.SetFormField(locator, field, value)
}

func (c *AIController) Select(locator string, value interface{}) error {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return errors.New("AI service is not enabled")
	}
	return c.app.aiService.Select(locator, value)
}

func (c *AIController) Batch(operations []AIBatchOperation, stopOnError bool, dryRun bool) (*AIBatchResult, error) {
	if c == nil || c.app == nil || c.app.aiService == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return c.app.aiService.Batch(operations, stopOnError, dryRun)
}

// Status returns the current AI subsystem status.
func (c *AIController) Status() AIStatus {
	if c == nil || c.app == nil {
		return AIStatus{}
	}
	return c.app.AIStatus()
}

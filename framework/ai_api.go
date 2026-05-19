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

func (c *AIController) service() (*aiservice.Service, error) {
	if c == nil || c.app == nil {
		return nil, errors.New("AI service is not enabled")
	}
	service := c.app.getAIService()
	if service == nil {
		return nil, errors.New("AI service is not enabled")
	}
	return service, nil
}

// AIController returns the current AI controller facade if AI support is enabled.
func (a *App) AIController() *AIController {
	if a == nil || a.getAIService() == nil {
		return nil
	}
	return &AIController{app: a}
}

// Inspect returns the latest semantic UI snapshot captured after render.
func (c *AIController) Inspect() (*state.Snapshot, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.Inspect()
}

func (c *AIController) Find(selector string) ([]AIComponentInfo, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.Find(selector)
}

func (c *AIController) Query(query AIStateQuery) (map[string]interface{}, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.QueryState(query)
}

func (c *AIController) WaitUntil(condition func(*state.Snapshot) bool, timeout time.Duration) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.WaitUntil(condition, timeout)
}

func (c *AIController) WaitFor(condition AIWaitCondition, timeout time.Duration) (*state.Snapshot, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.WaitFor(condition, timeout)
}

func (c *AIController) GetValue(locator string) (interface{}, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.GetValue(locator)
}

func (c *AIController) SetValue(locator string, value interface{}) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.SetValue(locator, value)
}

func (c *AIController) SetProp(locator, key string, value interface{}) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.SetProp(locator, key, value)
}

func (c *AIController) Click(locator string) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.Click(locator)
}

func (c *AIController) Input(locator, text string) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.Input(locator, text)
}

func (c *AIController) Dispatch(locator, actionType string, payload interface{}) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.Dispatch(locator, actionType, payload)
}

func (c *AIController) Navigate(direction AIDirection) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.Navigate(direction)
}

func (c *AIController) GetTree(kind string) (interface{}, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.GetTree(kind)
}

func (c *AIController) GetNode(locator string) (interface{}, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.GetNode(locator)
}

func (c *AIController) GetFormData(locator string) (map[string]interface{}, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.GetFormData(locator)
}

func (c *AIController) SetFormField(locator, field string, value interface{}) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.SetFormField(locator, field, value)
}

func (c *AIController) Select(locator string, value interface{}) error {
	service, err := c.service()
	if err != nil {
		return err
	}
	return service.Select(locator, value)
}

func (c *AIController) Batch(operations []AIBatchOperation, stopOnError bool, dryRun bool) (*AIBatchResult, error) {
	service, err := c.service()
	if err != nil {
		return nil, err
	}
	return service.Batch(operations, stopOnError, dryRun)
}

// Status returns the current AI subsystem status.
func (c *AIController) Status() AIStatus {
	if c == nil || c.app == nil {
		return AIStatus{}
	}
	return c.app.AIStatus()
}

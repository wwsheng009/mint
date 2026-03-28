package framework

import (
	"context"
	"errors"
	"fmt"
)

var errInvokeStopped = errors.New("app main loop stopped")

type invokeRequest struct {
	ctx    context.Context
	fn     func() (any, error)
	result chan invokeResult
}

type invokeResult struct {
	value any
	err   error
}

// Invoke executes fn on the app's main loop when the app is running.
// If the app is not yet running, fn executes inline.
func (a *App) Invoke(ctx context.Context, fn func() (any, error)) (any, error) {
	if fn == nil {
		return nil, errors.New("nil invoke function")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Before the app enters the main loop there is nothing to serialize against,
	// so execute inline.
	if a == nil || a.invokeQ == nil || !a.IsRunning() {
		return fn()
	}

	req := invokeRequest{
		ctx:    ctx,
		fn:     fn,
		result: make(chan invokeResult, 1),
	}

	select {
	case a.invokeQ <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.invokeDone:
		return nil, errInvokeStopped
	}

	select {
	case res := <-req.result:
		return res.value, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.invokeDone:
		return nil, errInvokeStopped
	}
}

func (a *App) handleInvokeRequest(req invokeRequest) {
	if req.result == nil {
		return
	}
	if req.fn == nil {
		req.result <- invokeResult{err: errors.New("nil invoke function")}
		return
	}
	if req.ctx != nil {
		select {
		case <-req.ctx.Done():
			req.result <- invokeResult{err: req.ctx.Err()}
			return
		default:
		}
	}

	var (
		value any
		err   error
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invoke panic: %v", r)
		}
		req.result <- invokeResult{value: value, err: err}
	}()

	value, err = req.fn()
}

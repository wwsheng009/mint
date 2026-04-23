package tooltip

import (
	"github.com/wwsheng009/mint/ui/components/toast"
)

// Toast type re-exports for backward compatibility.
// The canonical types are now in ui/components/toast.

// ToastType re-exported from toast package.
type ToastType = toast.ToastType

const (
	ToastInfo    = toast.ToastInfo
	ToastSuccess = toast.ToastSuccess
	ToastWarning = toast.ToastWarning
	ToastError   = toast.ToastError
)

// ToastVNode re-exported from toast package.
type ToastVNode = toast.ToastVNode

// ToastBuilder re-exported from toast package.
type ToastBuilder = toast.ToastBuilder

// ToastManager re-exported from toast package (canonical name is toast.Manager).
type ToastManager = toast.Manager

// NewToast creates a new Toast VNode (re-exported from toast package).
func NewToast(message string) *toast.ToastVNode {
	return toast.NewToast(message)
}

// NewToastBuilder creates a new Toast builder (re-exported from toast package).
func NewToastBuilder(message string) *toast.ToastBuilder {
	return toast.NewToastBuilder(message)
}

// NewToastInstance creates a new ToastInstance (re-exported from toast package).
func NewToastInstance(props interface{}) *toast.ToastInstance {
	return toast.NewToastInstance(props.(map[string]interface{}))
}

// NewManager creates a new ToastManager (re-exported from toast package).
func NewManager() *toast.Manager {
	return toast.NewManager()
}

// Info creates an info toast (re-exported from toast package).
func Info(message string) *toast.ToastVNode {
	return toast.Info(message)
}

// Success creates a success toast (re-exported from toast package).
func Success(message string) *toast.ToastVNode {
	return toast.Success(message)
}

// Warning creates a warning toast (re-exported from toast package).
func Warning(message string) *toast.ToastVNode {
	return toast.Warning(message)
}

// Error creates an error toast (re-exported from toast package).
func Error(message string) *toast.ToastVNode {
	return toast.Error(message)
}

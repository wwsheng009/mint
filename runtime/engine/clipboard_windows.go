//go:build windows
// +build windows

package engine

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard       = user32.NewProc("OpenClipboard")
	procCloseClipboard      = user32.NewProc("CloseClipboard")
	procEmptyClipboard      = user32.NewProc("EmptyClipboard")
	procSetClipboardData    = user32.NewProc("SetClipboardData")
	procGlobalAlloc         = kernel32.NewProc("GlobalAlloc")
	procGlobalLock          = kernel32.NewProc("GlobalLock")
	procGlobalUnlock        = kernel32.NewProc("GlobalUnlock")
	procGlobalFree          = kernel32.NewProc("GlobalFree")
	procRtlMoveMemory       = kernel32.NewProc("RtlMoveMemory")
)

const (
	GMEM_MOVEABLE = 0x0002
	CF_UNICODETEXT = 13
)

// copyToClipboardPlatform 复制文本到剪贴板（Windows 实现）
func copyToClipboardPlatform(text string) error {
	// 首先尝试使用 Windows API
	if err := copyWithWindowsAPI(text); err == nil {
		return nil
	}

	// 回退到 PowerShell
	return copyWithPowerShell(text)
}

func copyWithWindowsAPI(text string) error {
	utf16Text, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	// 打开剪贴板
	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard")
	}
	defer procCloseClipboard.Call()

	// 清空剪贴板
	ret, _, _ = procEmptyClipboard.Call()
	if ret == 0 {
		return fmt.Errorf("failed to empty clipboard")
	}

	// 分配全局内存
	size := len(utf16Text) * 2
	hGlobal, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(size))
	if hGlobal == 0 {
		return fmt.Errorf("failed to allocate global memory")
	}
	defer procGlobalFree.Call(hGlobal)

	// 锁定内存并复制数据
	ptr, _, _ := procGlobalLock.Call(hGlobal)
	if ptr == 0 {
		return fmt.Errorf("failed to lock global memory")
	}

	// 复制 UTF-16 数据
	copy((*[1 << 30]uint16)(unsafe.Pointer(ptr))[:len(utf16Text)], utf16Text)

	procGlobalUnlock.Call(hGlobal)

	// 设置剪贴板数据
	ret, _, _ = procSetClipboardData.Call(CF_UNICODETEXT, hGlobal)
	if ret == 0 {
		return fmt.Errorf("failed to set clipboard data")
	}

	return nil
}

func copyWithPowerShell(text string) error {
	// 使用 PowerShell 作为回退方案
	psCmd := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Clipboard]::SetText('%s')`, 
		strings.ReplaceAll(text, "'", "''"))
	
	cmd := exec.Command("powershell.exe", "-Command", psCmd)
	return cmd.Run()
}

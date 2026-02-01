//go:build !windows && !darwin
// +build !windows,!darwin

package engine

import (
	"fmt"
	"os/exec"
	"strings"
)

// copyToClipboardPlatform 复制文本到剪贴板（Linux 实现）
func copyToClipboardPlatform(text string) error {
	// 尝试使用 xclip
	if cmd := exec.Command("xclip", "-selection", "clipboard"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// 尝试使用 xsel
	if cmd := exec.Command("xsel", "--clipboard", "--input"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// 尝试使用 wl-copy (Wayland)
	if cmd := exec.Command("wl-copy"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no clipboard tool available (tried xclip, xsel, wl-copy)")
}

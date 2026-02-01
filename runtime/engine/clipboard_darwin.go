//go:build darwin
// +build darwin

package engine

import (
	"fmt"
	"os/exec"
	"strings"
)

// copyToClipboardPlatform 复制文本到剪贴板（macOS 实现）
func copyToClipboardPlatform(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	
	return nil
}

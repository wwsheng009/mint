//go:build !windows

package mcp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func listenPipe(path string) (net.Listener, string, string, func(), error) {
	if strings.TrimSpace(path) == "" {
		path = filepath.Join(os.TempDir(), fmt.Sprintf("mint-ai-%d.sock", os.Getpid()))
	}

	// Ensure the socket directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, "", "", nil, err
	}

	// Clean up any stale socket file.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, "", "", nil, err
	}

	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, "", "", nil, err
	}

	// Restrict access to the current user by default.
	_ = os.Chmod(path, 0o600)

	base := "unix://" + path
	return ln, base, base + "/mcp", func() { _ = os.Remove(path) }, nil
}

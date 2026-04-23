//go:build windows

package mcp

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Microsoft/go-winio"
)

func listenPipe(path string) (net.Listener, string, string, func(), error) {
	if strings.TrimSpace(path) == "" {
		path = fmt.Sprintf("mint-ai-%d", os.Getpid())
	}
	if !strings.HasPrefix(path, `\\.\pipe\`) {
		path = `\\.\pipe\` + strings.TrimPrefix(path, `\\.\pipe\`)
	}

	ln, err := winio.ListenPipe(path, nil)
	if err != nil {
		return nil, "", "", nil, err
	}

	base := "npipe://" + path
	return ln, base, base + "/mcp", func() {}, nil
}

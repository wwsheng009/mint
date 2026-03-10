package mcp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (s *Server) listen() (net.Listener, string, string, func(), error) {
	transport := strings.ToLower(strings.TrimSpace(s.cfg.Transport))
	if transport == "" {
		transport = "http"
	}
	switch transport {
	case "http":
		return listenHTTP(s.cfg.Host, s.cfg.Port)
	case "pipe":
		return listenPipe(s.cfg.Host)
	default:
		return nil, "", "", nil, fmt.Errorf("unsupported MCP transport: %s", transport)
	}
}

func listenHTTP(host string, port int) (net.Listener, string, string, func(), error) {
	if strings.TrimSpace(host) == "" {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", "", nil, err
	}
	base := "http://" + ln.Addr().String()
	return ln, base, base + "/mcp", func() {}, nil
}

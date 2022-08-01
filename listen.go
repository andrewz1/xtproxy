package xtproxy

import (
	"context"
	"net"
	"syscall"
	"time"
)

const (
	defaultKeepAlive = 10 * time.Second
)

type ListenConfig struct {
	net.ListenConfig
	ctx context.Context
}

var (
	lc = NewListenConfig(context.Background())
)

func NewListenConfig(ctx context.Context) *ListenConfig {
	cfg := &ListenConfig{
		ListenConfig: net.ListenConfig{
			KeepAlive: defaultKeepAlive,
		},
		ctx: ctx,
	}
	cfg.Control = cfg.lControl
	return cfg
}

func (c *ListenConfig) lControl(network, address string, rc syscall.RawConn) error {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return net.UnknownNetworkError(network)
	}
	var fd int
	if err := rc.Control(func(sockFd uintptr) {
		fd = int(sockFd)
	}); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
		return err
	}
	return nil
}

func (c *ListenConfig) Listen(network, address string) (net.Listener, error) {
	return c.ListenConfig.Listen(c.ctx, network, address)
}

func Listen(network, address string) (net.Listener, error) {
	return lc.Listen(network, address)
}

package xtproxy

import (
	"context"
	"net"
	"syscall"
	"time"
)

const (
	defaultKeepAlive = 10 * time.Second
	defaultLingerSec = 5
)

type ListenConfig struct {
	net.ListenConfig
	ctx context.Context
}

type Listener struct {
	net.Listener
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
	ln, err := c.ListenConfig.Listen(c.ctx, network, address)
	if err != nil {
		return nil, err
	}
	return &Listener{Listener: ln}, nil
}

func Listen(network, address string) (net.Listener, error) {
	return lc.Listen(network, address)
}

func (l *Listener) Accept() (net.Conn, error) {
	cn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			cn.Close()
		}
	}()
	tcn, ok := cn.(*net.TCPConn)
	if !ok {
		err = net.UnknownNetworkError(cn.LocalAddr().Network())
		return nil, err
	}
	if err = tcn.SetLinger(defaultLingerSec); err != nil {
		return nil, err
	}
	if err = tcn.SetNoDelay(true); err != nil {
		return nil, err
	}
	return tcn, nil
}

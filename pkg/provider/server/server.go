package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/manager"
	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	Handler      http.Handler
	Certificates *manager.TLSCertificates

	lock    sync.RWMutex
	started bool
	port    int
}

func (h *HTTPServer) listen(ctx context.Context) (net.Listener, error) {
	// use a host allocated port
	lnConfig := net.ListenConfig{}
	ln, err := lnConfig.Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create http listener for http server: %s", err)
	}

	// Wait for port to be assigned first
	time.Sleep(250 * time.Millisecond)

	// acquire the real port
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return nil, fmt.Errorf("failed to split host port from net listener: %s", err)
	}
	h.port, err = strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to identify port from net listener: %s", err)
	}
	logrus.Infof("Listening on port %d", h.port)
	return ln, nil
}

func (h *HTTPServer) serve(ctx context.Context, ln net.Listener) {
	server := http.Server{
		Handler:     h.Handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	if h.Certificates != nil {
		logrus.Debug("mtls is enabled")
		server.TLSConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			MinVersion: tls.VersionTLS12,
		}
		errCh := make(chan error, 1)
		go func() {
			errCh <- server.ServeTLS(ln, h.Certificates.CertFile, h.Certificates.KeyFile)
		}()
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			if err != nil {
				logrus.Fatal(err)
			}
		}
	} else {
		logrus.Warn("mtls has been disabled, running over http")
		errCh := make(chan error, 1)
		go func() {
			errCh <- server.Serve(ln)
		}()
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			if err != nil {
				logrus.Fatal(err)
			}
		}
	}
}

func (h *HTTPServer) Start(ctx context.Context) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.started {
		// already started
		logrus.Warn("Attempted to start server multiple times; not supported")
		return nil
	}
	ln, err := h.listen(ctx)
	if err != nil {
		return err
	}
	go h.serve(ctx, ln)
	h.started = true
	return nil
}

func (h *HTTPServer) Port() int {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.port
}

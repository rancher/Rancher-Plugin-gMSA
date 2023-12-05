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

func NewServer(ctx context.Context, handler http.Handler, certificates *manager.TLSCertificates) *Server {
	h := &Server{}
	server := http.Server{
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	if certificates != nil {
		logrus.Debug("mtls is enabled")
		server.TLSConfig = &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			MinVersion: tls.VersionTLS12,
		}
		h.serve = func(ln net.Listener) {
			errCh := make(chan error, 1)
			go func() {
				errCh <- server.ServeTLS(ln, certificates.CertFile, certificates.KeyFile)
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
	} else {
		logrus.Warn("mtls has been disabled, running over http")
		h.serve = func(ln net.Listener) {
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
	return h
}

type Server struct {
	lock sync.RWMutex

	serve func(net.Listener)

	started bool
	port    int
}

func (h *Server) Start(ctx context.Context) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.started {
		// already started
		return nil
	}
	ln, err := h.listen(ctx)
	if err != nil {
		return err
	}
	go h.serve(ln)
	h.started = true
	return nil
}

func (s *Server) Port() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.port
}

func (h *Server) listen(ctx context.Context) (net.Listener, error) {
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

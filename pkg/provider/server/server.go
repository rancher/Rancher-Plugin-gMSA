package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type HTTPServer struct {
	Handler http.Handler
}

func (h *HTTPServer) StartServer(errChan chan error, namespace string, disableMTLS bool) (string, error) {
	// use a host allocated port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("failed to create http listener for http server: %v", err)
	}

	go func() {
		s := http.Server{
			Handler: h.Handler,
		}

		if disableMTLS {
			err = s.Serve(ln)
		} else {
			s.TLSConfig = &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
				MinVersion: tls.VersionTLS12,
			}
			err = s.ServeTLS(ln, fmt.Sprintf("%s/%s/container/ssl/server/tls.crt", "/var/lib/rancher/gmsa", namespace), fmt.Sprintf("%s/%s/container/ssl/server/tls.crt", "/var/lib/rancher/gmsa", namespace))
		}

		errChan <- fmt.Errorf("HTTP server encountered a fatal error: %v", err.Error())
	}()

	// let the server come up and
	// be assigned a port
	time.Sleep(250 * time.Millisecond)
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return "", fmt.Errorf("failed to split host port from net listener: %v", err)
	}
	logrus.Info("Listening on port ", port)
	return port, nil
}

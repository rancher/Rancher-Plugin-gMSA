package pkg

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HTTPServer struct {
	Engine      *gin.Engine
	Credentials *CredentialClient
}

func (h *HTTPServer) StartServer(errChan chan error, namespace string, disableMTLS bool) (string, error) {
	// use a host allocated port
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("failed to create http listener for http server: %v", err)
	}

	go func() {
		s := http.Server{
			Handler: h.Engine,
		}

		if disableMTLS {
			err = s.Serve(ln)
		} else {
			s.TLSConfig = &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
				MinVersion: tls.VersionTLS12,
			}
			err = s.ServeTLS(ln, fmt.Sprintf(containerServerCrt, gmsaDirectory, namespace), fmt.Sprintf(containerServerKey, gmsaDirectory, namespace))
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

func NewGinServer(h *HTTPServer, debug bool) *gin.Engine {
	e := gin.Default()
	if !debug {
		// gin uses debug mode by default
		gin.SetMode(gin.ReleaseMode)
	}
	e.GET("/provider", h.handle)
	return e
}

func (h *HTTPServer) handle(c *gin.Context) {
	secret := c.GetHeader("object")
	if secret == "" {
		c.Status(http.StatusBadRequest)
		logrus.Info("Received request with no object")
		return
	}

	s, err := h.Credentials.Secrets.Get(secret, metav1.GetOptions{})
	// Handle forbidden requests in the same manner as 404's so no feedback is given to the caller
	if errors.IsForbidden(err) || errors.IsNotFound(err) {
		c.Status(http.StatusNotFound)
		logrus.Warnf("error retrieving secret %s: %v", secret, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Username:   string(s.Data["username"]),
		Password:   string(s.Data["password"]),
		DomainName: string(s.Data["domainName"]),
	})
}

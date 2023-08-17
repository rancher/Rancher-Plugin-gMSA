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

type HttpServer struct {
	Engine      *gin.Engine
	Credentials *CredentialClient
}

func (h *HttpServer) StartServer(errChan chan error, namespace string, debug bool) (string, error) {
	h.Engine.GET("/provider", h.handle)

	// use a host allocated port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", fmt.Errorf("failed to create http listener for http server: %v", err)
	}

	go func() {
		s := http.Server{
			Handler: h.Engine,
		}

		if debug {
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

func NewGinServer(h *HttpServer) *gin.Engine {
	e := gin.Default()
	e.GET("/provider", h.handle)
	return e
}

func (h *HttpServer) handle(c *gin.Context) {
	secret := c.GetHeader("object")
	if secret == "" {
		c.Status(http.StatusBadRequest)
		logrus.Info("Received request with no object")
		return
	}

	s, err := h.Credentials.Secrets.Get(c.GetHeader("object"), metav1.GetOptions{})
	if errors.IsForbidden(err) || errors.IsNotFound(err) {
		c.Status(http.StatusNotFound)
		logrus.Info(err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Username:   string(s.Data["username"]),
		Password:   string(s.Data["password"]),
		DomainName: string(s.Data["domainName"]),
	})
}

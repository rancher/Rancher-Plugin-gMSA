package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/getter"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func NewHandler(secretsGetter getter.NamespacedGeneric[*corev1.Secret]) http.Handler {
	h := handler{
		secrets: secretsGetter,
	}
	engine := gin.Default()
	engine.GET("/provider", func(ctx *gin.Context) {
		secret := ctx.GetHeader("object")
		status, response := h.handle(secret)
		if response == nil {
			ctx.Status(status)
			return
		}
		ctx.JSON(status, response)
	})
	return engine.Handler()
}

type handler struct {
	secrets getter.NamespacedGeneric[*corev1.Secret]
}

func (h *handler) handle(secret string) (code int, response *Response) {
	if secret == "" {
		logrus.Debug("Received request with no object")
		return http.StatusBadRequest, nil
	}
	s, err := h.secrets.Get(secret)
	if err != nil {
		if errors.IsForbidden(err) {
			logrus.Warnf("not allowed to retrieve secret %s: %s", secret, err)
			return http.StatusNotFound, nil
		}
		if errors.IsNotFound(err) {
			logrus.Debugf("secret %s is not found: %s", secret, err)
			return http.StatusNotFound, nil
		}
		logrus.Warnf("error retrieving secret %s: %s", secret, err)
		return http.StatusNotFound, nil
	}
	response, err = ParseResponse(s)
	if err != nil {
		logrus.Debug(err)
		return http.StatusNotFound, nil
	}
	return http.StatusOK, response
}

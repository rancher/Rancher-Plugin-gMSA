package secret

import (
	"context"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/server"
	corecontroller "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

type handler struct{}

func Register(ctx context.Context, secrets corecontroller.SecretController) {
	h := &handler{}
	secrets.OnChange(ctx, "on-ccg-secret", h.OnChange)
}

func (h *handler) OnChange(key string, secret *corev1.Secret) (*corev1.Secret, error) {
	if secret != nil {
		if _, err := server.ParseResponse(secret); err == nil {
			logrus.Debugf("found CCG secret %s", key)
		}
	}
	return secret, nil
}

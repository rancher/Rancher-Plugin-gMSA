package provider

import (
	"context"
	"fmt"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/controllers"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/getter"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/ratelimit"
	v1 "k8s.io/api/core/v1"
)

type CredentialClient struct {
	Secrets getter.NamespacedGeneric[*v1.Secret]
}

func NewClient(namespace string, kubeConfig string) (*CredentialClient, error) {
	cfg := kubeconfig.GetNonInteractiveClientConfig(kubeConfig)
	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("found not create client config: %v", err)
	}
	clientConfig.RateLimiter = ratelimit.None

	secrets, err := controllers.Run(context.TODO(), namespace, clientConfig)
	if err != nil {
		return nil, err
	}

	return &CredentialClient{
		Secrets: getter.Namespaced[*v1.Secret](secrets, namespace),
	}, nil
}

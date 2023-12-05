package provider

import (
	"fmt"

	v1 "github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/generated/norman/core/v1"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/ratelimit"
)

type CredentialClient struct {
	Secrets v1.SecretInterface
}

func NewClient(ns string, kubeConfig string) (*CredentialClient, error) {
	cfg := kubeconfig.GetNonInteractiveClientConfig(kubeConfig)
	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("found not create client config: %v", err)
	}
	clientConfig.RateLimiter = ratelimit.None

	secretCfg, err := v1.NewForConfig(*clientConfig)
	if err != nil {
		return nil, err
	}

	return &CredentialClient{
		Secrets: secretCfg.Secrets(ns),
	}, nil
}

type Response struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DomainName string `json:"domainName"`
}

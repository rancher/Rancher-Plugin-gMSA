package provider

import (
	"context"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/manager"
)

func Clean(ctx context.Context, namespace string) error {
	m := manager.New(namespace)
	return m.Clean(ctx)
}

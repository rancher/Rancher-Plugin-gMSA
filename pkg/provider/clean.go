package provider

import (
	"context"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/manager"
)

func Clean(ctx context.Context, namespace string) error {
	m := manager.New(namespace)
	return m.Clean(ctx)
}

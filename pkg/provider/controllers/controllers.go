package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/server"
	"github.com/rancher/lasso/pkg/cache"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/pkg/generated/controllers/core"
	corecontroller "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/k8scheck"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
)

type appContext struct {
	Core corecontroller.Interface

	starters []start.Starter
}

func (a *appContext) start(ctx context.Context) error {
	return start.All(ctx, 50, a.starters...)
}

func Run(ctx context.Context, namespace string, client *rest.Config) (corecontroller.SecretCache, error) {
	appCtx, err := newContext(ctx, namespace, client)
	if err != nil {
		return nil, err
	}

	// Note: This controller does not utilize a lock it's expected that multiple are running (one per node).
	// If this changes, make sure you add locking logic here.
	if err := k8scheck.Wait(ctx, *client); err != nil {
		return nil, err
	}

	appCtx.Core.Secret().OnChange(ctx, "watch-secrets", func(key string, secret *corev1.Secret) (*corev1.Secret, error) {
		if secret != nil {
			if _, err := server.ParseResponse(secret); err == nil {
				logrus.Debugf("found CCG secret %s", key)
			}
		}
		return secret, nil
	})

	if err := appCtx.start(ctx); err != nil {
		return nil, err
	}
	logrus.Info("All controllers have been started")

	_, err = appCtx.Core.Namespace().Get(namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot find secret namespace: %s", err)
	}

	return appCtx.Core.Secret().Cache(), nil
}

func controllerFactory(rest *rest.Config, namespace string) (controller.SharedControllerFactory, error) {
	rateLimit := workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 60*time.Second)
	clientFactory, err := client.NewSharedClientFactory(rest, nil)
	if err != nil {
		return nil, err
	}

	cacheFactory := cache.NewSharedCachedFactory(clientFactory, nil)
	return controller.NewSharedControllerFactory(cacheFactory, &controller.SharedControllerFactoryOptions{
		DefaultRateLimiter: rateLimit,
		DefaultWorkers:     50,
	}), nil
}

func newContext(ctx context.Context, namespace string, client *rest.Config) (*appContext, error) {
	scf, err := controllerFactory(client, namespace)
	if err != nil {
		return nil, err
	}

	core, err := core.NewFactoryFromConfigWithOptions(client, &generic.FactoryOptions{
		Namespace:               namespace,
		SharedControllerFactory: scf,
	})
	if err != nil {
		return nil, err
	}
	corev := core.Core().V1()

	return &appContext{
		Core: corev,

		starters: []start.Starter{
			core,
		},
	}, nil
}

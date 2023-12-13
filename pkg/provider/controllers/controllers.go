package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/controllers/secret"
	"github.com/rancher/lasso/pkg/cache"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/pkg/generated/controllers/core"
	corecontroller "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/k8scheck"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	appCtx, err := newContext(namespace, client)
	if err != nil {
		return nil, err
	}

	secret.Register(ctx, appCtx.Core.Secret())

	go func() {
		if err := k8scheck.Wait(ctx, *client); err != nil {
			panic(err)
		}
		if err := appCtx.start(ctx); err != nil {
			panic(fmt.Errorf("unable to start controllers: %s", err))
		}
		logrus.Info("All controllers have been started")
	}()

	return appCtx.Core.Secret().Cache(), nil
}

func controllerFactory(namespace string, rest *rest.Config) (controller.SharedControllerFactory, error) {
	rateLimit := workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 60*time.Second)
	clientFactory, err := client.NewSharedClientFactory(rest, nil)
	if err != nil {
		return nil, err
	}
	kindNamespace := map[schema.GroupVersionKind]string{
		{Group: "", Version: "v1", Kind: "Secret"}: namespace,
	}
	cacheFactory := cache.NewSharedCachedFactory(clientFactory, &cache.SharedCacheFactoryOptions{
		KindNamespace: kindNamespace,
	})
	return controller.NewSharedControllerFactory(cacheFactory, &controller.SharedControllerFactoryOptions{
		DefaultRateLimiter: rateLimit,
		DefaultWorkers:     50,
	}), nil
}

func newContext(namespace string, client *rest.Config) (*appContext, error) {
	scf, err := controllerFactory(namespace, client)
	if err != nil {
		return nil, err
	}

	core, err := core.NewFactoryFromConfigWithOptions(client, &generic.FactoryOptions{
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

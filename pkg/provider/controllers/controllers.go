package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/controllers/secret"
	"github.com/rancher/lasso/pkg/cache"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/lasso/pkg/metrics"
	"github.com/rancher/wrangler/v3/pkg/generated/controllers/core"
	corecontroller "github.com/rancher/wrangler/v3/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/rancher/wrangler/v3/pkg/k8scheck"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
)

type appContext struct {
	Core corecontroller.Interface

	ControllerFactory controller.SharedControllerFactory
}

func (a *appContext) start(ctx context.Context) error {
	transaction := controller.NewHandlerTransaction(ctx)

	ctx = metrics.WithContextID(ctx, "providercontext")
	if err := a.ControllerFactory.SharedCacheFactory().Start(ctx); err != nil {
		transaction.Rollback()
		return err
	}

	a.ControllerFactory.SharedCacheFactory().WaitForCacheSync(ctx)
	transaction.Commit()

	if err := a.ControllerFactory.Start(ctx, 50); err != nil {
		return err
	}

	return nil
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
	rateLimit := workqueue.NewTypedItemExponentialFailureRateLimiter[any](5*time.Millisecond, 60*time.Second)
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

	coreFactory, err := core.NewFactoryFromConfigWithOptions(client, &generic.FactoryOptions{
		SharedControllerFactory: scf,
	})
	if err != nil {
		return nil, err
	}

	return &appContext{
		ControllerFactory: scf,
		Core:              coreFactory.Core().V1(),
	}, nil
}

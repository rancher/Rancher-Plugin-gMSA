package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/controllers"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/getter"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/manager"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/server"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type Opts struct {
	Namespace     string
	ForcedPort    string
	DisableMTLS   bool
	SkipArtifacts bool
}

func Run(ctx context.Context, client *rest.Config, opts Opts) error {
	logrus.Infof("Starting controllers")
	secretCache, err := controllers.Run(ctx, opts.Namespace, client)
	if err != nil {
		return err
	}
	return run(ctx, secretCache, opts)
}

func run(ctx context.Context, secrets getter.Generic[*corev1.Secret], opts Opts) error {
	var tlsCertificates *manager.TLSCertificates
	if !opts.DisableMTLS && !opts.SkipArtifacts {
		logrus.Infof("Setting up certificates")
		m := manager.New(opts.Namespace)
		if err := m.Start(ctx); err != nil {
			return fmt.Errorf("failed to start certificate manager: %s", err)
		}
		tlsCertificates = m.ServerCertificates()
	}

	logrus.Infof("Starting server")
	server := server.HTTPServer{
		Handler:      server.NewHandler(getter.Namespaced(secrets, opts.Namespace)),
		Certificates: tlsCertificates,
		ForcePort:    opts.ForcedPort,
	}
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	if !opts.SkipArtifacts {
		// TODO: Adjust Directory Permissions
		baseDir := filepath.Join(utils.ProviderDirectory, opts.Namespace)
		if err := utils.CreateDirectory(baseDir); err != nil {
			return err
		}
		// Creating port file
		portFile := filepath.Join(baseDir, "port.txt")
		port := []byte(fmt.Sprintf("%d", server.Port()))

		logrus.Infof("Creating %s", portFile)
		if err := utils.SetFile(portFile, port); err != nil {
			return err
		}
	}

	<-ctx.Done()
	return nil
}

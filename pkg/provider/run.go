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

func Run(ctx context.Context, client *rest.Config, namespace string, disableMTLS, skipArtifacts bool) error {
	logrus.Infof("Starting controllers")
	secretCache, err := controllers.Run(ctx, namespace, client)
	if err != nil {
		return err
	}
	return run(ctx, secretCache, namespace, disableMTLS, skipArtifacts)
}

func run(ctx context.Context, secrets getter.Generic[*corev1.Secret], namespace string, disableMTLS, skipArtifacts bool) error {
	var tlsCertificates *manager.TLSCertificates
	if !disableMTLS && !skipArtifacts {
		logrus.Infof("Setting up certificates")
		m := manager.New(namespace)
		if err := m.Start(ctx); err != nil {
			return fmt.Errorf("failed to start certificate manager: %s", err)
		}
		tlsCertificates = m.ServerCertificates()
	}

	logrus.Infof("Starting server")
	server := server.HTTPServer{
		Handler:      server.NewHandler(getter.Namespaced(secrets, namespace)),
		Certificates: tlsCertificates,
	}
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	if !skipArtifacts {
		// TODO: Adjust Directory Permissions
		baseDir := filepath.Join(utils.ProviderDirectory, namespace)
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

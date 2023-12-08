package main

import (
	"context"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/controllers"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/getter"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/version"
	"github.com/gin-gonic/gin"
	command "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/ratelimit"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"

	"fmt"
)

var (
	debugConfig command.DebugConfig
)

func main() {
	cmd := &cobra.Command{
		Use:     "gmsa-account-provider",
		Version: version.FriendlyVersion(),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.AddCommand(
		command.AddDebug(command.Command(&GMSAAccountProvider{}, cobra.Command{
			Use:          "run",
			Short:        "Start the account provider api",
			SilenceUsage: true,
		}), &debugConfig),
		command.AddDebug(command.Command(&GMSAAccountProviderCleanup{}, cobra.Command{
			Use:          "cleanup",
			Short:        "Remove all files and certificates for the account provider instance",
			SilenceUsage: true,
		}), &debugConfig),
	)

	command.Main(cmd)
}

type GMSAAccountProvider struct {
	Kubeconfig    string `usage:"Kubeconfig file" env:"KUBECONFIG"`
	Namespace     string `usage:"Namespace to watch for Secrets" default:"cattle-windows-gmsa-system" env:"NAMESPACE"`
	DisableMTLS   bool   `usage:"Disable mTLS" default:"false" env:"DISABLE_MTLS"`
	SkipArtifacts bool   `usage:"Prevents any files from being written to the host. Implicitly disables mTLS." default:"false" env:"DISABLE_ARTIFACTS"`
}

func (a *GMSAAccountProvider) Run(cmd *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()
	if !debugConfig.Debug {
		// gin uses debug mode by default
		gin.SetMode(gin.ReleaseMode)
	}

	if err := utils.ValidateNamespace(a.Namespace); err != nil {
		return err
	}

	cfg := kubeconfig.GetNonInteractiveClientConfig(a.Kubeconfig)
	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return fmt.Errorf("found not create client config: %v", err)
	}
	clientConfig.RateLimiter = ratelimit.None

	secrets, err := controllers.Run(context.TODO(), a.Namespace, clientConfig)
	if err != nil {
		return err
	}

	server := provider.HTTPServer{
		Secrets: getter.Namespaced[*v1.Secret](secrets, a.Namespace),
	}
	server.Engine = provider.NewGinServer(&server)

	if !a.SkipArtifacts {
		// create all the files and directories we need on the host
		err = provider.CreateDynamicDirectory(a.Namespace)
		if err != nil {
			return fmt.Errorf("failed to create dynamic directory: %v", err)
		}

		err = provider.WriteCerts(a.Namespace)
		if err != nil {
			return fmt.Errorf("failed to write mTLS certificates to host: %v", err)
		}
	}

	errChan := make(chan error)
	port, err := server.StartServer(errChan, a.Namespace, a.DisableMTLS || a.SkipArtifacts)
	if err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	if !a.SkipArtifacts {
		err = provider.WritePortFile(a.Namespace, port)
		if err != nil {
			return fmt.Errorf("failed to create dynamic directory: %v", err)
		}
	}

	// block on http server error
	// or command context completion
	select {
	case err = <-errChan:
		return err
	case <-cmd.Context().Done():
		return nil
	}
}

type GMSAAccountProviderCleanup struct {
	Namespace string `usage:"Namespace to watch for Secrets" default:"cattle-windows-gmsa-system" env:"NAMESPACE"`
}

func (a *GMSAAccountProviderCleanup) Run(_ *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

	if err := utils.ValidateNamespace(a.Namespace); err != nil {
		return err
	}

	return provider.CleanupProvider(a.Namespace)
}

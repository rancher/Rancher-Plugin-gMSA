package main

import (
	"github.com/gin-gonic/gin"
	cli "github.com/rancher/Rancher-Plugin-gMSA/cmd/util"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/version"
	"github.com/rancher/wrangler/v3/pkg/kubeconfig"
	"github.com/rancher/wrangler/v3/pkg/ratelimit"
	"github.com/spf13/cobra"
)

var (
	debugConfig cli.DebugConfig
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
		cli.Command(&GMSAAccountProvider{}, cobra.Command{
			Use:          "run",
			Short:        "Start the account provider api",
			SilenceUsage: true,
		}, &debugConfig),
		cli.Command(&GMSAAccountProviderCleanup{}, cobra.Command{
			Use:          "cleanup",
			Short:        "Remove all files and certificates for the account provider instance",
			SilenceUsage: true,
		}, &debugConfig),
	)

	cli.Main(cmd)
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
	client, err := cfg.ClientConfig()
	if err != nil {
		return err
	}
	client.RateLimiter = ratelimit.None

	if err = provider.Run(cmd.Context(), client, provider.Opts{
		Namespace:     a.Namespace,
		ForcedPort:    "",
		DisableMTLS:   a.DisableMTLS,
		SkipArtifacts: a.SkipArtifacts,
	}); err != nil {
		return err
	}

	<-cmd.Context().Done()
	return nil
}

type GMSAAccountProviderCleanup struct {
	Namespace string `usage:"Namespace to watch for Secrets" default:"cattle-windows-gmsa-system" env:"NAMESPACE"`
}

func (a *GMSAAccountProviderCleanup) Run(cmd *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

	if err := utils.ValidateNamespace(a.Namespace); err != nil {
		return err
	}

	return provider.Clean(cmd.Context(), a.Namespace)
}

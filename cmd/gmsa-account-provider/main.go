package main

import (
	pkg "github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/plugin/provider"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/version"
	command "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"

	"fmt"
	"strings"
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
	Namespace     string `usage:"Namespace to watch for Secrets" default:"cattle-gmsa-system" env:"NAMESPACE"`
	DisableMTLS   bool   `usage:"Disable mTLS" default:"false" env:"DISABLE_MTLS"`
	SkipArtifacts bool   `usage:"Prevents any files from being written to the host. Implicitly disables mTLS." default:"false" env:"DISABLE_ARTIFACTS"`
}

func (a *GMSAAccountProvider) Run(cmd *cobra.Command, _ []string) error {
	if a.Namespace == "" {
		return fmt.Errorf("gmsa-account-provider must be started within a kubernetes namespace")
	}

	if len(strings.Split(a.Namespace, " ")) > 1 {
		return fmt.Errorf("rancher-gmsa-account-provider can only be started in a single namespace")
	}

	// cli debug
	debugConfig.MustSetupDebug()

	client, err := pkg.NewClient(a.Namespace, a.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to setup client: %v", err)
	}

	server := pkg.HTTPServer{
		Credentials: client,
	}
	server.Engine = pkg.NewGinServer(&server, debugConfig.Debug)

	if !a.SkipArtifacts {
		// create all the files and directories we need on the host
		err = pkg.CreateDynamicDirectory(a.Namespace)
		if err != nil {
			return fmt.Errorf("failed to create dynamic directory: %v", err)
		}

		err = pkg.WriteCerts(a.Namespace)
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
		err = pkg.WritePortFile(a.Namespace, port)
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
	Namespace string `usage:"Namespace to watch for Secrets" default:"cattle-gmsa-system" env:"NAMESPACE"`
}

func (a *GMSAAccountProviderCleanup) Run(_ *cobra.Command, _ []string) error {
	if a.Namespace == "" {
		return fmt.Errorf("gmsa-account-provider must be run within a kubernetes namespace")
	}

	if len(strings.Split(a.Namespace, " ")) > 1 {
		return fmt.Errorf("rancher-gmsa-account-provider can only be run within a single namespace")
	}

	return pkg.CleanupProvider(a.Namespace)
}

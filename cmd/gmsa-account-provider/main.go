package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	pkg "github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/plugin/provider"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/version"
	command "github.com/rancher/wrangler-cli"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debugConfig command.DebugConfig
)

type GMSAAccountProvider struct {
	Kubeconfig    string `usage:"Kubeconfig file" env:"KUBECONFIG"`
	Namespace     string `usage:"Namespace to watch for Secrets" default:"cattle-gmsa-system" env:"NAMESPACE"`
	DisableMTLS   bool   `usage:"Disable mTLS" default:"false" env:"DISABLE_MTLS"`
	SkipArtifacts bool   `usage:"Prevents any files from being written to the host. Implicitly disables mTLS." default:"false" env:"DISABLE_ARTIFACTS"`
}

func (a *GMSAAccountProvider) Run(cmd *cobra.Command, _ []string) error {
	if a.Namespace == "" {
		return fmt.Errorf("rancher-gmsa-account-provider can only be started in a single namespace")
	}

	// pprof and cli debug
	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			logrus.Errorf("could not start pprof: %v", err)
		}
	}()
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

func main() {
	cmd := command.Command(&GMSAAccountProvider{}, cobra.Command{
		Version: version.FriendlyVersion(),
	})
	cmd = command.AddDebug(cmd, &debugConfig)
	command.Main(cmd)
}

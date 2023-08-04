package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/version"
	command "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/ratelimit"
	"github.com/spf13/cobra"
)

var (
	debugConfig command.DebugConfig
)

type GMSAAccountProvider struct {
	Kubeconfig string `usage:"Kubeconfig file" env:"KUBECONFIG"`
	Namespace  string `usage:"Namespace to watch for Secrets" default:"cattle-gmsa-system" env:"NAMESPACE"`
}

func (a *GMSAAccountProvider) Run(cmd *cobra.Command, args []string) error {
	if len(a.Namespace) == 0 {
		return fmt.Errorf("rancher-gmsa-account-provider can only be started in a single namespace")
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	debugConfig.MustSetupDebug()

	cfg := kubeconfig.GetNonInteractiveClientConfig(a.Kubeconfig)
	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return err
	}
	clientConfig.RateLimiter = ratelimit.None

	// ctx := cmd.Context()
	// TODO: Add entrypoint commands here

	<-cmd.Context().Done()

	return nil
}

func main() {
	cmd := command.Command(&GMSAAccountProvider{}, cobra.Command{
		Version: version.FriendlyVersion(),
	})
	cmd = command.AddDebug(cmd, &debugConfig)
	command.Main(cmd)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	pkg "github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/plugin/provider"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/version"
	command "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
)

var (
	debugConfig command.DebugConfig
)

type GMSAAccountProvider struct {
	Kubeconfig string `usage:"Kubeconfig file" env:"KUBECONFIG"`
	Namespace  string `usage:"Namespace to watch for Secrets" default:"cattle-gmsa-system" env:"NAMESPACE"`
}

func (a *GMSAAccountProvider) Run(cmd *cobra.Command, _ []string) error {
	if len(a.Namespace) != 1 {
		return fmt.Errorf("rancher-gmsa-account-provider can only be started in a single namespace")
	}

	// pprof and cli debug
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	debugConfig.MustSetupDebug()

	namespace := os.Getenv("NAMESPACE")

	controller, err := pkg.NewClient(namespace, a.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to setup client: %v", err)
	}

	server := pkg.HttpServer{
		Credentials: controller,
	}
	server.Engine = pkg.NewGinServer(&server)

	// create all the files and directories we need on the host
	// these calls will be no-ops on OS's other than Windows

	err = pkg.CreateDir(namespace)
	if err != nil {
		return fmt.Errorf("failed to create dynamic directory: %v", err)
	}

	err = pkg.WriteCerts(namespace)
	if err != nil {
		return fmt.Errorf("failed to write mTLS certificates to host: %v", err)
	}

	errChan := make(chan error)
	port, err := server.StartServer(errChan, namespace, debugConfig.Debug)
	if err != nil {
		return fmt.Errorf("failed to start http server: %v", err)
	}

	err = pkg.WritePortFile(namespace, port)
	if err != nil {
		return fmt.Errorf("failed to create dynamic directory: %v", err)
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

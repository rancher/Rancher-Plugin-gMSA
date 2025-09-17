package main

import (
	_ "net/http/pprof"
	"time"

	cli "github.com/rancher/Rancher-Plugin-gMSA/cmd/util"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/installer"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debugConfig cli.DebugConfig
)

func main() {
	cmd := &cobra.Command{
		Use:     "ccg-plugin-installer",
		Version: version.FriendlyVersion(),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cmd.AddCommand(
		cli.Command(&CCGPluginInstaller{}, cobra.Command{
			Use:          "install",
			Aliases:      []string{"upgrade"},
			Short:        "Installs or upgrades the Rancher CCG Plugin as a DLL on your host",
			SilenceUsage: true,
		}, &debugConfig),
		cli.Command(&CCGPluginUninstaller{}, cobra.Command{
			Use:          "uninstall",
			Short:        "Uninstall the Rancher CCG Plugin",
			SilenceUsage: true,
		}, &debugConfig),
	)
	cli.Main(cmd)
}

type CCGPluginInstaller struct {
	Timeout int `usage:"Specify a timeout after executing main operation" default:"0" env:"CCG_PLUGIN_INSTALLER_TIMEOUT"`
}

func (i *CCGPluginInstaller) Run(_ *cobra.Command, _ []string) error {
	err := installer.Install()
	executeTimeout(i.Timeout)
	return err
}

type CCGPluginUninstaller struct {
	Timeout int `usage:"Specify a timeout after executing main operation" default:"0" env:"CCG_PLUGIN_INSTALLER_TIMEOUT"`
}

func (i *CCGPluginUninstaller) Run(_ *cobra.Command, _ []string) error {
	err := installer.Uninstall()
	executeTimeout(i.Timeout)
	return err
}

func executeTimeout(timeout int) {
	if timeout <= 0 {
		return
	}
	logrus.Infof("Sleeping for %d seconds before exiting...", timeout)
	time.Sleep(time.Duration(timeout) * time.Second)
}

package main

import (
	_ "net/http/pprof"
	"time"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/installer"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/version"
	command "github.com/rancher/wrangler-cli"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debugConfig command.DebugConfig
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
		command.AddDebug(command.Command(&CCGPluginInstaller{}, cobra.Command{
			Use:          "install",
			Aliases:      []string{"upgrade"},
			Short:        "Installs or upgrades the Rancher CCG Plugin as a DLL on your host",
			SilenceUsage: true,
		}), &debugConfig),
		command.AddDebug(command.Command(&CCGPluginUninstaller{}, cobra.Command{
			Use:          "uninstall",
			Short:        "Uninstall the Rancher CCG Plugin",
			SilenceUsage: true,
		}), &debugConfig),
	)
	command.Main(cmd)
}

type CCGPluginInstaller struct {
	Timeout int `usage:"Specify a timeout after executing main operation" default:"0" env:"CCG_PLUGIN_INSTALLER_TIMEOUT"`
}

func (i *CCGPluginInstaller) Run(_ *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

	err := installer.Install()
	executeTimeout(i.Timeout)
	return err
}

type CCGPluginUninstaller struct {
	Timeout int `usage:"Specify a timeout after executing main operation" default:"0" env:"CCG_PLUGIN_INSTALLER_TIMEOUT"`
}

func (i *CCGPluginUninstaller) Run(_ *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

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

package main

import (
	_ "net/http/pprof"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/plugin/manager"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/version"
	command "github.com/rancher/wrangler-cli"
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
			Short:        "Install the Rancher CCG Plugin as a DLL on your host",
			SilenceUsage: true,
		}), &debugConfig),
		command.AddDebug(command.Command(&CCGPluginUninstaller{}, cobra.Command{
			Use:          "uninstall",
			Short:        "Uninstall the Rancher CCG Plugin",
			SilenceUsage: true,
		}), &debugConfig),
		command.AddDebug(command.Command(&CCGPluginUpgrader{}, cobra.Command{
			Use:          "upgrade",
			Short:        "Upgrade the Rancher CCG Plugin",
			SilenceUsage: true,
		}), &debugConfig),
	)
	command.Main(cmd)
}

type CCGPluginInstaller struct {
}

func (i *CCGPluginInstaller) Run(_ *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

	return manager.Install()
}

type CCGPluginUninstaller struct {
}

func (i *CCGPluginUninstaller) Run(_ *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

	return manager.Uninstall()
}

type CCGPluginUpgrader struct {
}

func (i *CCGPluginUpgrader) Run(_ *cobra.Command, _ []string) error {
	debugConfig.MustSetupDebug()

	return manager.Upgrade()
}

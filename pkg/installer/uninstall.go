package installer

import (
	"fmt"
	"path/filepath"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/installer/embedded"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/installer/status"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
)

func Uninstall() error {
	installationStatus, err := status.CheckInstallationStatus(embedded.DLL)
	if err != nil {
		if err != status.ErrNotWindows {
			return fmt.Errorf("unable to determine current status of plugin: %s", err)
		}
		// only put a warning out if this is not a Windows host
		logrus.Warnf("%s", status.ErrNotWindows)
		return nil
	}

	if installationStatus.Uninstalled() {
		logrus.Infof(installationStatus.String())
		return nil
	}

	logrus.Infof("Uninstalling plugin...")

	if installationStatus.RequiresDeregistration() {
		if err := utils.SetFile(filepath.Join(utils.DLLDirectory, utils.DLLUninstallScriptPath), embedded.UninstallScript); err != nil {
			return err
		}
		logrus.Infof("Successfully wrote uninstallation script to disk")

		err := utils.RunPowershell(filepath.Join(utils.DLLDirectory, utils.DLLUninstallScriptPath))
		if err != nil {
			return fmt.Errorf("failed to execute uninstallation script: %v", err)
		}
		logrus.Infof("Successfully ran uninstallation script")
	}

	if installationStatus.RequiresDirectoryDeletion() {
		if err := utils.DeleteDirectory(utils.DLLDirectory); err != nil {
			return err
		}
		logrus.Infof("Successfully removed DLL directory")
	}

	installationStatus, err = status.CheckInstallationStatus(embedded.DLL)
	if err != nil {
		return fmt.Errorf("unable to determine final status of plugin: %s", err)
	}
	if installationStatus.Uninstalled() {
		logrus.Infof(installationStatus.String())
		return nil
	}
	return installationStatus.Error()
}

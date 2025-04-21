package installer

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/installer/embedded"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/installer/status"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
)

func Install() (err error) {
	installationStatus, err := status.CheckInstallationStatus(embedded.DLL)
	if err != nil {
		if !errors.Is(err, status.ErrNotWindows) {
			return fmt.Errorf("unable to determine current status of plugin: %s", err)
		}
		// only put a warning out if this is not a Windows host
		logrus.Warnf("%s", status.ErrNotWindows)
	}

	// Check if already up-to-date
	if installationStatus.Installed() {
		logrus.Infof(installationStatus.String())
		return nil
	}

	logrus.Infof("Installing plugin...")
	// Create directory if necessary
	if installationStatus.RequiresDirectoryCreation() {
		if err := utils.CreateDirectory(utils.DLLDirectory); err != nil {
			return err
		}
	}

	if installationStatus.RequiresUpgrade() {
		// Some useful docs regarding upgrades
		//  https://serverfault.com/questions/503721/replacing-dll-files-while-the-application-is-running
		//	https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates
		logrus.Debug("Moving existing CCG Plugin DLL to a temporary file")
		undoFunc, deleteFunc, err := utils.RenameTempFile(filepath.Join(utils.DLLDirectory, utils.DLLPath))
		if err != nil {
			return nil
		}
		defer func() {
			logrus.Debugf("Removing old DLL")
			deleteErr := deleteFunc()
			if deleteErr != nil {
				logrus.Errorf("%s", deleteErr)
			}
		}()
		defer func() {
			if err == nil {
				return
			}
			logrus.Debugf("Reverting DLL back to previous state")
			undoErr := undoFunc()
			if undoErr != nil {
				logrus.Errorf("%s", undoErr)
			}
		}()
	}

	if installationStatus.RequiresInstall() {
		if err := utils.SetFile(filepath.Join(utils.DLLDirectory, utils.DLLPath), embedded.DLL); err != nil {
			return err
		}
		logrus.Infof("Successfully wrote CCG Plugin DLL to disk")
	}

	if installationStatus.RequiresRegistration() {
		if err := utils.SetFile(filepath.Join(utils.DLLDirectory, utils.DLLInstallScriptPath), embedded.InstallScript); err != nil {
			return err
		}
		logrus.Infof("Successfully wrote installation script to disk")

		err := utils.RunPowershell(filepath.Join(utils.DLLDirectory, utils.DLLInstallScriptPath))
		if err != nil {
			return fmt.Errorf("failed to execute installation script: %v", err)
		}
		logrus.Infof("Successfully ran installation script")
	}

	installationStatus, err = status.CheckInstallationStatus(embedded.DLL)
	if err != nil {
		if !errors.Is(err, status.ErrNotWindows) {
			return fmt.Errorf("unable to determine final status of plugin: %s", err)
		}
	}
	if installationStatus.Installed() {
		logrus.Infof(installationStatus.String())
		return nil
	}
	return installationStatus.Error()
}

//go:build windows

package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/installer/embedded"
	"github.com/sirupsen/logrus"
)

func Install() error {
	directoryExists, fileExists, CCGEntryExists, ClassesRootKeyExists, err := verifyInstall()
	if err != nil {
		return fmt.Errorf("failed to detect installation status: %v", err)
	}

	if directoryExists && fileExists && CCGEntryExists && ClassesRootKeyExists {
		logrus.Info("Plugin already installed")
		return nil
	}

	logrus.Info("Beginning installation")

	if !fileExists {
		err = writeArtifacts()
		if err != nil {
			return fmt.Errorf("failed to write artifacts: %v", err)
		}
	}

	err = executeInstaller()
	if err != nil {
		return fmt.Errorf("failed to execute installation script: %v", err)
	}

	directoryExists, fileExists, CCGEntryExists, ClassesRootKeyExists, err = verifyInstall()
	if err != nil {
		return fmt.Errorf("failed to detect installation status: %v", err)
	}

	successfullyInstalled := true
	if !directoryExists {
		logrus.Infof("error encountered during installation: directory %s does not exist", baseDir)
		successfullyInstalled = false
	}

	if !fileExists {
		logrus.Infof("error encountered during installation: file %s does not exist", dllFilePath())
		successfullyInstalled = false
	}

	if !CCGEntryExists {
		logrus.Info("error encountered during installation: was not able to add plugin CLSID to CCG COM Classes Key")
		successfullyInstalled = false
	}

	if !ClassesRootKeyExists {
		logrus.Info("error encountered during installation: was not able to add plugin CLSID to HKEY_CLASSES_ROOT")
		successfullyInstalled = false
	}

	if !successfullyInstalled {
		return fmt.Errorf("failed to install plugin")
	}

	logrus.Info("Installation successful!")

	return nil
}

func writeArtifacts() error {
	err := os.Mkdir(baseDir, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create base directory: %v", err)
	}
	logrus.Infof("successfully created base directory %s", baseDir)

	err = os.WriteFile(dllFilePath(), embedded.DLL, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write dll file: %v", err)
	}
	logrus.Infof("successfully wrote plugin dll to disk")

	err = os.WriteFile(installScriptFilePath(), embedded.InstallScript, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write install script: %v", err)
	}
	logrus.Info("successfully wrote installer script to disk")

	return nil
}

func executeInstaller() error {
	// run installation command
	cmd := exec.Command("powershell.exe", "-File", installScriptFilePath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Info(string(out))
		return fmt.Errorf("failed to execute installation script: %v", err)
	}

	return nil
}

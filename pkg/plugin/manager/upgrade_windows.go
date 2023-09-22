package manager

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Some useful docs regarding upgrades
//  https://serverfault.com/questions/503721/replacing-dll-files-while-the-application-is-running
//	https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates
//
// upgrading shouldn't require any registry changes, as the GUID's will stay the same.

func Upgrade() error {
	upgradeNeeded, err := needsUpgrade()
	if err != nil {
		return fmt.Errorf("error encountered while determining upgradability: %v", err)
	}

	if !upgradeNeeded {
		logrus.Infof("Existing DLL and new DLL match, no upgrade required")
		return nil
	}

	err = renameDll()
	if err != nil {
		return fmt.Errorf("could not rename existing DLL during upgrade process: %v", err)
	}

	err = writeNewDll()
	if err != nil {
		return fmt.Errorf("could not write new DLL during upgrade process")
	}

	logrus.Infof("upgrade complete!")

	return nil
}

// needsUpgrade checks if there are any differences between the old DLL and the new DLL
// by directly comparing their contents
func needsUpgrade() (bool, error) {
	directoryExists, fileExists, CCGEntryExists, ClassesRootKeyExists, err := verifyInstall()
	if err != nil {
		return false, fmt.Errorf("failed to detect installation status during upgrade: %v", err)
	}

	if !directoryExists && !fileExists && !CCGEntryExists && !ClassesRootKeyExists {
		logrus.Warnf("could not find an existing plugin to upgrade, has the plugin been installed yet?")
		return false, nil
	}

	// this function needs recovery logic. If we somehow hit an error after renaming the file
	// then the node would have no gmsa functionality.
	logrus.Infof("determining if an upgrade is needed...")
	f, err := os.ReadFile(dllFilePath())
	if err != nil {
		return false, fmt.Errorf("could not read existing dll file: %v", err)
	}

	if !bytes.Equal(f, dll) {
		logrus.Infof("outdated DLL detected, beginning upgrade...")
		return true, nil
	}

	return false, nil
}

// renameDll will remove any old DLL files if they exist, rename the current DLL file,
// and write the updated embedded DLL file.
func renameDll() error {
	logrus.Infof("Renaming existing DLL file (%s -> %s)", dllFilePath(), outdatedDllFilePath())
	err := os.Remove(outdatedDllFilePath())
	if err != nil && !strings.Contains(err.Error(), "cannot find the file") {
		return fmt.Errorf("failed to delete outdated DLL: %v", err)
	}

	err = os.Rename(dllFilePath(), outdatedDllFilePath())
	if err != nil {
		return fmt.Errorf("could not rename existing DLL file during upgrade: %v", err)
	}
	logrus.Infof("successfully renamed existing dll file")

	return nil
}

// writeNewDll writes the embedded DLL to the host
func writeNewDll() error {
	logrus.Infof("writing updated dll to disk...")
	err := os.WriteFile(dllFilePath(), dll, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write dll file: %v", err)
	}
	logrus.Infof("successfully wrote updated plugin dll to disk")
	return nil
}

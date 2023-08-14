package manager

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

//go:embed RanchergMSACredentialProvider.dll
var dll []byte

//go:embed install-plugin.ps1
var installer []byte

//go:embed uninstall-plugin.ps1
var uninstaller []byte

//go:embed cleanup.ps1
var cleanup []byte

// Some useful docs regarding upgrades
//  https://serverfault.com/questions/503721/replacing-dll-files-while-the-application-is-running
//	https://learn.microsoft.com/en-us/windows/win32/dlls/dynamic-link-library-updates

const (
	// baseDir is where we expect the dll to live
	baseDir = "C:\\Program Files\\RanchergMSACredentialProvider"
	// CCGCOMClassKey is the Windows Registry Key which is used by CCG to invoke the dll
	CCGCOMClassKey = "SYSTEM\\CurrentControlSet\\Control\\CCG\\COMClasses\\{e4781092-f116-4b79-b55e-28eb6a224e26}"
	// ClassesRootKey is the Windows Registry Key which is added by regsvc upon registering the dll, and is also used to invoke the dll
	ClassesRootKey = "CLSID\\{E4781092-F116-4B79-B55E-28EB6A224E26}"

	installFileName   = "install-plugin.ps1"
	uninstallFileName = "uninstall-plugin.ps1"
	cleanupFileName   = "cleanup-gmsa-plugin.ps1"

	dllFileName = "RanchergMSACredentialProvider.dll"
	tlbFileName = "RanchergMSACredentialProvider.tlb"
)

func Install() error {
	installed, err := alreadyInstalled()
	if err != nil {
		return fmt.Errorf("failed to detect installation status: %v", err)
	}

	if installed {
		logrus.Info("plugin already installed")
		return nil
	}

	logrus.Info("beginning installation")

	err = writeArtifacts()
	if err != nil {
		return fmt.Errorf("failed to write artifacts: %v", err)
	}

	err = executeInstaller()
	if err != nil {
		return fmt.Errorf("failed to execute installation script: %v", err)
	}

	logrus.Info("Installation successful!")
	return nil
}

func Uninstall() error {
	logrus.Infof("attempting to uninstall plugin on Windows")
	installed, err := alreadyInstalled()
	if err != nil {
		return fmt.Errorf("failed to determine installation status: %v", err)
	}

	if !installed {
		logrus.Infof("Did not find anything to uninstall")
		return nil
	}

	logrus.Info("Beginning uninstallation process...")
	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, uninstallFileName), uninstaller, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write install script: %v", err)
	}

	logrus.Info("executing uninstallation script...")
	cmd := exec.Command("powershell.exe", "-File", fmt.Sprintf("%s\\%s", baseDir, uninstallFileName))
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute uninstallation script: %v", err)
	}

	logrus.Info("successfully executed uninstallation script")
	logrus.Info("attempting to remove DLL and tlb: ", dllFileName)
	// continuously try to remove the actual files.
	// We retry this process a few times because there may
	// still be instances of CCG referencing the DLL. Windows
	// will prevent the file from being deleted if any references
	// still exist. Eventually, the CCG instances will terminate and
	// all references will disappear, at which point the file can be
	// deleted.
	//
	// It goes without saying that if you're uninstalling this plugin,
	// you shouldn't be running workloads which need to use the plugin.
	//
	// This can and should be improved.
	successfulRemoval := false
	for i := 0; i < 10; i++ {
		err = os.Remove(dllFileName)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			logrus.Info("encountered error removing tlb directory, some CCG instances may still be referencing the plugin. Will retry in 1 minute")
			time.Sleep(1 * time.Minute)
			continue
		}
		err = os.Remove(tlbFileName)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			logrus.Info("encountered error removing tlb , some CCG instances may still be referencing the plugin. Will retry in 1 minute")
			time.Sleep(1 * time.Minute)
			break
		}
		if err == nil {
			successfulRemoval = true
			break
		}
	}

	if !successfulRemoval {
		logrus.Infof("ERROR: Failed to remove DLL directory: %v\n", err)
	} else {
		logrus.Info("DLL and tlb removal complete")
	}

	// write cleanup script to host
	err = os.WriteFile(cleanupFileName, cleanup, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write cleanup script to host: %v", err)
	}
	logrus.Infof("Removal successful! To removal all plugin artifacts please run %s manually", fmt.Sprintf("%s\\%s", baseDir, cleanupFileName))

	return nil
}

func alreadyInstalled() (bool, error) {
	// 1. Check that the DLL exists in the expected directory C:\Program Files\RanchergMSACredentialProvider
	_, err := os.Stat(fmt.Sprintf(baseDir))
	directoryDoesNotExist := err != nil

	_, err = os.Stat(fmt.Sprintf("%s\\%s", baseDir, dllFileName))
	fileDoesNotExist := err != nil

	// 2. Check the registry for a key in HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}
	CCGEntryExists, err := CCGCOMClassExists(CCGCOMClassKey)
	if err != nil {
		return false, fmt.Errorf("failed to query CCG Com class key: %v", err)
	}

	// 3. Check the CLSID HKEY_CLASSES_ROOT\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}
	ClassesRootKeyExists, err := CLSIDExists(ClassesRootKey)
	if err != nil {
		return false, fmt.Errorf("failed to query CLSID registry key: %v", err)
	}

	return !directoryDoesNotExist && !fileDoesNotExist && CCGEntryExists && ClassesRootKeyExists, nil
}

func writeArtifacts() error {
	err := os.Mkdir(baseDir, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create base directory: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, dllFileName), dll, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write dll file: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s\\%s", baseDir, installFileName), installer, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to write install script: %v", err)
	}

	return nil
}

func executeInstaller() error {
	// run installation command
	cmd := exec.Command("powershell.exe", "-File", fmt.Sprintf("%s\\%s", baseDir, installFileName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Info(string(out))
		return fmt.Errorf("failed to execute installation script: %v", err)
	}

	return nil
}

// CCGCOMClassExists is used to get the ccg com entry
func CCGCOMClassExists(registryKey string) (bool, error) {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(registry.LOCAL_MACHINE, registryKey, access)
	if err != nil {
		if err != registry.ErrNotExist {
			return false, fmt.Errorf("failed to open LOCAL_MACHINE registry key %s while determining COM Class existence: %v", registryKey, err)
		}
		return false, nil
	}

	return true, nil
}

// CLSIDExists is used to get the CLSID value
func CLSIDExists(registryKey string) (bool, error) {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(registry.CLASSES_ROOT, registryKey, access)
	if err != nil {
		if err != registry.ErrNotExist {
			return false, fmt.Errorf("failed to open CLASSES_ROOT registry key %s while determining CLSID existence: %v", registryKey, err)
		}
		return false, nil
	}

	return true, nil
}

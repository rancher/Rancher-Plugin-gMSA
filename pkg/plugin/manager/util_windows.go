//go:build windows

package manager

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

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

func verifyInstall() (bool, bool, bool, bool, error) {
	// 1. Check that the DLL directory exists C:\Program Files\RanchergMSACredentialProvider
	_, err := os.Stat(fmt.Sprintf(baseDir))
	directoryDoesNotExist := err != nil

	// 2. Check that the DLL exists inside the directory
	_, err = os.Stat(fmt.Sprintf("%s\\%s", baseDir, dllFileName))
	fileDoesNotExist := err != nil

	// 3. Check the registry for a key in HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\CCG\COMClasses\{e4781092-f116-4b79-b55e-28eb6a224e26}
	CCGEntryExists, err := CCGCOMClassExists(CCGCOMClassKey)
	if err != nil {
		return false, false, false, false, fmt.Errorf("failed to query CCG Com class key: %v", err)
	}

	// 4. Check the CLSID HKEY_CLASSES_ROOT\CLSID\{E4781092-F116-4B79-B55E-28EB6A224E26}
	ClassesRootKeyExists, err := CLSIDExists(ClassesRootKey)
	if err != nil {
		return false, false, false, false, fmt.Errorf("failed to query CLSID registry key: %v", err)
	}

	return !directoryDoesNotExist, !fileDoesNotExist, CCGEntryExists, ClassesRootKeyExists, nil
}

func dllFilePath() string {
	return fmt.Sprintf("%s\\%s", baseDir, dllFileName)
}

func outdatedDllFilePath() string {
	return fmt.Sprintf("%s\\%s", baseDir, oldDllFileName)
}

func tlbFileFilePath() string {
	return fmt.Sprintf("%s\\%s", baseDir, tlbFileName)
}

func installScriptFilePath() string {
	return fmt.Sprintf("%s\\%s", baseDir, installFileName)
}

func uninstallScriptFilePath() string {
	return fmt.Sprintf("%s\\%s", baseDir, uninstallFileName)
}

func cleanUpScriptFilePath() string {
	return fmt.Sprintf("%s\\%s", baseDir, cleanupFileName)
}

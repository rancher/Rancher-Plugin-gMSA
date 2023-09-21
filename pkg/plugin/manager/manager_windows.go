//go:build windows

package manager

import (
	_ "embed"
)

//go:embed RanchergMSACredentialProvider.dll
var dll []byte

//go:embed install-plugin.ps1
var installer []byte

//go:embed uninstall-plugin.ps1
var uninstaller []byte

//go:embed cleanup.ps1
var cleanup []byte

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

	dllFileName    = "RanchergMSACredentialProvider.dll"
	oldDllFileName = "out-dated-version-RanchergMSACredentialProvider.dll"
	tlbFileName    = "RanchergMSACredentialProvider.tlb"
)

// see install_windows.go, uninstall_windows.go, and upgrade_windows.go for further implementation

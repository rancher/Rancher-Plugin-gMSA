package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// DLLPath is where the DLL will be stored within the DLLDirectory
	DLLPath = "RanchergMSACredentialProvider.dll"

	// DLLInstallScriptPath is where DLL install script will be stored within the DLLDirectory
	DLLInstallScriptPath = "install-plugin.ps1"

	// DLLUninstallScriptPat is where the DLL uninstall script will be stored within the DLLDirectory
	DLLUninstallScriptPath = "uninstall-plugin.ps1"

	// DLLGuid is the GUID that we register the Rancher CCG Plugin DLL under
	//
	// Developer Note:
	// This value should never be changed unless absolutely necessary since the DLL GUID is
	// used to create GMSACredentialSpec resources. Changing this value would immediately render
	// all user-created GMSACredentialSpecs that were targeting this CCG Plugin no longer apply to
	// this installation of the CCG Plugin.
	DLLGuid = "E4781092-F116-4B79-B55E-28EB6A224E26"

	// CCGCOMClassKey is the Windows Registry Key which is used by CCG to invoke the DLL
	CCGCOMClassesKey = filepath.Join(
		"SYSTEM",
		"CurrentControlSet",
		"Control",
		"CCG",
		"COMClasses",
		fmt.Sprintf("{%s}", strings.ToLower(DLLGuid)),
	)

	// ClassesRootKey is the Windows Registry Key which is added by regsvc upon registering the dll. It is also used to invoke the DLL
	CLSIDKey = filepath.Join(
		"CLSID",
		fmt.Sprintf("{%s}", DLLGuid),
	)
)

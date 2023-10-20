package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	DLLGuid = "E4781092-F116-4B79-B55E-28EB6A224E26"

	// DLLPath is where the DLL will be stored
	DLLPath = "RanchergMSACredentialProvider.dll"

	// DLLInstallScriptPath is where DLL install script will be stored
	DLLInstallScriptPath = "install-plugin.ps1"

	// DLLUninstallScriptPat is where the DLL uninstall script will be stored
	DLLUninstallScriptPath = "uninstall-plugin.ps1"

	// CCGCOMClassKey is the Windows Registry Key which is used by CCG to invoke the dll
	CCGCOMClassesKey = filepath.Join(
		"SYSTEM",
		"CurrentControlSet",
		"Control",
		"CCG",
		"COMClasses",
		fmt.Sprintf("{%s}", strings.ToLower(DLLGuid)),
	)

	// ClassesRootKey is the Windows Registry Key which is added by regsvc upon registering the dll, and is also used to invoke the dll
	CLSIDKey = filepath.Join(
		"CLSID",
		fmt.Sprintf("{%s}", DLLGuid),
	)
)

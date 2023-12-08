package embedded

import (
	_ "embed"
)

//go:embed RanchergMSACredentialProvider.dll
var DLL []byte

//go:embed scripts/install-plugin.ps1
var InstallScript []byte

//go:embed scripts/uninstall-plugin.ps1
var UninstallScript []byte

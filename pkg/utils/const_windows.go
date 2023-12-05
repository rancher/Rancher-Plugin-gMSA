//go:build windows

package utils

import (
	"path/filepath"
)

const (
	Windows = true
)

var (
	// DLLDirectory contains resources for the CCG Plugin DLL
	DLLDirectory = filepath.Join("C:\\", "Program Files", "RanchergMSACredentialProvider")

	// ProviderDirectory contains namespace directories for CCG Plugin Account Providers
	ProviderDirectory = filepath.Join("C:\\", "var", "lib", "rancher", "gmsa")
)

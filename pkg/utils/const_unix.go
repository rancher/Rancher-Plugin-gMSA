//go:build !windows

package utils

import (
	"path/filepath"
)

const (
	Windows = false
)

var (
	// DLLDirectory contains resources for the CCG Plugin DLL
	DLLDirectory = filepath.Join("dist", "rancher-plugin-gmsa", "ccg-plugin-installer")

	// ProviderDirectory contains namespace directories for CCG Plugin Account Providers
	ProviderDirectory = filepath.Join("dist", "rancher-plugin-gmsa", "gmsa-account-provider")
)

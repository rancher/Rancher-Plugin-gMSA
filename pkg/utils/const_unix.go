//go:build !windows

package utils

import (
	"path/filepath"
)

const (
	Windows = false
)

var (
	// DLLDirectory is where all CCG Plugin resources will be stored
	DLLDirectory = filepath.Join("dist", "rancher-plugin-gmsa", "ccg-dll-installer")

	ProviderDirectory = filepath.Join("dist", "rancher-plugin-gmsa", "account-provider")
)

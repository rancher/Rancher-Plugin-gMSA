//go:build windows

package utils

import (
	"path/filepath"
)

const (
	Windows = true
)

var (
	// DLLDirectory is where all CCG Plugin resources will be stored
	DLLDirectory = filepath.Join("C:", "Program Files", "RanchergMSACredentialProvider")
)

var (
	ProviderDirectory = filepath.Join("C:", "var", "lib", "rancher", "gmsa")
)

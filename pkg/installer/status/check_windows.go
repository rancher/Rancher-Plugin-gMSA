//go:build windows

package status

import (
	"errors"
	"fmt"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"golang.org/x/sys/windows/registry"
)

// CCGCOMClassExists is used to get the ccg com entry
func ccgCOMClassExists(registryKey string) (bool, error) {
	if utils.DryRun {
		return DummyCcgCOMClassExists, nil
	}
	return ensurePathExistsInRegistry(registry.LOCAL_MACHINE, registryKey)
}

// CLSIDExists is used to get the CLSID value
func clsidExists(registryKey string) (bool, error) {
	if utils.DryRun {
		return DummyCLSIDExists, nil
	}
	return ensurePathExistsInRegistry(registry.CLASSES_ROOT, registryKey)
}

func ensurePathExistsInRegistry(key registry.Key, path string) (bool, error) {
	var access uint32 = registry.QUERY_VALUE
	_, err := registry.OpenKey(key, path, access)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("failed to open registry key %s while determining CLSID existence: %v", path, err)
}

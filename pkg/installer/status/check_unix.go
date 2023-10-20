//go:build !windows

package status

// CCGCOMClassExists is used to get the ccg com entry
func ccgCOMClassExists(registryKey string) (bool, error) {
	return DummyCcgCOMClassExists, nil
}

// CLSIDExists is used to get the CLSID value
func clsidExists(registryKey string) (bool, error) {
	return DummyCLSIDExists, nil
}

//go:build !windows

package status

// CCGCOMClassExists is used to get the ccg com entry
func ccgCOMClassExists(_ string) (bool, error) {
	return DummyCcgCOMClassExists, nil
}

// CLSIDExists is used to get the CLSID value
func clsidExists(_ string) (bool, error) {
	return DummyCLSIDExists, nil
}

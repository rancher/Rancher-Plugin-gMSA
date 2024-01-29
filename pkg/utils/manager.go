package utils

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/go-git/go-billy/v5"
)

var (
	// manager is the default filesystem manager provides access to filesystems for other functions in this module
	manager = &fsManager{}
)

// IsMemFs returns whether the default manager is using a in-memory filesystem
func IsMemFs() bool {
	return manager.isMemFs
}

// fsManager is a manager of filesystems
// This is used as an abstraction for accessing the underlying filesystem
type fsManager struct {
	lock        sync.Mutex
	filesystems map[string]billy.Filesystem

	isMemFs bool
}

// Filesystem returns an appropriate filesystem for a given path
// If the path is not associated with any filesystem, this function will throw an error
func (f *fsManager) Filesystem(path string) (billy.Filesystem, string, error) {
	var subpath string
	for path != "." && path != string(filepath.Separator) {
		filesystem, ok := f.filesystems[path]
		if ok {
			return filesystem, subpath, nil
		}
		subpath = filepath.Join(filepath.Base(path), subpath)
		path = filepath.Dir(path)
	}
	return nil, "", fmt.Errorf("did not recognize filesystem for %s", path)
}

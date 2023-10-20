package utils

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
)

var (
	DryRun bool
)

func SetupTestEnv() {
	useMemFs()
	DryRun = true
}

func SetupEnv() {
	useOsFs()
	// In Linux, we only ever do dry runs
	DryRun = !Windows
}

var (
	manager = &fileManager{}
)

type fileManager struct {
	lock        sync.Mutex
	filesystems map[string]billy.Filesystem

	isMemFs bool
}

func (f *fileManager) Filesystem(path string) (billy.Filesystem, string, error) {
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

func init() {
	SetupEnv()
}

func useOsFs() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.filesystems = map[string]billy.Filesystem{
		DLLDirectory:      osfs.New(DLLDirectory),
		ProviderDirectory: osfs.New(ProviderDirectory),
	}
	manager.isMemFs = false
}

func useMemFs() {
	var err error
	dllDirectory, providerDirectory := memfs.New(), memfs.New()

	dllDirectory, err = dllDirectory.Chroot(DLLDirectory)
	if err != nil {
		panic(fmt.Sprintf("cannot create memfs for %s: %s", DLLDirectory, err))
	}

	providerDirectory, err = providerDirectory.Chroot(ProviderDirectory)
	if err != nil {
		panic(fmt.Sprintf("cannot create memfs for %s: %s", ProviderDirectory, err))
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.filesystems = map[string]billy.Filesystem{
		DLLDirectory:      dllDirectory,
		ProviderDirectory: providerDirectory,
	}
	manager.isMemFs = true
}

func IsMemFs() bool {
	return manager.isMemFs
}

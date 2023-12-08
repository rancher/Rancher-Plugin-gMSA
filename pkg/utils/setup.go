package utils

import (
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
)

const (
	PowershellPathEnvVar  = "POWERSHELL_PATH"
	DefaultPowershellPath = "powershell.exe"
)

var (
	DryRun         bool
	PowershellPath string
)

func init() {
	// By default, use the real filesystem
	// If a test needs to override this, it can call SetupTestEnv
	SetupEnv()
}

// SetupEnv sets up the default fs manager to use the real host (OS) filesystems
func SetupEnv() {
	PowershellPath = getPowershellPath()
	DryRun = false
	useOsFs()
}

// SetupTestEnv sets up the default fs manager to use in-memory filesystems
// Any subsequent call to SetupTestEnv or SetupEnv will wipe out the filesystems created here.
func SetupTestEnv() {
	PowershellPath = getPowershellPath()
	DryRun = true
	useMemFs()
}

func getPowershellPath() string {
	if val := os.Getenv(PowershellPathEnvVar); len(val) != 0 {
		return val
	}
	return DefaultPowershellPath
}

// useOsFs sets up the default manager to use OS filesystems
func useOsFs() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.filesystems = map[string]billy.Filesystem{
		DLLDirectory:      osfs.New(DLLDirectory),
		ProviderDirectory: osfs.New(ProviderDirectory),
	}
	manager.isMemFs = false
}

// useMemFs sets up the default manager to use in-memory filesystems
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

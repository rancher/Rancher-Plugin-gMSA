package status

import (
	"fmt"

	"go.uber.org/multierr"
)

const (
	StatusUpToDate    = "The CCG Plugin is installed and up-to-date."
	StatusUninstalled = "The CCG Plugin is not installed."
)

var (
	ErrNotWindows = fmt.Errorf("not a Windows host")

	ErrDirectoryDoesNotExist         = fmt.Errorf("dll directory does not exist")
	ErrDLLDoesNotExist               = fmt.Errorf("dll file does not exist")
	ErrDLLIsOutOfDate                = fmt.Errorf("dll is out-of-date")
	ErrDLLNotRegisteredInComClasses  = fmt.Errorf("plugin CLSID is not added to CCG COM Classes Key")
	ErrDLLNotRegisteredInHKEYClasses = fmt.Errorf("plugin CLSID is not added to to HKEY_CLASSES_ROOT")
)

type DLLInstallationStatus struct {
	DirectoryExists bool
	Exists          bool
	NeedsUpgrade    bool

	AddedToCOMClassesKey   bool
	AddedToHKEYClassesRoot bool
}

// Install

func (d DLLInstallationStatus) Installed() bool {
	return d.DirectoryExists && d.Exists && !d.NeedsUpgrade && d.AddedToCOMClassesKey && d.AddedToHKEYClassesRoot
}

func (d DLLInstallationStatus) RequiresDirectoryCreation() bool {
	return !d.DirectoryExists
}

func (d DLLInstallationStatus) RequiresUpgrade() bool {
	return d.NeedsUpgrade
}

func (d DLLInstallationStatus) RequiresInstall() bool {
	return !d.Exists || d.NeedsUpgrade
}

func (d DLLInstallationStatus) RequiresRegistration() bool {
	return !d.AddedToCOMClassesKey || !d.AddedToHKEYClassesRoot
}

// Uninstall

func (d DLLInstallationStatus) Uninstalled() bool {
	return !d.DirectoryExists && !d.Exists && !d.AddedToCOMClassesKey && !d.AddedToHKEYClassesRoot
}

func (d DLLInstallationStatus) RequiresDeregistration() bool {
	return d.AddedToCOMClassesKey || d.AddedToHKEYClassesRoot
}

func (d DLLInstallationStatus) RequiresDirectoryDeletion() bool {
	return d.DirectoryExists
}

func (d DLLInstallationStatus) Error() error {
	var err error
	if d.Installed() {
		return nil
	}
	if d.Uninstalled() {
		return nil
	}
	if !d.DirectoryExists {
		err = multierr.Append(err, ErrDirectoryDoesNotExist)
	}
	if !d.Exists {
		err = multierr.Append(err, ErrDLLDoesNotExist)
	}
	if d.NeedsUpgrade {
		err = multierr.Append(err, ErrDLLIsOutOfDate)
	}
	if !d.AddedToCOMClassesKey {
		err = multierr.Append(err, ErrDLLNotRegisteredInComClasses)
	}
	if !d.AddedToHKEYClassesRoot {
		err = multierr.Append(err, ErrDLLNotRegisteredInHKEYClasses)
	}
	return fmt.Errorf("the DLL is not fully installed or uninstalled: %s", err)
}

func (d DLLInstallationStatus) String() string {
	if d.Installed() {
		return StatusUpToDate
	}
	if d.Uninstalled() {
		return StatusUninstalled
	}
	return d.Error().Error()
}

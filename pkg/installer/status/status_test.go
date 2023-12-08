//go:build !windows

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {
	testCases := []struct {
		Name   string
		Status DLLInstallationStatus

		Installed                 bool
		RequiresDirectoryCreation bool
		RequiresUpgrade           bool
		RequiresInstall           bool
		RequiresRegistration      bool
	}{
		{
			Name:   "Default Status",
			Status: DLLInstallationStatus{},

			Installed:                 false,
			RequiresDirectoryCreation: true,
			RequiresUpgrade:           false,
			RequiresInstall:           true,
			RequiresRegistration:      true,
		},
		{
			Name: "Only Directory Exists",
			Status: DLLInstallationStatus{
				DirectoryExists: true,
			},

			Installed:                 false,
			RequiresDirectoryCreation: false,
			RequiresUpgrade:           false,
			RequiresInstall:           true,
			RequiresRegistration:      true,
		},
		{
			Name: "Unregistered DLL Exists",
			Status: DLLInstallationStatus{
				DirectoryExists: true,
				Exists:          true,
				NeedsUpgrade:    false,
			},

			Installed:                 false,
			RequiresDirectoryCreation: false,
			RequiresUpgrade:           false,
			RequiresInstall:           false,
			RequiresRegistration:      true,
		},
		{
			Name: "Out-Of-Date DLL Exists",
			Status: DLLInstallationStatus{
				DirectoryExists: true,
				Exists:          true,
				NeedsUpgrade:    true,
			},

			Installed:                 false,
			RequiresDirectoryCreation: false,
			RequiresUpgrade:           true,
			RequiresInstall:           true,
			RequiresRegistration:      true,
		},
		{
			Name: "Registered DLL Exists",
			Status: DLLInstallationStatus{
				DirectoryExists:        true,
				Exists:                 true,
				AddedToCOMClassesKey:   true,
				AddedToHKEYClassesRoot: true,
			},

			Installed:                 true,
			RequiresDirectoryCreation: false,
			RequiresUpgrade:           false,
			RequiresInstall:           false,
			RequiresRegistration:      false,
		},
		{
			Name: "Partially Registered DLL Exists",
			Status: DLLInstallationStatus{
				DirectoryExists:      true,
				Exists:               true,
				AddedToCOMClassesKey: true,
			},

			Installed:                 false,
			RequiresDirectoryCreation: false,
			RequiresUpgrade:           false,
			RequiresInstall:           false,
			RequiresRegistration:      true,
		},
		{
			Name: "Registered Out-Of-Date DLL Exists",
			Status: DLLInstallationStatus{
				DirectoryExists:        true,
				Exists:                 true,
				NeedsUpgrade:           true,
				AddedToCOMClassesKey:   true,
				AddedToHKEYClassesRoot: true,
			},

			Installed:                 false,
			RequiresDirectoryCreation: false,
			RequiresUpgrade:           true,
			RequiresInstall:           true,
			RequiresRegistration:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name+" Installed", func(t *testing.T) {
			assert.Equal(t, tc.Installed, tc.Status.Installed())
		})
		t.Run(tc.Name+" RequiresDirectoryCreation", func(t *testing.T) {
			assert.Equal(t, tc.RequiresDirectoryCreation, tc.Status.RequiresDirectoryCreation())
		})
		t.Run(tc.Name+" RequiresUpgrade", func(t *testing.T) {
			assert.Equal(t, tc.RequiresUpgrade, tc.Status.RequiresUpgrade())
		})
		t.Run(tc.Name+" RequiresInstall", func(t *testing.T) {
			assert.Equal(t, tc.RequiresInstall, tc.Status.RequiresInstall())
		})
		t.Run(tc.Name+" RequiresRegistration", func(t *testing.T) {
			assert.Equal(t, tc.RequiresRegistration, tc.Status.RequiresRegistration())
		})
		t.Run(tc.Name+" StringAndError", func(t *testing.T) {
			if tc.Installed {
				assert.Equal(t, StatusUpToDate, tc.Status.String())
				assert.Nil(t, tc.Status.Error())
			} else {
				err := tc.Status.Error()
				if err == nil {
					assert.Equal(t, StatusUninstalled, tc.Status.String())
				} else {
					assert.NotEqual(t, StatusUpToDate, tc.Status.String())
					assert.NotEqual(t, StatusUninstalled, tc.Status.String())
				}
			}
		})
	}
}

func TestUninstall(t *testing.T) {
	testCases := []struct {
		Name   string
		Status DLLInstallationStatus

		Uninstalled               bool
		RequiresDeregistration    bool
		RequiresDirectoryDeletion bool
	}{
		{
			Name:   "Default Status",
			Status: DLLInstallationStatus{},

			Uninstalled:               true,
			RequiresDeregistration:    false,
			RequiresDirectoryDeletion: false,
		},
		{
			Name: "Full Uninstall",
			Status: DLLInstallationStatus{
				DirectoryExists:        true,
				Exists:                 true,
				AddedToCOMClassesKey:   true,
				AddedToHKEYClassesRoot: true,
			},

			Uninstalled:               false,
			RequiresDeregistration:    true,
			RequiresDirectoryDeletion: true,
		},
		{
			Name: "DLL Deleted",
			Status: DLLInstallationStatus{
				DirectoryExists:        true,
				Exists:                 false,
				AddedToCOMClassesKey:   true,
				AddedToHKEYClassesRoot: true,
			},

			Uninstalled:               false,
			RequiresDeregistration:    true,
			RequiresDirectoryDeletion: true,
		},
		{
			Name: "DLL Directory Deleted",
			Status: DLLInstallationStatus{
				DirectoryExists:        false,
				Exists:                 false,
				AddedToCOMClassesKey:   true,
				AddedToHKEYClassesRoot: true,
			},

			Uninstalled:               false,
			RequiresDeregistration:    true,
			RequiresDirectoryDeletion: false,
		},
		{
			Name: "Partially Deregistered",
			Status: DLLInstallationStatus{
				DirectoryExists:        true,
				Exists:                 true,
				AddedToCOMClassesKey:   false,
				AddedToHKEYClassesRoot: true,
			},

			Uninstalled:               false,
			RequiresDeregistration:    true,
			RequiresDirectoryDeletion: true,
		},
		{
			Name: "Deregistered But Not Deleted",
			Status: DLLInstallationStatus{
				DirectoryExists:        true,
				Exists:                 true,
				AddedToCOMClassesKey:   false,
				AddedToHKEYClassesRoot: false,
			},

			Uninstalled:               false,
			RequiresDeregistration:    false,
			RequiresDirectoryDeletion: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name+" Uninstalled", func(t *testing.T) {
			assert.Equal(t, tc.Uninstalled, tc.Status.Uninstalled())
		})
		t.Run(tc.Name+" RequiresDeregistration", func(t *testing.T) {
			assert.Equal(t, tc.RequiresDeregistration, tc.Status.RequiresDeregistration())
		})
		t.Run(tc.Name+" RequiresDirectoryDeletion", func(t *testing.T) {
			assert.Equal(t, tc.RequiresDirectoryDeletion, tc.Status.RequiresDirectoryDeletion())
		})
		t.Run(tc.Name+" StringAndError", func(t *testing.T) {
			if tc.Uninstalled {
				assert.Equal(t, StatusUninstalled, tc.Status.String())
				assert.Nil(t, tc.Status.Error())
			} else {
				err := tc.Status.Error()
				if err == nil {
					assert.Equal(t, StatusUpToDate, tc.Status.String())
				} else {
					assert.NotEqual(t, StatusUpToDate, tc.Status.String())
					assert.NotEqual(t, StatusUninstalled, tc.Status.String())
				}
			}
		})
	}
}

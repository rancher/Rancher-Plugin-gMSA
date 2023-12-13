//go:build !windows

package installer

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/installer/embedded"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/installer/status"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {
	testCases := []struct {
		Name              string
		DirectoryExists   bool
		FileExists        bool
		FileMatchesDLL    bool
		CcgCOMClassExists bool
		CLSIDExists       bool

		ExpectError bool
	}{
		{
			Name: "Uninstalled",
			// we are leaving these true since there is no way the installation process can handle this in unix
			CcgCOMClassExists: true,
			CLSIDExists:       true,
		},
		{
			Name:            "Directory Exists",
			DirectoryExists: true,
			// we are leaving these true since there is no way the installation process can handle this in unix
			CcgCOMClassExists: true,
			CLSIDExists:       true,
		},
		{
			Name:            "Upgrade",
			DirectoryExists: true,
			FileExists:      true,
			// we are leaving these true since there is no way the installation process can handle this in unix
			CcgCOMClassExists: true,
			CLSIDExists:       true,
		},
		{
			Name:            "Already Installed",
			DirectoryExists: true,
			FileExists:      true,
			FileMatchesDLL:  true,
			// we are leaving these true since there is no way the installation process can handle this in unix
			CcgCOMClassExists: true,
			CLSIDExists:       true,
		},
		{
			Name: "Failed To Register",
			// we are leaving these false to simulate a failure to register
			CcgCOMClassExists: false,
			CLSIDExists:       false,

			ExpectError: true,
		},
	}

	embedded.DLL = []byte("my-dummy-content-here")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Log("Using mocked filesystem for this test")
			utils.SetupTestEnv()

			var err error
			if tc.DirectoryExists {
				err = utils.CreateDirectory(utils.DLLDirectory)
				assert.Nil(t, err, "expected to be able to create directory in mock filesystem")
			}
			if tc.FileExists {
				if !tc.DirectoryExists {
					// making sure this is a valid test
					t.Fatalf("cannot have test that says file exists if directory does not exits: %s", tc.Name)
				}
				var fileContents []byte
				if tc.FileMatchesDLL {
					fileContents = []byte(embedded.DLL)
				} else {
					fileContents = []byte("needs-upgrade")
				}
				err = utils.SetFile(filepath.Join(utils.DLLDirectory, utils.DLLPath), fileContents)
				assert.Nil(t, err, "expected DLL file to be writeable to mock filesystem")
			}

			ccgCOMClassExists, clsidExists := status.DummyCcgCOMClassExists, status.DummyCLSIDExists
			defer func() {
				status.DummyCcgCOMClassExists, status.DummyCLSIDExists = ccgCOMClassExists, clsidExists
			}()
			status.DummyCcgCOMClassExists, status.DummyCLSIDExists = tc.CcgCOMClassExists, tc.CLSIDExists

			err = Install()
			if tc.ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

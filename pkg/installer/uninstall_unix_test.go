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

func TestUninstall(t *testing.T) {
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
		},
		{
			Name:            "Directory Exists",
			DirectoryExists: true,
		},
		{
			Name:            "Upgrade Needed",
			DirectoryExists: true,
			FileExists:      true,
		},
		{
			Name:            "Installed",
			DirectoryExists: true,
			FileExists:      true,
			FileMatchesDLL:  true,
		},
		{
			Name: "Deregistration Failed",
			// in unix, this should try to do deregistration but fail
			CcgCOMClassExists: true,
			CLSIDExists:       true,

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

			err = Uninstall()
			if tc.ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

//go:build !windows

package status

import (
	"path/filepath"
	"testing"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCheckInstallationStatus(t *testing.T) {
	testCases := []struct {
		Name              string
		DirectoryExists   bool
		FileExists        bool
		FileMatchesDLL    bool
		CcgCOMClassExists bool
		CLSIDExists       bool

		ExpectedStatusMessage string
		ExpectError           bool
	}{
		{
			Name: "Default",

			ExpectedStatusMessage: StatusUninstalled,
		},
		{
			Name:            "Directory Left Behind",
			DirectoryExists: true,

			ExpectError: true,
		},
		{
			Name:            "Upgrade",
			DirectoryExists: true,
			FileExists:      true,

			ExpectError: true,
		},
		{
			Name:            "Reregistration Required",
			DirectoryExists: true,
			FileExists:      true,
			FileMatchesDLL:  true,

			ExpectError: true,
		},
		{
			Name:              "Installed",
			DirectoryExists:   true,
			FileExists:        true,
			FileMatchesDLL:    true,
			CcgCOMClassExists: true,
			CLSIDExists:       true,

			ExpectedStatusMessage: StatusUpToDate,
		},
	}

	packagedDLL := "my-dummy-content-here"

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
					fileContents = []byte(packagedDLL)
				} else {
					fileContents = []byte("needs-upgrade")
				}
				err = utils.SetFile(filepath.Join(utils.DLLDirectory, utils.DLLPath), fileContents)
				assert.Nil(t, err, "expected DLL file to be writeable to mock filesystem")
			}

			ccgCOMClassExists, clsidExists := DummyCcgCOMClassExists, DummyCLSIDExists
			defer func() {
				DummyCcgCOMClassExists, DummyCLSIDExists = ccgCOMClassExists, clsidExists
			}()
			DummyCcgCOMClassExists, DummyCLSIDExists = tc.CcgCOMClassExists, tc.CLSIDExists

			status, err := CheckInstallationStatus([]byte(packagedDLL))
			assert.Nil(t, err, "expected to be able to check installation status successfully")
			if tc.ExpectError {
				assert.NotNil(t, status.Error())
			} else {
				assert.Equal(t, tc.ExpectedStatusMessage, status.String())
			}
		})
	}
}

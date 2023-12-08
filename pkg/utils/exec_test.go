package utils

import (
	"os"
	"path/filepath"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
)

func TestGetPowershellPath(t *testing.T) {
	testCases := []struct {
		Name string

		ExpectedPowershellPath string
		EnvironmentVariables   map[string]string
	}{
		{
			Name:                   "Default",
			ExpectedPowershellPath: DefaultPowershellPath,
		},
		{
			Name:                   "Overridden Path",
			ExpectedPowershellPath: "pwsh.exe",
			EnvironmentVariables: map[string]string{
				PowershellPathEnvVar: "pwsh.exe",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Cleanup(func() {
				for k := range tc.EnvironmentVariables {
					os.Unsetenv(k)
				}
			})
			for k, v := range tc.EnvironmentVariables {
				err := os.Setenv(k, v)
				assert.Nil(t, err, "unable to set environment variable for test")
			}
			assert.Equal(t, tc.ExpectedPowershellPath, getPowershellPath())
		})
	}
}

func TestRunPowershell(t *testing.T) {
	basicCommand := "Write-Output \"hi\""
	dummyPs1 := filepath.Join(ProviderDirectory, "dummy.ps1")
	t.Run("Run On Memfs", func(t *testing.T) {
		SetupTestEnv()

		err := SetFile(dummyPs1, []byte(basicCommand))
		assert.Nil(t, err, "failed to set %s with content '%s'", dummyPs1, basicCommand)

		err = RunPowershell(dummyPs1)
		assert.Nil(t, err, "failed to run powershell file", dummyPs1)
	})
}

func TestRunPowershellCommand(t *testing.T) {
	basicCommand := "Write-Output \"hi\""
	t.Run("Run On Memfs", func(t *testing.T) {
		SetupTestEnv()
		err := RunPowershellCommand(basicCommand)
		assert.Nil(t, err, "failed to run command '%s'", basicCommand)
	})
}

func TestRunPowershellCommandWithOutput(t *testing.T) {
	basicCommand := "Write-Output \"hi\""
	t.Run("Run On Memfs", func(t *testing.T) {
		SetupTestEnv()
		out, err := RunPowershellCommandWithOutput(basicCommand)
		assert.Nil(t, err, "failed to run command '%s'", basicCommand)
		assert.Nil(t, out, "expected nil output on dry run")
	})
}

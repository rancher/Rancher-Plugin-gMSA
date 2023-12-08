//go:build windows

package utils

import (
	"os/exec"
)

func runPowershell(args ...string) (out []byte, err error) {
	if DryRun {
		return dryRunPowershell(args...)
	}
	cmd := exec.Command(PowershellPath, args...)
	return cmd.CombinedOutput()
}

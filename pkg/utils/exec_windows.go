//go:build windows

package utils

import (
	"os/exec"
)

func runPowershell(args ...string) (out []byte, err error) {
	if DryRun {
		return dryRunPowershell(args...)
	}
	printPowershell(args...)
	return exec.Command(PowershellPath, args...).CombinedOutput()
}

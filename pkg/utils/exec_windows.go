//go:build windows

package utils

import (
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func runPowershell(args ...string) (out []byte, err error) {
	if DryRun {
		logrus.Warnf("Skipped executing %s %s", PowershellPath, strings.Join(args, " "))
		return nil, nil
	}
	cmd := exec.Command(PowershellPath, args...)
	return cmd.CombinedOutput()
}

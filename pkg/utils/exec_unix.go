//go:build !windows

package utils

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func runPowershell(args ...string) (out []byte, err error) {
	logrus.Warnf("Skipped executing %s %s", PowershellPath, strings.Join(args, " "))
	return nil, nil
}

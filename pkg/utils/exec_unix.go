//go:build !windows

package utils

func runPowershell(args ...string) (out []byte, err error) {
	return dryRunPowershell(args...)
}

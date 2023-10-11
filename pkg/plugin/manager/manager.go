//go:build !windows

package manager

import (
	"fmt"
)

func Install() error {
	return fmt.Errorf("cannot install plugin: not a Windows host")
}

func Uninstall() error {
	return fmt.Errorf("cannot uninstall plugin: not a Windows host")
}

func Upgrade() error {
	return fmt.Errorf("cannot upgrade plugin: not a Windows host")
}

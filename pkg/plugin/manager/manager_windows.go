package manager

import "github.com/sirupsen/logrus"

func Install() error {
	logrus.Infof("attempted to install the plugin on Windows")
	return nil
}

func Uninstall() error {
	logrus.Infof("attempted to uninstall plugin on Windows")
	return nil
}

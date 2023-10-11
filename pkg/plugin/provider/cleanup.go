package pkg

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func UninstallProvider(namespace string) error {
	logrus.Infof("Uninstalling account provider in namespace %s", namespace)

	// cleanup certs
	if err := RemoveCerts(namespace); err != nil {
		return fmt.Errorf("failed to unimport and delete certificate files: %v", err)
	}

	// remove dynamic dir
	if err := RemoveDynamicDirectory(namespace); err != nil {
		return fmt.Errorf("failed to delete dynamic directory for namespace %s: %v", namespace, err)
	}

	logrus.Infof("successfully uninstalled account provider in namespace %s", namespace)
	return nil
}

package pkg

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

const gmsaDirectory = "/var/lib/rancher/gmsa"

func CreateDir(namespace string) error {
	if runtime.GOOS != "windows" {
		logrus.Warn("Not running on a Windows system, skipping creation of dynamic directory")
		return nil
	}

	// TODO: Adjust Directory Permissions
	if _, err := os.Stat(gmsaDirectory); os.IsNotExist(err) {
		err = os.Mkdir(gmsaDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create base directory: %v", err)
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", gmsaDirectory, namespace)); os.IsNotExist(err) {
		err = os.Mkdir(fmt.Sprintf("%s/%s", gmsaDirectory, namespace), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create dynamic sub directory: %v", err)
		}
	}

	return nil
}

func WritePortFile(namespace, port string) error {
	if runtime.GOOS != "windows" {
		logrus.Warn("Not running on a Windows system, skipping creation of port file")
		return nil
	}

	portFile := fmt.Sprintf("%s/%s/%s", gmsaDirectory, namespace, "port.txt")
	// TODO: adjust certFile permissions
	if _, err := os.Stat(portFile); os.IsNotExist(err) {
		// create the certFile
		err = os.WriteFile(portFile, []byte(port), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create port.txt: %v", err)
		}
	}

	// update certFile with new port
	f, err := os.OpenFile(portFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open port certFile: %v", err)
	}

	_, err = f.WriteString(port)
	if err != nil {
		return fmt.Errorf("failed to update port certFile: %v", err)
	}

	return f.Close()
}

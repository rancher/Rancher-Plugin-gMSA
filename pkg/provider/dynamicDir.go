package provider

import (
	"github.com/sirupsen/logrus"

	"fmt"
	"os"
)

const (
	gmsaDirectory    = "/var/lib/rancher/gmsa"
	rancherDirectory = "/var/lib/rancher"
)

func CreateDynamicDirectory(namespace string) error {
	// TODO: Adjust Directory Permissions

	// this directory may not exist in scenarios where this chart
	// is deployed onto non-rancher clusters.
	if _, err := os.Stat(rancherDirectory); os.IsNotExist(err) {
		err = os.Mkdir(rancherDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create rancher directory: %v", err)
		}
	}

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

func RemoveDynamicDirectory(namespace string) error {
	subDirectory := fmt.Sprintf("%s/%s", gmsaDirectory, namespace)
	logrus.Infof("Removing directory %s", subDirectory)
	if _, err := os.Stat(subDirectory); !os.IsNotExist(err) {
		err = os.RemoveAll(subDirectory)
		if err != nil {
			return fmt.Errorf("error encountered removing subdirectory %s: %v", subDirectory, err)
		}
	}
	return nil
}

func WritePortFile(namespace, port string) error {
	portFile := fmt.Sprintf("%s/%s/%s", gmsaDirectory, namespace, "port.txt")
	// TODO: adjust file permissions
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
		return fmt.Errorf("failed to open port file: %v", err)
	}

	_, err = f.WriteString(port)
	if err != nil {
		return fmt.Errorf("failed to update port file: %v", err)
	}

	return f.Close()
}

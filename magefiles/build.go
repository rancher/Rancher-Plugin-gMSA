//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/sirupsen/logrus"

	"fmt"
)

func Build() error {
	mg.SerialDeps(Validate, Clean, CreateDirs, Dependencies, Test)
	return buildGo()
}

func Quick() error {
	mg.SerialDeps(Clean, CreateDirs, Dependencies)
	return buildGo()
}

func BuildDLL() error {
	mg.SerialDeps(Setup)
	if g.OS != "windows" {
		logrus.Errorf("this target can only be run on Windows")
		return nil
	}

	logrus.Infof("Building DLL artifact")

	buildContainerImage := fmt.Sprintf("%s/gmsa-ccg-plugin-builder:latest", repo)
	buildContainerName := "dllbuild"
	containerFilePath := "/app/bin/Debug/RanchergMSACredentialProvider.dll"
	containerDockerFilePath := "package/ccg-plugin-installer/Dockerfile.builder.windows"
	projectFilePath := "pkg/plugin/manager/RanchergMSACredentialProvider.dll"

	// todo; we should be using delayed signing here

	logrus.Infof("Build container name will be %s", buildContainerImage)

	err := docker.Build(nil, containerDockerFilePath, buildContainerImage, ".", true)
	if err != nil {
		logrus.Errorf("could not build docker image: %v", err)
		return err
	}

	logrus.Infof("Creating dll build container")
	_, err = docker.Create(buildContainerName, buildContainerImage)
	if err != nil {
		return err
	}

	logrus.Infof("Copying contents of dll build container into project")
	err = docker.Copy(buildContainerName, containerFilePath, projectFilePath)
	if err != nil {
		logrus.Errorf("encountered error copying dll from container: %v", err)
	}

	logrus.Infof("Removing temporary build container")
	if err = docker.Remove(buildContainerName); err != nil {
		return err
	}

	logrus.Infof("Done! %s has been updated", projectFilePath)
	return nil
}

func buildGo() error {
	logrus.Infof("Beginning to build binaries")
	for _, application := range applications {
		logrus.Infof("Building %s", application)
		if err := g.Build(flags, commandLocation(application), binaryOutputLocation(application, binOutput)); err != nil {
			return fmt.Errorf("failed to build %s: %v", application, err)
		}
	}
	return nil
}

func binaryOutputLocation(application, folder string) string {

	// applicationName-platform-arch
	format := "%s-%s-%s"

	// .exe is not strictly required, but
	// it is good practice to add it so that we can
	// more easily distinguish between binaries
	if g.OS == "windows" {
		format = "%s-%s-%s.exe"
	}

	return fmt.Sprintf("%s/%s", folder, fmt.Sprintf(format, application, g.OS, g.Arch))
}

func commandLocation(application string) string {
	return fmt.Sprintf("cmd/%s/main.go", application)
}

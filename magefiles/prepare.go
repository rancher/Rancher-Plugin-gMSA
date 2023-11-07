//go:build mage

package main

import (
	"github.com/aiyengar2/Rancher-Plugin-gMSA/magetools"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/sirupsen/logrus"

	"fmt"
	"os"
	"strings"
)

func Test() error {
	mg.SerialDeps(Setup)
	logrus.Infof("Running go tests")
	return g.Test(flags, "./...")
}

func Validate() error {
	logrus.Infof("Running validations")
	envs := map[string]string{"GOOS": "windows", "ARCH": "amd64", "CGO_ENABLED": "0", "MAGEFILE_VERBOSE": "1"}

	logrus.Infof("[Validate] Running: golangci-lint \n")
	err := sh.RunWithV(envs, "golangci-lint", "run")
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			logrus.Warnf("golangci-lint not found, skipping linter checks")
		} else {
			return err
		}
	}

	logrus.Infof("[Validate] Running: go fmt \n")
	if err := sh.RunWithV(envs, "go", "fmt", "./..."); err != nil {
		return err
	}

	logrus.Infof("validate has completed successfully \n")
	return nil
}

func Dependencies() error {
	mg.SerialDeps(Setup)
	logrus.Infof("Downloading dependencies")
	return g.Mod("download")
}

func CreateDirs() error {
	logrus.Infof("Creating dist and bin directories")
	err := os.Mkdir("bin", os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	err = os.Mkdir("dist", os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}
	return nil
}

func Version() error {
	repo = magetools.GetRepo()

	c, err := magetools.GetCommit()
	if err != nil {
		return err
	}
	commit = c

	dt := os.Getenv("DRONE_TAG")
	isClean, err := magetools.IsGitClean()
	if err != nil {
		return err
	}
	if dt != "" && isClean {
		version = dt
		return nil
	}

	tag, err := magetools.GetLatestTag()
	if err != nil {
		return err
	}
	if tag != "" && isClean {
		version = tag
		return nil
	}

	version = commit
	if !isClean {
		version = commit + "-dirty"
		dirty = true
		logrus.Printf("[Version] dirty version encountered: %s \n", version)
	}

	// check if this is a release version and fail if the version contains `dirty`
	if strings.Contains(version, "dirty") && os.Getenv("DRONE_TAG") != "" || tag != "" {
		return fmt.Errorf("[Version] releases require a non-dirty tag: %s", version)
	}
	logrus.Printf("[Version] version: %s \n", version)

	return nil
}

func Clean() error {
	logrus.Info("Removing bin and dist directories")
	if err := sh.Rm(artifactOutput); err != nil {
		return err
	}
	return sh.Rm(binOutput)
}

//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/sirupsen/logrus"

	"fmt"
)

// Package builds the docker files for both applications. When run on windows, each
// application will produce two images, one for ltsc2019 and one for ltsc2022.
// Note: Windows 2019 cannot build 2022 images, 2022 and build both, because of this
// CI runners must use Windows 2022.
func Package() error {
	mg.SerialDeps(Build, stageBinaries)

	logrus.Infof("Beginning to package applications")

	for _, application := range applications {
		if g.OS == "windows" {
			for _, nanoServerVersion := range nanoServerVersions {
				// go-ism, can't take pointers of range statements
				nsv := nanoServerVersion

				args := map[string]*string{
					"OS":                 &g.OS,
					"ARCH":               &g.Arch,
					"NANOSERVER_VERSION": &nsv,
				}
				if err := docker.Build(args,
					dockerfileLocation(application),
					buildDockerTag(application, fmt.Sprintf("%s-%s", version, nanoServerVersion)),
					".", false); err != nil {
					return err
				}
			}
		} else {
			args := map[string]*string{
				"OS":   &g.OS,
				"ARCH": &g.Arch,
			}
			err := docker.Build(args, dockerfileLocation(application), buildDockerTag(application, version), ".", false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func stageBinaries() error {
	for _, application := range applications {
		err := sh.Copy(binaryOutputLocation(application, artifactOutput), binaryOutputLocation(application, binOutput))
		if err != nil {
			return fmt.Errorf("could stage %s: %v", application, err)
		}
	}
	return nil
}

func dockerfileLocation(application string) string {
	loc := fmt.Sprintf("package/%s/Dockerfile", application)
	if g.OS == "windows" {
		return loc + ".windows"
	}
	return loc
}

func buildDockerTag(application string, versionTag string) string {
	return fmt.Sprintf("%s/%s:%s", repo, application, versionTag)
}

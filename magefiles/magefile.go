//go:build mage

package main

import (
	"github.com/aiyengar2/Rancher-Plugin-gMSA/magetools"
	"github.com/magefile/mage/mg"

	"fmt"
	"path/filepath"
)

/*
	Global variables and select targets should be placed in this file. Targets specific
	to a particular build stage should be placed within the respective file.

	Additional targets can be easily added by adding new, exported, functions.
*/

var Default = Build
var g *magetools.Go
var docker *magetools.Docker

var dirty bool
var version string
var commit string
var repo string

var artifactOutput = filepath.Join("./dist")
var binOutput = filepath.Join("./bin")

var applications = []string{
	"gmsa-account-provider",
	"ccg-plugin-installer",
}

var nanoServerVersions = []string{
	"ltsc2022",
	"ltsc2019",
}

func CI() error {
	mg.SerialDeps(Version, Validate)

	if dirty {
		return fmt.Errorf("Git is dirty")
	}

	return Package()
}

func FullBuild() error {
	mg.SerialDeps(BuildDLL, Package)
	return nil
}

func flags(version string, commit string) string {
	return fmt.Sprintf(`-s -w -X github.com/rancher/Rancher-Plugin-gMSA/pkg/version.Version=%s -X github.com/rancher/Rancher-Plugin-gMSA/pkg/version.GitCommit=%s -extldflags "-static"`, version, commit)
}

func Setup() {
	mg.SerialDeps(Version)
	g = magetools.NewGo(version, commit, "0", "1")
	docker = magetools.NewDocker()
}

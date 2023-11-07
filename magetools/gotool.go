package magetools

import (
	"github.com/magefile/mage/sh"

	"os/exec"
)

type Go struct {
	Arch       string
	OS         string
	Version    string
	Commit     string
	CGoEnabled string
	Verbose    string
}

func NewGo(version, commit, cgoEnabled, verbose string) *Go {
	os, _ := sh.Output("go", "env", "GOOS")
	if os == "Windows_NT" {
		os = "windows"
	}

	goarch, _ := sh.Output("go", "env", "GOHOSTARCH")

	return &Go{
		Arch:       goarch,
		OS:         os,
		Version:    version,
		Commit:     commit,
		CGoEnabled: cgoEnabled,
		Verbose:    verbose,
	}
}

func (g *Go) Build(flags func(string, string) string, target, output string) error {
	envs := map[string]string{"GOOS": g.OS, "ARCH": g.Arch, "CGO_ENABLED": g.CGoEnabled, "MAGEFILE_VERBOSE": g.Verbose}
	return sh.RunWithV(envs, "go", "build", "-o", output, "--ldflags="+flags(g.Version, g.Commit), target)
}

func (g *Go) Test(flags func(string, string) string, target string) error {
	envs := map[string]string{"GOOS": g.OS, "ARCH": g.Arch, "CGO_ENABLED": g.CGoEnabled, "MAGEFILE_VERBOSE": g.Verbose}
	return sh.RunWithV(envs, "go", "test", "-v", "-cover", "--ldflags="+flags(g.Version, g.Commit), target)
}

func (g *Go) Mod(cmd string) error {
	envs := map[string]string{"GOOS": g.OS, "ARCH": g.Arch}
	return sh.RunWithV(envs, "go", "mod", cmd)
}

func (g *Go) GOOS() (string, error) {
	out, err := exec.Command("go", "env", "GOOS").Output()
	return string(out), err
}

func (g *Go) GOARCH() (string, error) {
	out, err := exec.Command("go", "env", "GOHOSTARCH").Output()
	return string(out), err
}

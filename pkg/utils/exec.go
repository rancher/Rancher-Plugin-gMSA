package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/wrangler/pkg/randomtoken"
	"github.com/sirupsen/logrus"
)

const (
	PowershellPathEnvVar = "POWERSHELL_PATH"

	DefaultPowershellPath = "powershell.exe"
)

var (
	PowershellPath = getPowershellPath()

	CertStoreLocation = filepath.Join("Cert:", "LocalMachine", "Root")

	pfxDryRunContent = []byte("<insert-pfx-file-contents-here>")
)

func getPowershellPath() string {
	if val := os.Getenv(PowershellPathEnvVar); len(val) != 0 {
		return val
	}
	return DefaultPowershellPath
}

func RunPowershell(path string) error {
	args := []string{"-File", path}
	out, err := runPowershell(args...)
	if err != nil {
		if out != nil {
			logrus.Info(string(out))
		}
		return err
	}
	if out != nil {
		logrus.Debug(out)
	}
	return nil
}

func RunPowershellCommand(cmd string) error {
	_, err := RunPowershellCommandWithOutput(cmd)
	return err
}

func RunPowershellCommandWithOutput(cmd string) ([]byte, error) {
	args := append([]string{"-Command"}, strings.Split(cmd, " ")...)
	out, err := runPowershell(args...)
	if err != nil {
		if out != nil {
			logrus.Info(string(out))
		}
		return nil, err
	}
	if out != nil {
		logrus.Debug(out)
	}
	return out, nil
}

func ImportCertificate(certPath string) error {
	cmd := fmt.Sprintf("Import-Certificate -FilePath %s -CertStoreLocation %s -Verbose", certPath, CertStoreLocation)
	logrus.Infof("Importing certificate from %s into %s", certPath, CertStoreLocation)
	return RunPowershellCommand(cmd)
}

func UnimportCertificate(certPath string) error {
	cmd := fmt.Sprintf(`certutil %s | Select-String -Pattern "Cert Hash\(sha1\):*"`, certPath)
	out, err := RunPowershellCommandWithOutput(cmd)
	if err != nil {
		return err
	}
	var thumbprint string
	if out == nil {
		thumbprint = "<THUMBPRINT>"
	} else {
		// Powershell was not actually executed
		thumb := strings.Split(strings.ReplaceAll(strings.ReplaceAll(string(out), "\n", ""), "\r", ""), " ")
		if len(thumb) != 3 {
			return fmt.Errorf("encountered error determining thumbprint of %s, certutil did not return properly formatted hash: \nCert Util Output: %s \nExtracted SHA1 Field: %s", certPath, out, thumb)
		}
		thumbprint = thumb[2]
	}
	thumbprintKey := filepath.Join(CertStoreLocation, thumbprint)
	cmd = fmt.Sprintf("Test-Path %s", thumbprintKey)
	out, err = RunPowershellCommandWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("unable to find certificate %s in cert store at %s: %s", certPath, thumbprintKey, err)
	}
	if string(out) == "False" {
		// certificate already unimported
		return nil
	}
	cmd = fmt.Sprintf("Get-ChildItem %s | Remove-Item", thumbprintKey)
	err = RunPowershellCommand(cmd)
	if err != nil {
		return fmt.Errorf("unable to remove certificate %s listed in cert store under %s: %s", certPath, thumbprintKey, err)
	}
	return nil
}

func CreateAndImportPfx(certFile string, keyFile string, pfxFile string) error {
	password, err := randomtoken.Generate()
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("certutil -p %s -MergePFX %s %s", password, certFile, pfxFile)
	if err := RunPowershellCommand(cmd); err != nil {
		return err
	}

	logrus.Infof("Using %s as the password for the client pfx file at %s", password, pfxFile)
	cmd = strings.Join([]string{
		fmt.Sprintf("$ssPassword = ConvertTo-SecureString \"%s\" -AsPlainText -Force", password),
		fmt.Sprintf("Import-P fxCertificate -Filepath %s -CertStoreLocation %s -Password $ssPassword", pfxFile, CertStoreLocation),
	}, ";\n")
	err = RunPowershellCommand(cmd)
	if err != nil {
		return err
	}
	if !DryRun {
		return nil
	}
	// This indicates that powershell commands are not running anyways, so just create a dummy file
	return SetFile(pfxFile, pfxDryRunContent)
}

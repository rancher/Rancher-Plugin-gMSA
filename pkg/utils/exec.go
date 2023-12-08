package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	CertStoreLocation = filepath.Join("Cert:", "LocalMachine", "Root")
)

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
		logrus.Debugf("%s", out)
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
		logrus.Debugf("%s", out)
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
		return fmt.Errorf("could not find certificate: %s", err)
	}
	var thumbprint string
	if out == nil {
		// Powershell was not actually executed
		thumbprint = "<THUMBPRINT>"
	} else {
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

func CreateAndImportPfx(certFile string, pfxFile string) error {
	err := DeleteFile(pfxFile)
	if err != nil {
		return err
	}
	password := "password"
	cmd := fmt.Sprintf(`certutil -p "%s" -MergePFX %s %s`, password, certFile, pfxFile)
	if err := RunPowershellCommand(cmd); err != nil {
		return err
	}
	cmd = strings.Join([]string{
		fmt.Sprintf("$ssPassword = ConvertTo-SecureString \"%s\" -AsPlainText -Force", password),
		fmt.Sprintf("Import-PfxCertificate -Filepath %s -CertStoreLocation %s -Password $ssPassword", pfxFile, CertStoreLocation),
	}, ";\n")
	err = RunPowershellCommand(cmd)
	if err != nil {
		return err
	}
	if !DryRun {
		return nil
	}
	// This indicates that powershell commands are not running anyways, so just create a dummy file
	return SetFile(pfxFile, []byte(""))
}

func dryRunPowershell(args ...string) (out []byte, err error) {
	logrus.Warnf("Skipped executing %s %s", PowershellPath, strings.Join(args, " "))
	return nil, nil
}

func printPowershell(args ...string) (out []byte, err error) {
	logrus.Debugf("%s %s", PowershellPath, strings.Join(args, " "))
	return nil, nil
}

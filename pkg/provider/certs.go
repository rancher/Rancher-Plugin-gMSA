package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	containerSslDir = "%s/%s/container"

	containerServerCa  = "%s/%s/container/ssl/server/ca.crt"
	containerServerCrt = "%s/%s/container/ssl/server/tls.crt"
	containerServerKey = "%s/%s/container/ssl/server/tls.key"

	containerClientDir = containerSslDir + "/ssl/client"
	containerClientCa  = containerClientDir + "/ca.crt"
	containerClientCrt = containerClientDir + "/tls.crt"
	containerClientKey = containerClientDir + "/tls.key"

	containerRootCaDir = "%s/%s/container/ssl/ca"
	containerRootCa    = containerRootCaDir + "/ca.crt"
	containerRootCrt   = containerRootCaDir + "/tls.crt"

	hostSslDir = "%s/%s/ssl"

	hostRootCaDir = hostSslDir + "/ca"
	hostRootCa    = hostRootCaDir + "/ca.crt"
	hostRootCrt   = hostRootCaDir + "/tls.crt"

	hostClientDir = hostSslDir + "/client"
	hostClientCa  = hostClientDir + "/ca.crt"
	hostClientCrt = hostClientDir + "/tls.crt"
	hostClientPfx = hostClientDir + "/tls.pfx"
	hostClientKey = hostClientDir + "/tls.key"

	hostServerDir = hostSslDir + "/server"
	hostServerCa  = hostServerDir + "/ca.crt"
	hostServerCrt = hostServerDir + "/tls.crt"
)

func createDirectory(directory string) error {
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create directory %s: %v", directory, err)
	}
	return nil
}

type certFile struct {
	// isKey indicates the file should be written to the host
	// but not imported as a certificate
	isKey bool
	// pfxConvert indicates that the certificate should be
	// passed to certutil. If a certificate has pfxConvert = true
	// then there needs to be an associated key file in the same directory
	// with the same name (tls.crt & tls.key)
	pfxConvert bool
	// where in the container fs the file is
	containerFile string
	// the file on host we should write to
	hostFile string
	// where in the host fs the file should be written to
	hostDir string
}

func getCertFiles(activeDirectoryName string) []certFile {
	//todo; trim this down to only what is needed
	return []certFile{
		{
			containerFile: fmt.Sprintf(containerRootCa, gmsaDirectory, activeDirectoryName),
			hostFile:      fmt.Sprintf(hostRootCa, gmsaDirectory, activeDirectoryName),
			hostDir:       fmt.Sprintf(hostRootCaDir, gmsaDirectory, activeDirectoryName),
		},
		{
			containerFile: fmt.Sprintf(containerRootCrt, gmsaDirectory, activeDirectoryName),
			hostFile:      fmt.Sprintf(hostRootCrt, gmsaDirectory, activeDirectoryName),
			hostDir:       fmt.Sprintf(hostRootCaDir, gmsaDirectory, activeDirectoryName),
		},
		{
			isKey:         true,
			containerFile: fmt.Sprintf(containerClientKey, gmsaDirectory, activeDirectoryName),
			hostFile:      fmt.Sprintf(hostClientKey, gmsaDirectory, activeDirectoryName),
			hostDir:       fmt.Sprintf(hostClientDir, gmsaDirectory, activeDirectoryName),
		},
		{
			pfxConvert:    true,
			containerFile: fmt.Sprintf(containerClientCrt, gmsaDirectory, activeDirectoryName),
			hostFile:      fmt.Sprintf(hostClientCrt, gmsaDirectory, activeDirectoryName),
			hostDir:       fmt.Sprintf(hostClientDir, gmsaDirectory, activeDirectoryName),
		},
		{
			containerFile: fmt.Sprintf(containerServerCrt, gmsaDirectory, activeDirectoryName),
			hostFile:      fmt.Sprintf(hostServerCrt, gmsaDirectory, activeDirectoryName),
			hostDir:       fmt.Sprintf(hostServerDir, gmsaDirectory, activeDirectoryName),
		},
		{
			containerFile: fmt.Sprintf(containerServerCa, gmsaDirectory, activeDirectoryName),
			hostFile:      fmt.Sprintf(hostServerCa, gmsaDirectory, activeDirectoryName),
			hostDir:       fmt.Sprintf(hostServerDir, gmsaDirectory, activeDirectoryName),
		},
	}
}

func WriteCerts(namespace string) error {
	if runtime.GOOS != "windows" {
		logrus.Warn("Not running on a Windows system, will not write certificates to system")
		return nil
	}

	logrus.Debugf("creating base certificate directory")
	if err := createDirectory(fmt.Sprintf(hostSslDir, gmsaDirectory, namespace)); err != nil {
		return fmt.Errorf("error encountered creating base directory'%s': %v", fmt.Sprintf(hostSslDir, gmsaDirectory, namespace), err)
	}

	files := getCertFiles(namespace)
	for _, file := range files {
		err := createDirectory(file.hostDir)
		if err != nil {
			return fmt.Errorf("failed to setup base host certificate directories: %v", err)
		}
	}

	logrus.Debugf("writing certificates for %s", namespace)
	for _, file := range files {
		bytes, err := os.ReadFile(file.containerFile)
		if err != nil {
			return fmt.Errorf("failed to read %s from container: %v", file.hostFile, err)
		}

		err = os.WriteFile(file.hostFile, bytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to write %s to host: %v", file.hostFile, err)
		}

		switch {
		case file.isKey:
			continue
		case file.pfxConvert:
			err = generateAndImportPfx(file, namespace)
			if err != nil {
				return fmt.Errorf("failed to create and import pfx file: %v", err)
			}
		default:
			err = importCertificate(file)
			if err != nil {
				return fmt.Errorf("failed to import certificate: %v", err)
			}
		}
	}

	return nil
}

func generateAndImportPfx(file certFile, namespace string) error {
	err := pfxClean(namespace)
	if err != nil {
		return fmt.Errorf("error encountered cleaning outdated pfx file: %v", err)
	}

	err = pfxConvert(file)
	if err != nil {
		return fmt.Errorf("error encountered generating pfx file: %v", err)
	}

	err = pfxImport(file)
	if err != nil {
		return fmt.Errorf("error encountered importing pfx file: %v", err)
	}

	err = removeKeyFile(file)
	if err != nil {
		return fmt.Errorf("error removing keyfile for file %s", file.hostFile)
	}

	// todo; should we also get rid of the cert?
	return nil
}

func importCertificate(file certFile) error {
	cmd := exec.Command("powershell", "-Command", "Import-Certificate", "-FilePath", file.hostFile, "-CertStoreLocation", "Cert:\\LocalMachine\\Root", "-Verbose")
	logrus.Debugf("Importing certificate: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add certificate to LocalMachine Root store (%s): %v", cmd.String(), err)
	}
	logrus.Debug(string(out))
	return nil
}

func RemoveCerts(namespace string) error {
	if runtime.GOOS != "windows" {
		logrus.Warn("Not running on a Windows system, no certificates to remove")
		return nil
	}
	for _, file := range getCertFiles(namespace) {
		if file.isKey {
			// keys aren't directly imported, so they won't appear in the cert store
			continue
		}
		if err := UnImportCertificate(file, namespace); err != nil {
			return fmt.Errorf("error encountered removing certificate %s from store: %v", file.hostFile, err)
		}
	}
	return nil
}

func UnImportCertificate(file certFile, namespace string) error {
	dynamicDir := fmt.Sprintf("%s/%s", gmsaDirectory, namespace)

	logrus.Infof("Removing certificate %s", file.hostFile)

	// get cert thumbprint using certutil. Thumbprints are equal to the sha1 hash of the certificate.
	// Cert util will return a sha1 field in the following format
	// Cert Hash(sha1): <someHash>
	certUtilArgs := []string{"-Command",
		"certutil",
		filepath.Join("C:%s", file.hostFile), "|", "Select-String", "-Pattern", `"Cert Hash\(sha1\)"`,
	}

	cmd := exec.Command("powershell", certUtilArgs...)
	logrus.Debugf("Crafted Command: %s", cmd.String())

	o, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Infof("Encountered error calling certutil, output: %s", string(o))
		return fmt.Errorf("failed to obtain sha1 thumbPrint of cert in %s: %v", dynamicDir, err)
	}

	tmp := string(o)
	tmp = strings.ReplaceAll(tmp, "\n", "")
	tmp = strings.ReplaceAll(tmp, "\r", "")
	thumb := strings.Split(tmp, " ")
	if len(thumb) != 3 {
		return fmt.Errorf("encountered error determining thumbprint of %s, certutil did not return properly formatted hash: \nCert Util Output: %s \nExtracted SHA1 Field: %s", file.hostFile, string(o), thumb)
	}

	thumbPrint := thumb[2]
	thumbPrintLocation := filepath.Join("Cert:", "LocalMachine", "Root", thumbPrint)

	// check if cert exists in store
	pwshArgs := []string{"-Command",
		"Test-Path", thumbPrintLocation}

	o, err = exec.Command("powershell", pwshArgs...).CombinedOutput()
	if err != nil {
		logrus.Infof("Error determining certificate path, output: %s", string(o))
		return fmt.Errorf("error encountered testing certificate path %s: %v", thumbPrintLocation, err)
	}
	logrus.Debugf("Crafted Command: %s", cmd.String())

	if string(o) == "False" {
		// cert does not exist in store, nothing to remove
		return nil
	}

	pwshArgs = []string{"-Command",
		"Get-ChildItem", thumbPrintLocation, "|", "Remove-Item"}

	cmd = exec.Command("powershell", pwshArgs...)
	logrus.Debugf("Crafted Command: %s", cmd.String())

	o, err = cmd.CombinedOutput()
	if err != nil {
		logrus.Infof("Encountered error removing certificate, Remove-Item output: %s", string(o))
		return fmt.Errorf("failed to remove certificate %s with cert location %s: %v", file.hostFile, thumbPrintLocation, err)
	}

	logrus.Infof("Successfully removed %s", file.hostFile)
	return nil
}

func pfxClean(namespace string) error {
	_, err := os.Stat(fmt.Sprintf(fmt.Sprintf(hostClientPfx, gmsaDirectory, namespace)))
	if err == nil {
		err = os.Remove(fmt.Sprintf(fmt.Sprintf(hostClientPfx, gmsaDirectory, namespace)))
		if err != nil {
			return fmt.Errorf("failed to remove outdated pfx file: %v", err)
		}
	}
	return nil
}

func pfxConvert(file certFile) error {
	// todo; gen random password and ensure things still work
	cmd := exec.Command("powershell", "-Command", "cd", file.hostDir, ";", "certutil", "-p", "\"password\"", "-MergePFX", "tls.crt", "tls.pfx")
	logrus.Debugf("generating PFX certFile: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certutil failed to generate pfx file (%s): %v", cmd.String(), err)
	}
	logrus.Debugf("PFX generation logs: %s\n", string(out))
	logrus.Debugf("PFX generation error: %v\n", err)
	return nil
}

func pfxImport(file certFile) error {
	// import the pfx cert onto the system
	cmd := exec.Command("powershell", "-Command", "cd", file.hostDir, ";", "$secureString = ConvertTo-SecureString password -AsPlainText -Force", ";", "Import-PfxCertificate", "-Filepath", "tls.pfx", "-CertStoreLocation", "Cert:\\LocalMachine\\Root", "-Password", "$secureString")
	logrus.Debugf("Importing PFX certFile: %s\n", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Import-PfxCertificate failed to import generated pfx file (%s): %v", cmd.String(), err)
	}
	logrus.Debugf("PFX Import logs: %s\n", string(out))
	logrus.Debugf("PFX Image Error: %s\n", err)
	return nil
}

// removeKeyFile removes the keyfile associated with a certificate.
func removeKeyFile(file certFile) error {
	err := os.Remove(strings.ReplaceAll(file.hostFile, ".crt", ".key"))
	if err != nil {
		return fmt.Errorf("failed to remove key file associated with %s: %v", file.hostFile, err)
	}
	return nil
}

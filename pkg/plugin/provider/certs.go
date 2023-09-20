package pkg

import (
	"fmt"
	"os"
	"os/exec"
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
	err := os.Mkdir(directory, os.ModePerm)
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

	err := createDirectory(fmt.Sprintf(hostSslDir, gmsaDirectory, namespace))
	if err != nil {
		return fmt.Errorf("failed to setup base host certificate directory: %v", err)
	}

	files := getCertFiles(namespace)
	for _, file := range files {
		err := createDirectory(file.hostDir)
		if err != nil {
			return fmt.Errorf("failed to setup base host certificate directories: %v", err)
		}
	}

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
			_, err := os.Stat(fmt.Sprintf(fmt.Sprintf(hostClientPfx, gmsaDirectory, namespace)))
			if err == nil {
				err = os.Remove(fmt.Sprintf(fmt.Sprintf(hostClientPfx, gmsaDirectory, namespace)))
				if err != nil {
					return fmt.Errorf("failed to remove outdated pfx file: %v", err)
				}
			}

			err = generateAndImportPfx(file)
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

func generateAndImportPfx(file certFile) error {
	err := pfxConvert(file)
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

func pfxConvert(file certFile) error {
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

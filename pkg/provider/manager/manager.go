package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

var (
	Certificates = map[string][]string{
		"ca":     {"ca.crt", "tls.crt"},
		"server": {"ca.crt", "tls.crt", "tls.key"},
		"client": {"ca.crt", "tls.crt", "tls.key"},
	}

	CopyCertificates = map[string][]string{
		"ca":     {"ca.crt", "tls.crt"},
		"server": {"ca.crt", "tls.crt"},
		"client": {"ca.crt", "tls.crt", "tls.key"},
	}
)

func New(namespace string) CertificateManager {
	return &manager{
		namespace: namespace,
	}
}

type CertificateManager interface {
	Start(ctx context.Context) error
	Clean(ctx context.Context) error
	ServerCertificates() *TLSCertificates
}

type TLSCertificates struct {
	CertFile string
	KeyFile  string
}

type manager struct {
	namespace string

	lock    sync.RWMutex
	started bool
}

func (m *manager) Start(_ context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.started = true
	// watch container certificate directories
	var containerCertFiles []string
	var err error
	for certDir, certFiles := range Certificates {
		logrus.Debugf("looking for certs %s in %s", certFiles, m.containerSSL(certDir))
		for _, certFile := range certFiles {
			containerCertFile := m.containerSSL(certDir, certFile)
			containerCertFiles = append(containerCertFiles, containerCertFile)
			exists, fileErr := utils.FileExists(containerCertFile)
			if fileErr != nil {
				return fileErr
			}
			if !exists {
				err = multierr.Append(err, fmt.Errorf("could not find %s", m.containerSSL(certDir, certFile)))
			} else {
				logrus.Debugf("certificate %s exists", m.containerSSL(certDir, certFile))
			}
		}

	}
	if err != nil {
		logrus.Warnf("Could not find certificates. Consider running with --disable-mtls")
		return fmt.Errorf("could not find certificates: %s", err)
	}

	logrus.Info("Copying container certificates to host directory")
	if err := m.copyCertificates(); err != nil {
		return err
	}

	logrus.Info("Creating and importing PFX file for client certificates")
	err = utils.CreateAndImportPfx(
		m.hostSSL("client", "tls.crt"),
		m.hostSSL("client", "tls.pfx"),
	)
	if err != nil {
		return err
	}

	// copy over container certificates
	logrus.Info("Importing certificates")
	if err := m.importCertificates(); err != nil {
		return err
	}

	return nil
}

func (m *manager) Clean(_ context.Context) error {
	logrus.Info("Unimporting certificates for client certificates")
	if err := m.unimportCertificates(); err != nil {
		return err
	}
	logrus.Infof("Cleaning up %s", m.baseDir())
	if err := utils.DeleteDirectory(m.baseDir()); err != nil {
		return err
	}
	return nil
}

func (m *manager) ServerCertificates() *TLSCertificates {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if !m.started {
		// Either Start was initiated without valid certs in place or this was never started
		return nil
	}
	return &TLSCertificates{
		CertFile: m.containerSSL("server", "tls.crt"),
		KeyFile:  m.containerSSL("server", "tls.key"),
	}
}

func (m *manager) copyCertificates() error {
	for certDir, certFiles := range CopyCertificates {
		if err := utils.DeleteDirectory(m.hostSSL(certDir)); err != nil {
			return err
		}
		if err := utils.CreateDirectory(m.hostSSL(certDir)); err != nil {
			return err
		}
		for _, certFile := range certFiles {
			containerCertFile := m.containerSSL(certDir, certFile)
			hostCertFile := m.hostSSL(certDir, certFile)
			cert, err := utils.GetFile(containerCertFile)
			if err != nil {
				return err
			}
			err = utils.SetFile(hostCertFile, cert)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *manager) importCertificates() error {
	for certDir, certFiles := range CopyCertificates {
		for _, certFile := range certFiles {
			if filepath.Ext(certFile) != ".crt" {
				continue
			}
			hostCertFile := m.hostSSL(certDir, certFile)
			err := utils.ImportCertificate(hostCertFile)
			if err != nil {
				return fmt.Errorf("failed to import certificate: %v", err)
			}
		}
	}
	return nil
}

func (m *manager) unimportCertificates() error {
	for certDir, certFiles := range CopyCertificates {
		for _, certFile := range certFiles {
			if filepath.Ext(certFile) != ".crt" {
				continue
			}
			hostCertFile := m.hostSSL(certDir, certFile)
			exists, err := utils.FileExists(hostCertFile)
			if err != nil {
				return fmt.Errorf("failed to find certificate %s: %v", hostCertFile, err)
			}
			if !exists {
				// there is no certificate to unimport
				continue
			}
			err = utils.UnimportCertificate(hostCertFile)
			if err != nil {
				return fmt.Errorf("failed to unimport certificate %s: %v", hostCertFile, err)
			}
		}
	}
	return nil
}

func (m *manager) baseDir() string {
	return filepath.Join(utils.ProviderDirectory, m.namespace)
}

func (m *manager) containerSSL(path ...string) string {
	path = append([]string{m.baseDir(), "container", "ssl"}, path...)
	return filepath.Join(path...)
}

func (m *manager) hostSSL(path ...string) string {
	path = append([]string{m.baseDir(), "ssl"}, path...)
	return filepath.Join(path...)
}

package manager

import (
	"path/filepath"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/testutils"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
)

func CreateDummyCerts(namespace string) {
	privateKey, publicCert, err := testutils.SelfSignedCertificate()
	if err != nil {
		logrus.Fatal(err)
	}
	m := manager{
		namespace: namespace,
	}
	for certDir, certFiles := range Certificates {
		containerCertDir := m.containerSSL(certDir)
		if err := utils.CreateDirectory(containerCertDir); err != nil {
			logrus.Fatal(err)
			return
		}
		for _, certFile := range certFiles {
			var content []byte
			switch filepath.Ext(certFile) {
			case ".crt":
				content = publicCert
			case ".key":
				content = privateKey
			default:
				logrus.Fatalf("cannot generate dummy cert for file %s", certFile)
			}
			containerCertFile := m.containerSSL(certDir, certFile)
			err = utils.SetFile(containerCertFile, content)
			if err != nil {
				logrus.Fatal(err)
				return
			}
		}
	}
}

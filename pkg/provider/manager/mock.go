package manager

import (
	"embed"
	"path/filepath"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
)

func CreateDummyCerts(namespace string) {
	m := manager{
		namespace: namespace,
	}
	for certDir, certFiles := range UserProvidedCertificates {
		containerCertDir := m.containerSSL(certDir)
		if err := utils.CreateDirectory(containerCertDir); err != nil {
			logrus.Fatal(err)
			return
		}
		for _, certFile := range certFiles {
			content, err := getDummyContent(certDir, certFile)
			if err != nil {
				logrus.Fatal(err)
			}
			containerCertFile := m.containerSSL(certDir, certFile)
			err = utils.SetFile(containerCertFile, []byte(content))
			if err != nil {
				logrus.Fatal(err)
				return
			}
		}
	}
}

//go:embed testdata
var testData embed.FS

func getDummyContent(certDir, certFile string) (string, error) {
	content, err := testData.ReadFile(filepath.Join("testdata", certDir, certFile))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

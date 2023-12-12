package manager

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var (
	defaultNamespace = "cattle-windows-gmsa-system"
)

func TestCertificateManager(t *testing.T) {
	testCases := []struct {
		Name      string
		Namespace string

		ExpectedCertificates TLSCertificates
	}{
		{
			Name:      "Default",
			Namespace: defaultNamespace,
			ExpectedCertificates: TLSCertificates{
				CertFile: filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "server", "tls.crt"),
				KeyFile:  filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "server", "tls.key"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			utils.SetupTestEnv()
			CreateDummyCerts(tc.Namespace)

			m := New(tc.Namespace)

			// Phase 0: check that createDummyCerts populated only the expected certs
			expectedPaths := []string{
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "ca", "ca.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "ca", "tls.crt"),

				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "client", "ca.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "client", "tls.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "client", "tls.key"),

				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "server", "ca.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "server", "tls.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "container", "ssl", "server", "tls.key"),
			}
			paths, err := getPaths(utils.ProviderDirectory)
			assert.Nil(t, err)
			assert.Equal(t, expectedPaths, paths, "CreateDummyCerts did not produce expected paths")

			// Phase 1: Before starting, nil certificates should be returned
			var certificates *TLSCertificates
			assert.Equal(t, certificates, m.ServerCertificates(), "certificate should be empty before start")

			// Phase 2a: Start should work
			err = m.Start(context.Background())
			assert.Nil(t, err, "could not start manager")

			// Phase 2b: after starting, certificates should be filled in
			certificates = &tc.ExpectedCertificates
			assert.Equal(t, certificates, m.ServerCertificates(), "certificate should be empty before start")

			// Phase 2c: after starting, additional certs should be copied over
			startExpectedPaths := append(expectedPaths, []string{
				// generated from powershell
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "ca", "ca.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "ca", "tls.crt"),

				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "client", "ca.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "client", "tls.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "client", "tls.key"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "client", "tls.pfx"),

				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "server", "ca.crt"),
				filepath.Join(utils.ProviderDirectory, tc.Namespace, "ssl", "server", "tls.crt"),
			}...)
			sort.Strings(startExpectedPaths)
			paths, err = getPaths(utils.ProviderDirectory)
			sort.Strings(paths)
			assert.Nil(t, err)
			assert.Equal(t, startExpectedPaths, paths, "did not find expected paths after start")

			// Phase 3a: Clean should work
			err = m.Clean(context.Background())
			assert.Nil(t, err, "could not clean up manager")

			// Phase 3b: Nothing should be left behind
			var noPaths []string
			paths, err = getPaths(utils.ProviderDirectory)
			assert.Nil(t, err)
			assert.Equal(t, noPaths, paths)
		})
	}

	t.Run("Missing Certs", func(t *testing.T) {
		utils.SetupTestEnv()

		m := New(defaultNamespace)
		err := m.Start(context.Background())
		assert.NotNil(t, err, "manager should not have started with missing certs")
	})
}

func getPaths(dir string) ([]string, error) {
	var paths []string
	var g func(string) error
	g = func(dir string) error {
		readDir, err := utils.ReadDirectory(dir)
		if err != nil {
			return err
		}
		for _, fileInfo := range readDir {
			fullname := filepath.Join(dir, fileInfo.Name())
			if fileInfo.IsDir() {
				if err := g(fullname); err != nil {
					return err
				}
				continue
			}
			paths = append(paths, fullname)
		}
		return nil
	}
	return paths, g(dir)
}

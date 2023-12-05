package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/manager"
	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

var (
	defaultNamespace = "cattle-windows-gmsa-system"

	baseDirectory = filepath.Join(utils.ProviderDirectory, defaultNamespace)

	expectedUserCerts = []string{
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "ca"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "ca", "ca.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "ca", "tls.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "client"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "client", "ca.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "client", "tls.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "client", "tls.key"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "server"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "server", "ca.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "server", "tls.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "server", "tls.key"),
	}

	// generated on running powershell command
	expectedPfxFile = filepath.Join(utils.ProviderDirectory, defaultNamespace, "container", "ssl", "client", "tls.pfx")

	expectedCopiedCerts = []string{
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "ca"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "ca", "ca.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "ca", "tls.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "client"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "client", "ca.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "client", "tls.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "client", "tls.key"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "client", "tls.pfx"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "server"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "server", "ca.crt"),
		filepath.Join(utils.ProviderDirectory, defaultNamespace, "ssl", "server", "tls.crt"),
	}

	expectedPortFile = filepath.Join(utils.ProviderDirectory, defaultNamespace, "port.txt")
)

type dummySecretGetter struct{}

func (g *dummySecretGetter) Get(_, _ string) (*corev1.Secret, error) {
	return nil, fmt.Errorf("unimplemented")
}

func TestRun(t *testing.T) {
	testCases := []struct {
		Name              string
		UserCertsProvided bool
		DisableMTLS       bool
		SkipArtifacts     bool

		UseRealFilesystem bool
		ExpectFailure     bool
	}{
		{
			Name:              "Default",
			UserCertsProvided: true,
			DisableMTLS:       false,

			// when tls is enabled and user certs are provided, we cannot use
			// a mocked filesystem since this will fail on the http server initializing
			UseRealFilesystem: true,
		},
		{
			Name:              "TLS and No Certs",
			UserCertsProvided: false,
			DisableMTLS:       false,

			ExpectFailure: true,
		},
		{
			Name:              "No TLS",
			UserCertsProvided: true,
			DisableMTLS:       true,
		},
		{
			Name:              "No TLS and No Certs",
			UserCertsProvided: false,
			DisableMTLS:       true,
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		assert.FailNow(t, "unable to get current working directory")
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.UseRealFilesystem {
				t.Log("Using a real filesystem for this test")
				utils.SetupEnv()
				testWd, err := os.MkdirTemp("", "")
				if err != nil {
					assert.FailNowf(t, "unable to set up temporary working directory for test %s", tc.Name)
				}
				t.Logf("Executing test in temporary working directory %s", testWd)
				t.Cleanup(func() {
					os.Remove(testWd)
					os.Chdir(cwd)
				})
				err = os.Chdir(testWd)
				if err != nil {
					assert.FailNowf(t, "unable to switch to temporary working directory for test %s", tc.Name)
				}
			} else {
				t.Log("Using mocked filesystem for this test")
				utils.SetupTestEnv()
			}
			if tc.UserCertsProvided {
				manager.CreateDummyCerts(defaultNamespace)
			}

			initialPaths, err := getPathsWithDirectories(utils.ProviderDirectory)
			assert.Nil(t, err, "unable to get paths before clean")
			sort.Strings(initialPaths)

			var expectedPaths []string
			if tc.UserCertsProvided {
				expectedPaths = append(expectedPaths, baseDirectory)
				expectedPaths = append(expectedPaths, expectedUserCerts...)
			}
			assert.Equal(t, expectedPaths, initialPaths, "unable to verify that certs have already been created")

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(1 * time.Second)
				cancel()
			}()
			defer func() {
				if t.Failed() {
					// Wait for context to finish canceling the run operation
					//
					// Without this, an error returned on run that comes after
					// the watcher.QuitOnChange operation is executed would return
					// prematurely while still panicking on file changes
					//
					// As a result, if we are using a real filesystem for this test,
					// a panic would be triggered on cleaning up the temporary working
					// directory since the watcher is still watching it for changes.
					<-ctx.Done()
					return
				}
			}()
			err = run(ctx, &dummySecretGetter{}, defaultNamespace, tc.DisableMTLS, tc.SkipArtifacts)
			if tc.ExpectFailure {
				assert.NotNil(t, err, "should not have been able to run")
				paths, err := getPathsWithDirectories(utils.ProviderDirectory)
				assert.Nil(t, err, "unable to get paths after run")
				assert.Equal(t, expectedPaths, paths, "failure to run should create no new files")
				return
			}
			assert.Nil(t, err, "unable to run")

			if expectedPaths == nil {
				// add base directory with port.txt file since it hasn't been added yet
				expectedPaths = append(expectedPaths, baseDirectory)
			} else if !tc.DisableMTLS {
				expectedPaths = append(expectedPaths, expectedCopiedCerts...)
				expectedPaths = append(expectedPaths, expectedPfxFile)
			}
			expectedPaths = append(expectedPaths, expectedPortFile)
			sort.Strings(expectedPaths)
			paths, err := getPathsWithDirectories(utils.ProviderDirectory)
			sort.Strings(paths)
			assert.Nil(t, err, "unable to get paths after run")
			assert.Equal(t, expectedPaths, paths, "run should add expected paths")
		})
	}
}

func getPathsWithDirectories(dir string) ([]string, error) {
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
			}
			paths = append(paths, fullname)
		}
		return nil
	}
	return paths, g(dir)
}

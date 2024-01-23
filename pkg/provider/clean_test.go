package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/manager"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestClean(t *testing.T) {
	t.Run("Empty Clean", func(t *testing.T) {
		utils.SetupTestEnv()

		initialPaths, err := getPathsWithDirectories(utils.ProviderDirectory)
		assert.Nil(t, err, "unable to get paths before clean")
		assert.Nil(t, err, "unable to verify if paths are initially empty")

		err = Clean(context.Background(), defaultNamespace)
		assert.Nil(t, err, "unable to clean")

		paths, err := getPathsWithDirectories(utils.ProviderDirectory)
		assert.Nil(t, err, "unable to get paths after clean")
		assert.Equal(t, initialPaths, paths, "clean before run should do no operations")
	})

	t.Run("Cleanup Full", func(t *testing.T) {
		utils.SetupTestEnv()
		manager.CreateDummyCerts(defaultNamespace)

		initialPaths, err := getPathsWithDirectories(utils.ProviderDirectory)
		assert.Nil(t, err, "unable to get paths before clean")
		sort.Strings(initialPaths)
		assert.Equal(t, append([]string{baseDirectory}, expectedUserCerts...), initialPaths, "unable to verify that certs have already been created")

		err = Clean(context.Background(), defaultNamespace)
		assert.Nil(t, err, "unable to clean")

		paths, err := getPathsWithDirectories(utils.ProviderDirectory)
		assert.Nil(t, err, "unable to get paths after clean")
		assert.Nil(t, paths, "clean should have cleaned up directory")
	})
}

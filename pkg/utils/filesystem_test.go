package utils

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemOperations(t *testing.T) {
	SetupTestEnv()

	t.Run("ReadDir on empty directory", func(t *testing.T) {
		fileInfos, err := ReadDirectory(ProviderDirectory)
		assert.Nil(t, err, "expected to be able to read empty directory")
		assert.Nil(t, fileInfos, "expected no files to be returned from empty directory")
	})

	paths, err := getPathsWithDirectories(ProviderDirectory)
	assert.Nil(t, err, "could not get initial paths")
	assert.Nil(t, paths, "initial filesystem should have no paths")

	t.Run("Out Of Scope Path", func(t *testing.T) {
		outOfScopePath := "out-of-scope"

		_, err := ReadDirectory(outOfScopePath)
		assert.NotNil(t, err, "ReadDirectory should have returned error for out of scope path")

		err = CreateDirectory(outOfScopePath)
		assert.NotNil(t, err, "CreateDirectory should have returned error for out of scope path")

		_, err = DirectoryExists(outOfScopePath)
		assert.NotNil(t, err, "DirectoryExists should have returned error for out of scope path")

		err = DeleteDirectory(outOfScopePath)
		assert.NotNil(t, err, "DeleteDirectory should have returned error for out of scope path")

		_, err = FileExists(outOfScopePath)
		assert.NotNil(t, err, "FileExists should have returned error for out of scope path")

		_, err = FileExists(outOfScopePath)
		assert.NotNil(t, err, "FileExists should have returned error for out of scope path")

		_, err = FileHas(outOfScopePath, nil)
		assert.NotNil(t, err, "FileHas should have returned error for out of scope path")

		_, err = GetFile(outOfScopePath)
		assert.NotNil(t, err, "GetFile should have returned error for out of scope path")

		err = DeleteFile(outOfScopePath)
		assert.NotNil(t, err, "DeleteFile should have returned error for out of scope path")

		err = SetFile(outOfScopePath, nil)
		assert.NotNil(t, err, "SetFile should have returned error for out of scope path")

		_, _, err = RenameTempFile(outOfScopePath)
		assert.NotNil(t, err, "RenameTempFile should have returned error for out of scope path")

		paths, err := getPathsWithDirectories(ProviderDirectory)
		assert.Nil(t, err, "could not get paths after performing invalid operations")
		assert.Nil(t, paths, "all errors for out-of-scope path should not result in any filesystem mutations")
	})

	dummyDirectory := filepath.Join(ProviderDirectory, "dummy")
	t.Run("Check Directory Operations", func(t *testing.T) {
		exists, err := DirectoryExists(dummyDirectory)
		assert.Nil(t, err, "should be able to check if directory exists")
		assert.False(t, exists, "directory %s should not exist", dummyDirectory)

		// Create

		err = CreateDirectory(dummyDirectory)
		assert.Nil(t, err, "should have been able to create directory %s", dummyDirectory)

		exists, err = DirectoryExists(dummyDirectory)
		assert.Nil(t, err, "should be able to check if directory exists after create")
		assert.True(t, exists, "directory %s should exist after create", dummyDirectory)

		err = CreateDirectory(dummyDirectory)
		assert.Nil(t, err, "creating existing directory %s should not return error", dummyDirectory)

		exists, err = DirectoryExists(dummyDirectory)
		assert.Nil(t, err, "should be able to check if directory exists after second create")
		assert.True(t, exists, "directory %s should exist even after second create", dummyDirectory)

		fileInfos, err := ReadDirectory(dummyDirectory)
		assert.Nil(t, err, "expected to be able to read empty directory")
		assert.Nil(t, fileInfos, "expected no files to be returned from empty directory")

		// Delete

		err = DeleteDirectory(dummyDirectory)
		assert.Nil(t, err, "should have been able to delete directory %s", dummyDirectory)

		exists, err = DirectoryExists(dummyDirectory)
		assert.Nil(t, err, "should be able to check if directory exists after delete")
		assert.False(t, exists, "directory %s should not exist after delete", dummyDirectory)

		err = DeleteDirectory(dummyDirectory)
		assert.Nil(t, err, "deleting non-existent directory %s should not return error", dummyDirectory)

		exists, err = DirectoryExists(dummyDirectory)
		assert.Nil(t, err, "should be able to check if directory exists after second delete")
		assert.False(t, exists, "directory %s should not exist even after second delete", dummyDirectory)
	})
}

func getPathsWithDirectories(dir string) ([]string, error) {
	var paths []string
	var g func(string) error
	g = func(dir string) error {
		readDir, err := ReadDirectory(dir)
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

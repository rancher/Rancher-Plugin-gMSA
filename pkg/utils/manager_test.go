package utils

import (
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/stretchr/testify/assert"
)

func TestFilesystem(t *testing.T) {
	path1Fs := memfs.New()
	path2Fs := memfs.New()
	manager := &fsManager{
		filesystems: map[string]billy.Filesystem{
			filepath.Join("path", "1"): path1Fs,
			filepath.Join("path", "2"): path2Fs,
		},
	}

	testCases := []struct {
		Name string
		Path string

		ExpectedFilesystem billy.Filesystem
		ExpectedSubpath    string
		ShouldThrowError   bool
	}{
		{
			Name: "Base Path 1",
			Path: filepath.Join("path", "1"),

			ExpectedFilesystem: path1Fs,
			ExpectedSubpath:    "",
		},
		{
			Name: "Base Path 2",
			Path: filepath.Join("path", "2"),

			ExpectedFilesystem: path2Fs,
			ExpectedSubpath:    "",
		},
		{
			Name: "Sub Path 1",
			Path: filepath.Join("path", "1", "rancher"),

			ExpectedFilesystem: path1Fs,
			ExpectedSubpath:    "rancher",
		},
		{
			Name: "Sub Path 2",
			Path: filepath.Join("path", "2", "cattle"),

			ExpectedFilesystem: path2Fs,
			ExpectedSubpath:    "cattle",
		},
		{
			Name: "Nested Path 1",
			Path: filepath.Join("path", "1", "rancher", "cattle", "windows"),

			ExpectedFilesystem: path1Fs,
			ExpectedSubpath:    filepath.Join("rancher", "cattle", "windows"),
		},
		{
			Name: "Nested Path 2",
			Path: filepath.Join("path", "2", "cattle", "windows", "rancher"),

			ExpectedFilesystem: path2Fs,
			ExpectedSubpath:    filepath.Join("cattle", "windows", "rancher"),
		},
		{
			Name: "Untracked Path",
			Path: filepath.Join("path", "doesnotexist"),

			ShouldThrowError: true,
		},
		{
			Name: "Root Path",
			Path: "path",

			ShouldThrowError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			fs, subpath, err := manager.Filesystem(tc.Path)
			if tc.ShouldThrowError {
				assert.NotNil(t, err, "expected error to be thrown")
				return
			}
			assert.Equal(t, tc.ExpectedFilesystem, fs)
			assert.Equal(t, tc.ExpectedSubpath, subpath)
		})
	}
}

func TestUseFs(t *testing.T) {
	t.Run("Use OS Filesystem", func(t *testing.T) {
		SetupEnv()
		assert.False(t, manager.isMemFs, "expected isMemFs to be false")
		assert.False(t, IsMemFs(), "expected IsMemFs to return false")
		for path, fs := range manager.filesystems {
			assert.IsTypef(t, osfs.New("."), fs, "expected filesystem at %s to be osFs", path)
		}
	})
	t.Run("Use Memory Filesystem", func(t *testing.T) {
		SetupTestEnv()
		assert.True(t, manager.isMemFs, "expected isMemFs to be true")
		assert.True(t, IsMemFs(), "expected IsMemFs to return true")
		for path, fs := range manager.filesystems {
			assert.IsTypef(t, memfs.New(), fs, "expected filesystem at %s to be memFs", path)
		}
	})
}

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	butil "github.com/go-git/go-billy/v5/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

var DeleteBackoff = wait.Backoff{
	Steps:    5,
	Duration: 100 * time.Millisecond,
	Factor:   2.0,
	Jitter:   0.1,
}

func ReadDirectory(dir string) ([]os.FileInfo, error) {
	fs, path, err := manager.Filesystem(dir)
	if err != nil {
		return nil, err
	}
	return fs.ReadDir(path)
}

func CreateDirectory(dir string) error {
	fs, path, err := manager.Filesystem(dir)
	if err != nil {
		return err
	}
	return fs.MkdirAll(path, os.ModePerm)
}

func DirectoryExists(dir string) (bool, error) {
	fs, path, err := manager.Filesystem(dir)
	if err != nil {
		return false, err
	}
	fileInfo, err := fs.Lstat(path)
	if err == nil {
		return fileInfo.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func DeleteDirectory(dir string) error {
	fs, path, err := manager.Filesystem(dir)
	if err != nil {
		return err
	}
	err = wait.ExponentialBackoff(DeleteBackoff, func() (bool, error) {
		// continuously try to remove
		// We retry this process a few times because there may
		// still be instances of CCG referencing the DLL. Windows
		// will prevent the file from being deleted if any references
		// still exist. Eventually, the CCG instances will terminate and
		// all references will disappear, at which point the file can be
		// deleted.
		//
		// It goes without saying that if you're uninstalling this plugin,
		// you shouldn't be running workloads which need to use the plugin.
		// directory exists, so remove it
		err = butil.RemoveAll(fs, path)
		if err == nil {
			return true, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		logrus.Infof("Encountered error while deleting %s: %s", dir, err)
		return false, nil
	})
	return err
}

func FileExists(dir string) (bool, error) {
	fs, path, err := manager.Filesystem(dir)
	if err != nil {
		return false, err
	}
	fileInfo, err := fs.Lstat(path)
	if err == nil {
		return !fileInfo.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func FileHas(file string, content []byte) (bool, error) {
	exists, err := FileExists(file)
	if err != nil {
		return false, err
	}
	if !exists {
		return content == nil, nil
	}
	// file exists, so get its contents
	fileContent, err := GetFile(file)
	if err != nil {
		return false, err
	}
	// file has contents, so compare it
	return bytes.Equal(fileContent, content), nil
}

func GetFile(file string) ([]byte, error) {
	// file exists, so get its contents
	fs, path, err := manager.Filesystem(file)
	if err != nil {
		return nil, err
	}
	return butil.ReadFile(fs, path)
}

func DeleteFile(file string) error {
	fs, path, err := manager.Filesystem(file)
	if err != nil {
		return err
	}
	wait.ExponentialBackoff(DeleteBackoff, func() (bool, error) {
		// continuously try to remove
		// We retry this process a few times because there may
		// still be instances of CCG referencing the DLL. Windows
		// will prevent the file from being deleted if any references
		// still exist. Eventually, the CCG instances will terminate and
		// all references will disappear, at which point the file can be
		// deleted.
		//
		// It goes without saying that if you're uninstalling this plugin,
		// you shouldn't be running workloads which need to use the plugin.
		var exists bool
		exists, err = FileExists(file)
		if err != nil {
			return false, err
		}
		if !exists {
			return true, nil
		}
		// directory exists, so remove it
		err = fs.Remove(path)
		if err == nil {
			return true, nil
		}
		logrus.Debugf("Encountered error while deleting %s: %s", file, err)
		return false, nil
	})
	return err
}

func SetFile(file string, content []byte) error {
	fs, path, err := manager.Filesystem(file)
	if err != nil {
		return err
	}
	f, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %s", file, err)
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", file, err)
	}
	return nil
}

func RenameTempFile(file string) (func() error, func() error, error) {
	exists, err := FileExists(file)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, os.ErrNotExist
	}
	fs, path, err := manager.Filesystem(file)
	if err != nil {
		return nil, nil, err
	}
	f, err := fs.TempFile(fs.Root(), "")
	if err != nil {
		return nil, nil, err
	}
	if err := f.Close(); err != nil {
		return nil, nil, err
	}
	tempFile := f.Name()
	_, tempPath, err := manager.Filesystem(tempFile)
	if err != nil {
		return nil, nil, err
	}
	undoFunc := func() error {
		if err := fs.Rename(tempPath, path); err != nil {
			return fmt.Errorf("unable to rename %s back to %s: %s", tempFile, file, err)
		}
		return nil
	}
	deleteFunc := func() error {
		if err := DeleteFile(tempFile); err != nil {
			return fmt.Errorf("unable to delete temporary file %s: %s", tempFile, err)
		}
		return nil
	}
	renameFunc := os.Rename
	if IsMemFs() {
		// in a dry-run scenario, running os.Remove would fail so this mimics that process
		// Note: the reason why we use os.Remove in the real-life scenario is because that is
		// how Microsoft recommends performing DLL Upgrades
		renameFunc = func(oldpath string, newpath string) error {
			content, err := GetFile(oldpath)
			if err != nil {
				return err
			}
			err = SetFile(newpath, content)
			if err != nil {
				return err
			}
			err = DeleteFile(oldpath)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return undoFunc, deleteFunc, renameFunc(file, tempFile)
}

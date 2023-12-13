package status

import (
	"path/filepath"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	DummyCcgCOMClassExists = false
	DummyCLSIDExists       = false
)

func CheckInstallationStatus(dll []byte) (status DLLInstallationStatus, err error) {
	status = DLLInstallationStatus{}

	// 1. Check that the DLL directory exists C:\Program Files\RanchergMSACredentialProvider
	status.DirectoryExists, err = utils.DirectoryExists(utils.DLLDirectory)
	if err != nil {
		return status, err
	}

	if status.DirectoryExists {
		logrus.Debug("detected dll directory exists")
	} else {
		logrus.Debug("dll directory does not exist")
	}

	// 2. Check that the DLL exists inside the directory
	status.Exists, err = utils.FileExists(filepath.Join(utils.DLLDirectory, utils.DLLPath))
	if err != nil {
		return status, err
	}

	if status.Exists {
		logrus.Debug("detected dll file exists")
	} else {
		logrus.Debug("dll directory does not exist")
	}

	status.AddedToCOMClassesKey, err = ccgCOMClassExists(utils.CCGCOMClassesKey)
	if err != nil {
		return status, err
	}

	if status.AddedToCOMClassesKey {
		logrus.Debug("detected dll has been added to COM classes key")
	} else {
		logrus.Debug("dll directory has not been registered under COM classes")
	}

	status.AddedToHKEYClassesRoot, err = clsidExists(utils.CLSIDKey)
	if err != nil {
		return status, err
	}

	if status.AddedToHKEYClassesRoot {
		logrus.Debug("detected dll has been added to the HKEY classes root")
	} else {
		logrus.Debug("dll has not been added to HKEY classes root")
	}

	if status.Exists {
		doesNotNeedUpgrade, err := utils.FileHas(filepath.Join(utils.DLLDirectory, utils.DLLPath), dll)
		if err != nil {
			return status, err
		}
		status.NeedsUpgrade = !doesNotNeedUpgrade
		if status.NeedsUpgrade {
			logrus.Debug("detected dll requires an upgrade")
		} else {
			logrus.Debug("dll does not require upgrade")
		}
	}
	return status, err
}

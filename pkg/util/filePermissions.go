package util

import (
	"os"

	"github.com/apex/log"
	fsutil "github.com/develar/go-fs-util"
)

// https://github.com/electron-userland/electron-builder/issues/3452#issuecomment-438619535
// quite a lot sources don't have proper permissions to be distributed
func FixPermissions(filePath string, fileMode os.FileMode) error {
	original, fixed, err := fsutil.FixPermissions(filePath, fileMode)
	if err != nil {
		return err
	}

	if original == fixed {
		return nil
	}

	log.WithFields(log.Fields{
		"file":                filePath,
		"reason":              "group or other cannot read",
		"originalPermissions": original.String(),
		"newPermissions":      fixed.String(),
	}).Debug("fix permissions")
	return nil
}

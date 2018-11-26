package util

import (
	"github.com/apex/log"
	"github.com/develar/errors"
	"github.com/phayes/permbits"
)

// https://github.com/electron-userland/electron-builder/issues/3452#issuecomment-438619535
// quite a lot sources don't have proper permissions to be distributed
func FixPermissions(filePath string, permissions permbits.PermissionBits) (bool, error) {
	originalPermissions := permissions
	if permissions.UserExecute() {
		permissions.SetGroupExecute(true)
		permissions.SetOtherExecute(true)
	}

	permissions.SetGroupRead(true)
	permissions.SetOtherRead(true)

	if originalPermissions == permissions {
		return false, nil
	}

	log.WithFields(log.Fields{
		"file":                filePath,
		"reason":              "group or other cannot read",
		"originalPermissions": originalPermissions.String(),
		"newPermissions":      permissions.String(),
	}).Debug("fix permissions")
	err := permbits.Chmod(filePath, permissions)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

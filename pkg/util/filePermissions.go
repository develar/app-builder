package util

import (
	"os"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/go-fs-util"
	"go.uber.org/zap"
)

// https://github.com/electron-userland/electron-builder/issues/3452#issuecomment-438619535
// quite a lot sources don't have proper permissions to be distributed
func FixPermissions(filePath string, fileMode os.FileMode, isForceSetIfExecutable bool) error {
	original, fixed, err := fsutil.FixPermissions(filePath, fileMode, isForceSetIfExecutable)
	if err != nil {
		return err
	}

	if original == fixed {
		return nil
	}

	log.Debug("fix permissions",
		zap.String("file", filePath),
		zap.String("reason", "group or other cannot read"),
		zap.Stringer("originalPermissions", original),
		zap.Stringer("newPermissions", fixed),
	)
	return nil
}

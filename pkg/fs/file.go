package fs

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/phayes/permbits"
)

func SetNormalDirPermissions(path string) error {
	// https://github.com/electron-userland/electron-builder/issues/2682
	// always set dir permission to 0755 regardless of what was originally
	if runtime.GOOS != "windows" {
		return permbits.Chmod(path, 0755)
	}
	return nil
}

// https://github.com/electron-userland/electron-builder/issues/2654#issuecomment-369972916
// https://github.com/electron-userland/electron-builder/issues/3452#issuecomment-438619535
func SetNormalFilePermissions(path string) error {
	if runtime.GOOS != "windows" {
		return permbits.Chmod(path, 0644)
	}
	return nil
}

func ReadFile(file string, size int) ([]byte, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]byte, size)
	_, err = reader.Read(result)
	return result, fsutil.CloseAndCheckError(err, reader)
}

func createFileAndCreateParentDirIfNeeded(name string) (*os.File, error) {
	flag := os.O_WRONLY|os.O_CREATE|os.O_TRUNC
	// cannot use file mode as is because of *** *** *** umask
	file, err := os.OpenFile(name, flag, 0666)
	if err == nil {
		return file, nil
	}

	if !os.IsNotExist(err) {
		return nil, errors.WithStack(err)
	}

	dir := filepath.Dir(name)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = SetNormalDirPermissions(dir)
	if err != nil {
		return nil, err
	}

	file, err = os.OpenFile(name, flag, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return file, nil
}

func WriteFileAndRestoreNormalPermissions(source io.Reader, to string, fileMode os.FileMode, buffer []byte) error {
	destinationFile, err := createFileAndCreateParentDirIfNeeded(to)
	if err != nil {
		return err
	}

	_, err = io.CopyBuffer(destinationFile, source, buffer)
	if err != nil {
		_ = destinationFile.Close()
		return errors.WithStack(err)
	}

	err = destinationFile.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	err = fixPermissions(to, fileMode)
	if err != nil {
		return err
	}

	return nil
}

func fixPermissions(filePath string, fileMode os.FileMode) error {
	originalPermissions := permbits.PermissionBits(fileMode)
	permissions := originalPermissions

	if originalPermissions.UserExecute() {
		permissions.SetGroupExecute(true)
		permissions.SetOtherExecute(true)
	}

	permissions.SetUserRead(true)
	permissions.SetGroupRead(true)
	permissions.SetOtherRead(true)

	permissions.SetSetuid(false)
	permissions.SetSetgid(false)

	return errors.WithStack(permbits.Chmod(filePath, permissions))
}
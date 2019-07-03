package fs

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/oxtoacart/bpool"
	"go.uber.org/zap"
)

var bufferPool = bpool.NewBytePool(runtime.NumCPU(), 64*1024)

type FileCopier struct {
	IsUseHardLinks bool
}

// go doesn't provide native copy operation (CoW)
func (t *FileCopier) copyDir(from string, to string) error {
	fileNames, err := fsutil.ReadDirContent(from)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, name := range fileNames {
		if name == ".DS_Store" {
			continue
		}

		err = t.copyDirOrFile(filepath.Join(from, name), filepath.Join(to, name), false)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func CopyUsingHardlink(from string, to string) error {
	var fileCopier FileCopier
	fileCopier.IsUseHardLinks = true
	return fileCopier.CopyDirOrFile(from, to)
}

func CopyDirOrFile(from string, to string) error {
	var fileCopier FileCopier
	return fileCopier.CopyDirOrFile(from, to)
}

func (t *FileCopier) CopyDirOrFile(from string, to string) error {
	if runtime.GOOS == "windows" {
		t.IsUseHardLinks = false
	}

	log.Debug("copy files", zap.String("from", from), zap.String("to", to), zap.Bool("isUseHardLinks", t.IsUseHardLinks))
	err := t.copyDirOrFile(from, to, true)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (t *FileCopier) copyDirOrFile(from string, to string, isCreateParentDirs bool) error {
	fromInfo, err := os.Lstat(from)
	if err != nil {
		return errors.WithStack(err)
	}

	if fromInfo.IsDir() {
		// cannot use file mode as is because of *** *** *** umask
		if isCreateParentDirs {
			err = fsutil.EnsureDir(to)
		} else {
			err = os.Mkdir(to, 0777)
		}
		if err != nil && !os.IsExist(err) {
			return errors.WithStack(err)
		}

		err = SetNormalDirPermissions(to)
		if err != nil {
			return err
		}

		return t.copyDir(from, to)
	}

	if isCreateParentDirs {
		err := fsutil.EnsureDir(filepath.Dir(to))
		if err != nil {
			return err
		}
	}

	if (fromInfo.Mode() & os.ModeSymlink) != 0 {
		return t.createSymlink(from, to)
	} else {
		return t.CopyFile(from, to, isCreateParentDirs, fromInfo)
	}
}

func (t *FileCopier) CopyFile(from string, to string, isCreateParentDirs bool, fromInfo os.FileInfo) error {
	if t.IsUseHardLinks {
		err := os.Link(from, to)
		if err == nil {
			return nil
		}

		t.IsUseHardLinks = false
		log.Debug("cannot copy using hard link", zap.Error(err), zap.String("from", from), zap.String("to", to))
	}

	return CopyFileAndRestoreNormalPermissions(from, to, fromInfo.Mode())
}

func CopyFileAndRestoreNormalPermissions(from string, to string, fileMode os.FileMode) error {
	sourceFile, err := os.Open(from)
	if err != nil {
		return errors.WithStack(err)
	}

	defer util.Close(sourceFile)
	buffer := bufferPool.Get()
	err = WriteFileAndRestoreNormalPermissions(sourceFile, to, fileMode, buffer)
	bufferPool.Put(buffer)
	if err != nil {
		return err
	}
	return nil
}

func (t *FileCopier) createSymlink(from string, to string) error {
	link, err := os.Readlink(from)
	if err != nil {
		return errors.WithStack(err)
	}

	if filepath.IsAbs(link) {
		link, err = filepath.Rel(filepath.Dir(from), link)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = os.Symlink(link, to)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

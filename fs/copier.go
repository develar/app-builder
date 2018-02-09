package fs

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

type FileCopier struct {
	IsUseHardLinks bool

	links []LinkInfo
}

type LinkInfo struct {
	file string
	link string
}

// go doesn't provide native copy operation (CoW)
func (fileCopier *FileCopier) copyDir(from string, to string, fromInfo os.FileInfo) error {
	fileNames, err := ReadDirContent(from)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, name := range fileNames {
		if name == ".DS_Store" {
			continue
		}

		err = fileCopier.copyDirOrFile(filepath.Join(from, name), filepath.Join(to, name), false)
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

func (fileCopier *FileCopier) CopyDirOrFile(from string, to string) error {
	if runtime.GOOS == "windows" {
		fileCopier.IsUseHardLinks = false
	}

	log.WithFields(log.Fields{
		"from": from,
		"to": to,
		"isUseHardLinks": fileCopier.IsUseHardLinks,
	}).Debug("copy files")

	err := fileCopier.copyDirOrFile(from, to, true)
	if err != nil {
		return errors.WithStack(err)
	}

	if fileCopier.links != nil {
		for _, linkInfo := range fileCopier.links {
			err = os.Symlink(linkInfo.link, linkInfo.file)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func (fileCopier *FileCopier) copyDirOrFile(from string, to string, isCreateParentDirs bool) error {
	fromInfo, err := os.Lstat(from)
	if err != nil {
		return errors.WithStack(err)
	}

	if fromInfo.IsDir() {
		if isCreateParentDirs {
			err = os.MkdirAll(to, fromInfo.Mode())
		} else {
			err = os.Mkdir(to, fromInfo.Mode())
		}
		if err != nil {
			return errors.WithStack(err)
		}

		return fileCopier.copyDir(from, to, fromInfo)
	} else if fromInfo.Mode() & os.ModeSymlink != 0 {
		return fileCopier.copySymlink(from, to)
	}

	if fileCopier.IsUseHardLinks {
		err = os.Link(from, to)
		if err == nil {
			return nil
		}

		fileCopier.IsUseHardLinks = false
		log.WithError(err).WithField("from", from).WithField("to", to).Debug("cannot copy using hard link")
	}

	return copyFile(from, to, fromInfo)
}

// symlink cannot be created during copy because symlink can point to not yet copied target file
func (fileCopier *FileCopier) copySymlink(from string, to string) error {
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

	fileCopier.links = append(fileCopier.links, LinkInfo{
		file: to,
		link: link,
	})

	return nil
}

func copyFile(from string, to string, fromInfo os.FileInfo) error {
	s, err := os.Open(from)
	if err != nil {
		return errors.WithStack(err)
	}

	defer s.Close()
	d, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fromInfo.Mode())
	if err != nil {
		s.Close()
		return errors.WithStack(err)
	}

	_, err = io.Copy(d, s)
	if err != nil {
		d.Close()
		return errors.WithStack(err)
	}

	err = d.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

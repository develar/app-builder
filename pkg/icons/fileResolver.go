package icons

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/develar/errors"
)

// returns file if exists, null if file not exists, or error if unknown error
func resolveSourceFileOrNull(sourceFile string, roots []string) (string, os.FileInfo, error) {
	if filepath.IsAbs(sourceFile) {
		cleanPath := filepath.Clean(sourceFile)
		fileInfo, err := os.Stat(cleanPath)
		if err == nil {
			return cleanPath, fileInfo, nil
		}
		return "", nil, errors.WithStack(err)
	}

	for _, root := range roots {
		resolvedPath := filepath.Join(root, sourceFile)
		fileInfo, err := os.Stat(resolvedPath)
		switch {
		case err == nil:
			return resolvedPath, fileInfo, nil
		case os.IsNotExist(err):
			log.WithFields(log.Fields{
				"path": resolvedPath,
			}).Debug("path doesn't exist")
		default:
			log.WithFields(log.Fields{
				"path":  resolvedPath,
				"error": err,
			}).Debug("tried resolved path, but got error")
		}
	}

	return "", nil, nil
}

func resolveSourceFile(sourceFiles []string, roots []string) (string, os.FileInfo, error) {
	for _, sourceFile := range sourceFiles {
		resolvedPath, fileInfo, err := resolveSourceFileOrNull(sourceFile, roots)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		if fileInfo != nil {
			return resolvedPath, fileInfo, nil
		}
	}

	return "", nil, nil
}

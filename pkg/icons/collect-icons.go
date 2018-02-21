package icons

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/errors"
)

func CollectIcons(sourceDir string) ([]IconInfo, error) {
	files, err := fs.ReadDirContent(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("icon directory %s doesn't exist", sourceDir)
		}

		fileInfo, statErr := os.Stat(sourceDir)
		if statErr == nil && !fileInfo.IsDir() {
			return nil, errors.Errorf("icon directory %s is not a directory", sourceDir)
		}
		return nil, errors.WithStack(err)
	}

	var result []IconInfo
	re := regexp.MustCompile("[0-9]+")
	for _, file := range files {
		name := file
		if !(strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".PNG")) {
			continue
		}

		sizeString := re.FindString(name)
		if sizeString == "" {
			continue
		}

		size, err := strconv.Atoi(sizeString)
		if err != nil {
			// unrealistic case
			return nil, errors.WithStack(err)
		}

		iconPath := filepath.Join(sourceDir, name)
		result = append(result, IconInfo{iconPath, size})
	}

	if len(result) == 0 {
		return nil, errors.Errorf("icon directory %s doesn't contain icons", sourceDir)
	}

	sortBySize(result)
	return result, nil
}

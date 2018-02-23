package icons

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/errors"
)

func CollectIcons(sourceDir string) ([]IconInfo, string, error) {
	files, err := fs.ReadDirContent(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", errors.Errorf("icon directory %s doesn't exist", sourceDir)
		}

		fileInfo, statErr := os.Stat(sourceDir)
		if statErr == nil && !fileInfo.IsDir() {
			return nil, "", errors.Errorf("icon directory %s is not a directory", sourceDir)
		}
		return nil, "", errors.WithStack(err)
	}

	var result []IconInfo
	re := regexp.MustCompile("[0-9]+")
	var iconFilename string
	sizeToFileName := make(map[int]*IconInfo)
	for _, name := range files {
		if !(strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".PNG")) {
			continue
		}

		sizeString := re.FindString(name)
		if sizeString == "" {
			if name == "icon.png" {
				iconFilename = name
			}
			continue
		}

		size, err := strconv.Atoi(sizeString)
		if err != nil {
			// unrealistic case
			return nil, "", errors.WithStack(err)
		}

		iconPath := filepath.Join(sourceDir, name)

		existing := sizeToFileName[size]
		if existing != nil {
			// 16x16.png vs 16x16-dev.png - select shorter name
			if len(name) >= len(filepath.Base(existing.File)) {
				continue
			} else {
				existing.File = iconPath
				break
			}
		}

		iconInfo := IconInfo{iconPath, size}
		sizeToFileName[size] = &iconInfo
		result = append(result, iconInfo)
	}

	if len(result) == 0 {
		if len(iconFilename) == 0 {
			return nil, "", errors.Errorf("icon directory %s doesn't contain icons", sourceDir)
		}

		log.WithField("iconDir", sourceDir).Debug("icon directory doesn't contain icons ([0-9]+.png), but icon.png exists")
		return nil, filepath.Join(sourceDir, iconFilename), nil
	}

	sortBySize(result)
	return result, "", nil
}

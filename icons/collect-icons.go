package icons

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func CollectIcons(sourceDir string) (*IconListResult, error) {
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("icon directory %s doesn't exist", sourceDir)
		}

		fileInfo, statErr := os.Stat(sourceDir)
		if statErr == nil && !fileInfo.IsDir() {
			return nil, fmt.Errorf("icon directory %s is not a directory", sourceDir)
		}
		return nil, err
	}

	var result []IconInfo
	maxSize := 0
	maxIconPath := ""
	for _, file := range files {
		name := file.Name()
		if !(strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".PNG")) {
			continue
		}

		re := regexp.MustCompile("[0-9]+")
		sizeString := re.FindString(name)
		if sizeString == "" {
			continue
		}

		size, err := strconv.Atoi(sizeString)
		if err != nil {
			// unrealistic case
			return nil, err
		}

		iconPath := filepath.Join(sourceDir, name)
		result = append(result, IconInfo{iconPath, size})

		if size > maxSize {
			maxSize = size
			maxIconPath = iconPath
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("icon directory %s doesn't contain icons", sourceDir)
	}

	return &IconListResult{
		MaxIconPath: maxIconPath,
		MaxIconSize: maxSize,
		Icons:       result,
	}, nil
}

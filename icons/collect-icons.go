package icons

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func CollectIcons(sourceDir string) (*IconListResult, error) {
	dir, err := os.Open(sourceDir)
	if err != nil {
		return nil, err
	}

	files, err := dir.Readdirnames(0)
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
		name := file
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

	sort.Slice(result, func(i, j int) bool { return result[i].Size < result[j].Size })

	return &IconListResult{
		MaxIconPath: maxIconPath,
		MaxIconSize: maxSize,
		Icons:       result,
	}, nil
}

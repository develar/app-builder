package icons

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/develar/app-builder/pkg/linuxTools"
	"github.com/develar/app-builder/pkg/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"go.uber.org/zap"
)

var sameSize = map[string]string{
	"icp5": "ic11",
	"icp6": "ic12",
	"ic08": "ic13",
	"ic09": "ic14",
}

var typeToSize = map[string]int{
	"icp4": 16,
	"icp5": 32,
	"icp6": 64,
	"ic07": 128,
	"ic08": 256,
	"ic09": 512,
	"ic10": 1024,
	"ic11": 32,
	"ic12": 64,
	"ic13": 256,
	"ic14": 512,
}

func ConvertIcnsToPngUsingOpenJpeg(icnsPath string, outDir string) ([]IconInfo, error) {
	reader, err := os.Open(icnsPath)
	defer util.Close(reader)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	bufferedReader := bufio.NewReader(reader)
	subImageInfoList, err := ReadIcns(bufferedReader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var result []IconInfo
	for s1, s2 := range sameSize {
		// icns contains retina icons but with the same size
		if _, ok := subImageInfoList[s1]; ok {
			delete(subImageInfoList, s2)
		}
	}

	outFileNamePrefix := filepath.Join(outDir, strings.TrimSuffix(filepath.Base(icnsPath), filepath.Ext(icnsPath))) + "_"
	for imageType, subImage := range subImageInfoList {
		if isIgnoredType(imageType) {
			log.Debug("skip unsupported icns sub image format", zap.String("type", imageType), zap.String("file", icnsPath))
			continue
		}

		imageOffset := int64(subImage.Offset)
		_, err = reader.Seek(imageOffset, 0)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		bufferedReader.Reset(reader)

		var outFileName string

		config, formatName, err := image.DecodeConfig(bufferedReader)
		if err != nil {
			outFileName = outFileNamePrefix + imageType + ".jp2"
			result = append(result, IconInfo{
				File: outFileName,
				Size: typeToSize[imageType],
			})
		} else {
			outFileName = outFileNamePrefix + fmt.Sprintf("%d.%s", config.Width, formatName)
			result = append(result, IconInfo{
				File: outFileName,
				Size: config.Width,
			})
		}

		_, err = reader.Seek(imageOffset, 0)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		outWriter, err := os.Create(outFileName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		_, err = io.Copy(outWriter, io.LimitReader(reader, int64(subImage.Length)))
		err = fsutil.CloseAndCheckError(err, outWriter)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	err = util.MapAsync(len(result), func(taskIndex int) (func() error, error) {
		imageInfo := &result[taskIndex]
		jpeg2File := imageInfo.File
		if !strings.HasSuffix(jpeg2File, ".jp2") {
			return nil, nil
		}

		opjDecompressPath := "opj_decompress"
		opjLibPath := ""
		if !util.IsEnvTrue("USE_SYSTEM_OPG") && runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			opjDecompressPath, err = linuxTools.GetLinuxTool("opj_decompress")
			if err != nil {
				return nil, errors.WithStack(err)
			}

			opjLibPath = filepath.Join(filepath.Dir(opjDecompressPath), "lib")
		}

		pngFile := fmt.Sprintf("%s%d.png", outFileNamePrefix, imageInfo.Size)
		imageInfo.File = pngFile

		return func() error {
			command := exec.Command(opjDecompressPath, "-quiet", "-i", jpeg2File, "-o", pngFile)
			if len(opjLibPath) != 0 {
				env := os.Environ()
				env = append(env,
					fmt.Sprintf("LD_LIBRARY_PATH=%s", opjLibPath+":"+os.Getenv("LD_LIBRARY_PATH")),
				)
				command.Env = env
			}

			_, err = util.Execute(command)
			if err != nil {
				return err
			}

			err = os.Remove(jpeg2File)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

func isIgnoredType(imageType string) bool {
	return imageType == "ic04" || imageType == "ic05" ||
		strings.HasPrefix(imageType, "icm") || strings.HasPrefix(imageType, "ics") || strings.HasPrefix(imageType, "is") || strings.HasPrefix(imageType, "s") || strings.HasPrefix(imageType, "ich") ||
		imageType == "icl4" ||
		imageType == "icl8" ||
		imageType == "il32" ||
		imageType == "l8mk" ||
		imageType == "ih32" ||
		imageType == "h8mk" ||
		imageType == "it32" ||
		imageType == "t8mk"
}

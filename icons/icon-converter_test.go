package icons

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Flaque/filet"
	"github.com/biessek/golang-ico"
	"github.com/stretchr/testify/assert"
)

func getTestDataPath(t *testing.T) string {
	testDataPath, err := filepath.Abs(filepath.Join("..", "testData"))
	assert.NoError(t, err)
	return testDataPath
}

func TestCheckIcoImageSize(t *testing.T) {
	_, err := ConvertIcon([]string{filepath.Join(getTestDataPath(t), "icon.ico")}, nil, "ico")
	assert.NoError(t, err)
}

func TestIcnsToIco(t *testing.T) {
	files, err := ConvertIcon([]string{filepath.Join(getTestDataPath(t), "icon.icns")}, nil, "ico")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	file := files[0].File

	assert.True(t, strings.HasSuffix(file, ".ico"))

	data, err := ioutil.ReadFile(file)
	assert.NoError(t, err)

	assert.Equal(t, GetIcoSizes(data), []Sizes([]Sizes{
		{Width: 256, Height: 256},
	}))
}

func TestIcnsToPng(t *testing.T) {
	//defer filet.CleanUp(t)

	tmpDir := filet.TmpDir(t, "/tmp")

	result, err := ConvertIcnsToPngUsingOpenJpeg(filepath.Join(getTestDataPath(t), "icon.icns"), tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(result))
}

func TestIcnsToPng2(t *testing.T) {
	//defer filet.CleanUp(t)

	tmpDir := filet.TmpDir(t, "/tmp")

	result, err := ConvertIcnsToPngUsingOpenJpeg(filepath.Join(getTestDataPath(t), "icon-jpeg2.icns"), tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

func TestLargePngTo256Ico(t *testing.T) {
	files, err := ConvertIcon([]string{filepath.Join(getTestDataPath(t), "512x512.png")}, nil, "ico")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files))
	file := files[0].File

	assert.True(t, strings.HasSuffix(file, ".ico"))

	reader, err := os.Open(file)
	assert.NoError(t, err)
	defer reader.Close()
	images, err := ico.DecodeAll(reader)
	assert.NoError(t, err)

	assert.Equal(t, len(images), 1)

	imageSize := images[0].Bounds().Max
	assert.Equal(t, 256, imageSize.X)
	assert.Equal(t, 256, imageSize.Y)
}

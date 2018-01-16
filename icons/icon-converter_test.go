package icons

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/biessek/golang-ico"
	"github.com/stretchr/testify/assert"
)

func TestCheckIcoImageSize(t *testing.T) {
	_, err := ConvertIcon(filepath.Join("..", "testData", "icon.ico"), nil, "ico")
	assert.NoError(t, err)
}

func TestIcnsToIco(t *testing.T) {
	file, err := ConvertIcon(filepath.Join("..", "testData", "icon.icns"), nil, "ico")
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(file, ".ico"))

	data, err := ioutil.ReadFile(file)
	assert.NoError(t, err)

	assert.Equal(t, GetIcoSizes(data), []Sizes([]Sizes{
		{Width: 256, Height: 256},
	}))
}

func TestLargePngTo256Ico(t *testing.T) {
	file, err := ConvertIcon(filepath.Join("..", "testData", "512x512.png"), nil, "ico")
	assert.NoError(t, err)
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

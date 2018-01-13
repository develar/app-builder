package icons

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIcoImageSize(t *testing.T) {
	_, err := ConvertIcon(filepath.Join("..", "testData", "icon.ico"), nil, "ico")
	assert.Nil(t, err)
}

func TestIcnsToIco(t *testing.T) {
	file, err := ConvertIcon(filepath.Join("..", "testData", "icon.icns"), nil, "ico")
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(file, ".ico"))
}
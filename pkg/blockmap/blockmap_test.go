package blockmap

import (
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"crypto/sha512"
	"encoding/base64"
)

func TestAppend(t *testing.T) {
	file, err := ioutil.TempFile("", "append")
	if err != nil {
		t.Error(err)
	}

	log.Print()

	file.WriteString(strings.Repeat("hello world. ", 1024))
	Close(file)

	inputInfo, err := BuildBlockMap(file.Name(), DefaultChunkerConfiguration, DEFLATE, "")

	fileData, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Error(err)
	}

	hash := sha512.New()
	hash.Write(fileData)
	assert.Equal(t, base64.StdEncoding.EncodeToString(hash.Sum(nil)), inputInfo.Sha512)
	assert.Equal(t, len(fileData), inputInfo.Size)

	serializedInputInfo, err := json.Marshal(inputInfo)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "{\"size\":13423,\"sha512\":\"zPFW3WAFUKFvAfBdNXHDIuZekSW/qf33lf5OgKXBKg9oOobwVH9X/DRHExC9087Cxkp3nqFrwtreWZHLso3D6g==\",\"blockMapSize\":107}", string(serializedInputInfo))
}

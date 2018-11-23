package blockmap

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/aclements/go-rabin/rabin"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/json-iterator/go"
	"github.com/minio/blake2b-simd"
)

type BlockMap struct {
	Version string         `json:"version"`
	Files   []BlockMapFile `json:"files"`
}

type BlockMapFile struct {
	Name   string `json:"name"`
	Offset uint64 `json:"offset"`

	Checksums []string `json:"checksums"`
	Sizes     []int    `json:"sizes"`
}

type InputFileInfo struct {
	Size   int    `json:"size"`
	Sha512 string `json:"sha512"`

	BlockMapSize *int `json:"blockMapSize,omitempty"`

	hash *hash.Hash
}

type ChunkerConfiguration struct {
	Window int
	Avg    int
	Min    int
	Max    int
}

type CompressionFormat int

const (
	GZIP    = 0
	DEFLATE = 1
)

var DefaultChunkerConfiguration = ChunkerConfiguration{
	Window: 64,
	Avg:    16 * 1024,
	Min:    8 * 1024,
	Max:    32 * 1024,
}

func BuildBlockMap(inFile string, chunkerConfiguration ChunkerConfiguration, compressionFormat CompressionFormat, outFile string) (*InputFileInfo, error) {
	checksums, sizes, inputInfo, err := computeBlocks(inFile, chunkerConfiguration)
	if err != nil {
		return nil, err
	}

	blockMap := BlockMap{
		Version: "2",
		Files: []BlockMapFile{
			{
				Name:      "file",
				Offset:    0,
				Checksums: *checksums,
				Sizes:     *sizes,
			},
		},
	}

	serializedBlockMap, err := jsoniter.ConfigFastest.Marshal(&blockMap)
	if err != nil {
		return nil, err
	}

	if len(outFile) == 0 {
		archiveSize, err := appendResult(serializedBlockMap, inFile, compressionFormat, inputInfo.hash)
		if err != nil {
			return nil, err
		}

		inputInfo.Size += archiveSize + 4
		inputInfo.BlockMapSize = &archiveSize
	} else {
		err = writeResult(serializedBlockMap, outFile, compressionFormat)
		if err != nil {
			return nil, err
		}
	}

	inputInfo.Sha512 = base64.StdEncoding.EncodeToString((*inputInfo.hash).Sum(nil))
	return inputInfo, nil
}

func appendResult(data []byte, inFile string, compressionFormat CompressionFormat, hash *hash.Hash) (int, error) {
	archiveBuffer := new(bytes.Buffer)
	err := archiveData(data, compressionFormat, archiveBuffer)
	if err != nil {
		return -1, errors.WithStack(err)
	}

	outFileDescriptor, err := os.OpenFile(inFile, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return -1, errors.WithStack(err)
	}

	defer util.Close(outFileDescriptor)

	archiveSize := archiveBuffer.Len()
	_, err = io.Copy(outFileDescriptor, io.TeeReader(archiveBuffer, *hash))
	if err != nil {
		return -1, errors.WithStack(err)
	}

	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, uint32(archiveSize))
	_, err = outFileDescriptor.Write(sizeBytes)
	if err != nil {
		return -1, errors.WithStack(err)
	}

	_, err = (*hash).Write(sizeBytes)
	if err != nil {
		return -1, errors.WithStack(err)
	}

	return archiveSize, nil
}

func writeResult(data []byte, outFile string, compressionFormat CompressionFormat) error {
	if outFile == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}

	outFileDescriptor, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer util.Close(outFileDescriptor)

	return archiveData(data, compressionFormat, outFileDescriptor)
}

func archiveData(data []byte, compressionFormat CompressionFormat, destinationWriter io.Writer) error {
	var archiveWriter io.WriteCloser
	var err error
	if compressionFormat == DEFLATE {
		archiveWriter, err = flate.NewWriter(destinationWriter, flate.BestCompression)
	} else {
		archiveWriter, err = gzip.NewWriterLevel(destinationWriter, gzip.BestCompression)
	}
	if err != nil {
		return err
	}

	defer util.Close(archiveWriter)

	_, err = archiveWriter.Write(data)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func computeBlocks(inFile string, configuration ChunkerConfiguration) (*[]string, *[]int, *InputFileInfo, error) {
	inputFileDescriptor, err := os.Open(inFile)
	if err != nil {
		return nil, nil, nil, err
	}
	defer util.Close(inputFileDescriptor)

	var checksums []string
	var sizes []int

	chunkHash, err := blake2b.New(&blake2b.Config{Size: 18})
	if err != nil {
		return nil, nil, nil, err
	}

	inputHash := sha512.New()

	copyBuffer := new(bytes.Buffer)
	r := io.TeeReader(inputFileDescriptor, copyBuffer)
	c := rabin.NewChunker(rabin.NewTable(rabin.Poly64, configuration.Window), r, configuration.Min, configuration.Avg, configuration.Max)
	for i := 0; ; i++ {
		copyLength, err := c.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, nil, err
		}

		_, err = io.Copy(chunkHash, io.TeeReader(io.LimitReader(copyBuffer, int64(copyLength)), inputHash))
		if err != nil {
			return nil, nil, nil, errors.New("error writing hash")
		}

		checksums = append(checksums, base64.StdEncoding.EncodeToString(chunkHash.Sum(nil)))
		sizes = append(sizes, copyLength)

		chunkHash.Reset()
	}

	inputFileStat, err := inputFileDescriptor.Stat()
	if err != nil {
		return nil, nil, nil, err
	}

	sum := 0
	for _, s := range sizes {
		sum += s
	}

	fileSize := int(inputFileStat.Size())
	if sum != fileSize {
		return nil, nil, nil, fmt.Errorf("expected size sum: %d. Actual: %d", fileSize, sum)
	}

	return &checksums, &sizes, &InputFileInfo{
		Size: fileSize,
		hash: &inputHash,
	}, nil
}

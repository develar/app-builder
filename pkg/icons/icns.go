package icons

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"image/png"
	"io"
	"io/ioutil"

	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/disintegration/imaging"
)

//noinspection GoSnakeCaseUsage
const (
	ICNS_256        = "ic08"
	ICNS_256_RETINA = "ic13"
	ICNS_512        = "ic09"
	ICNS_512_RETINA = "ic14"
	ICNS_1024       = "ic10"
)

var (
	icnsHeader = []byte{0x69, 0x63, 0x6e, 0x73}

	icnsExpectedSizes = []int{16, 32, 64, 128, 256, 512, 1024}

	// all icon sizes mapped to their respective possible OSTypes, this includes old OSTypes such as ic08 recognized on 10.5
	sizeToType = map[int][]string{
		16:   {"icp4"},
		32:   {"icp5", "ic11"},
		64:   {"icp6", "ic12"},
		128:  {"ic07"},
		256:  {ICNS_256, ICNS_256_RETINA},
		512:  {ICNS_512, ICNS_512_RETINA},
		1024: {ICNS_1024},
	}
)

func ConvertToIcns(inputInfo InputFileInfo) (string, error) {
	// create a new buffer to hold the series of icons generated via resizing
	icns := new(bytes.Buffer)

	for _, size := range icnsExpectedSizes {
		if size > inputInfo.MaxIconSize {
			// do not upscale
			continue
		}

		var imageData []byte
		var err error
		existingFile, exists := inputInfo.SizeToPath[size]
		if exists {
			imageData, err = ioutil.ReadFile(existingFile)
			if err != nil {
				return "", errors.WithStack(err)
			}
		} else {
			maxImage, err := inputInfo.GetMaxImage()
			if err != nil {
				return "", errors.WithStack(err)
			}

			imageBuffer := new(bytes.Buffer)
			err = png.Encode(imageBuffer, imaging.Resize(maxImage, size, size, imaging.Lanczos))
			if err != nil {
				return "", errors.WithStack(err)
			}

			imageData = imageBuffer.Bytes()
		}

		// each icon type is prefixed with a 4-byte OSType marker and a 4-byte size header (which includes the ostype/size header).
		// add the size of the total icon to lengthBytes in big-endian format.
		lengthBytes := make([]byte, 4, 4)
		binary.BigEndian.PutUint32(lengthBytes, uint32(len(imageData)+8))

		// iterate through every OSType and append the icon to icns
		for _, ostype := range sizeToType[size] {
			_, err = icns.Write([]byte(ostype))
			if err != nil {
				return "", errors.WithStack(err)
			}
			_, err = icns.Write(lengthBytes)
			if err != nil {
				return "", errors.WithStack(err)
			}
			_, err = icns.Write(imageData)
			if err != nil {
				return "", errors.WithStack(err)
			}
		}
	}

	// each ICNS file is prefixed with a 4 byte header and 4 bytes marking the length of the file, MSB first
	lengthBytes := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(icns.Len()+8))

	outFile, err := util.TempFile("", ".icns")
	if err != nil {
		return "", errors.WithStack(err)
	}

	defer outFile.Close()

	outFile.Write(icnsHeader)
	outFile.Write(lengthBytes)
	io.Copy(outFile, icns)

	return outFile.Name(), nil
}

func IsIcns(reader *bufio.Reader) (bool, error) {
	data, err := reader.Peek(4)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return data[0] == 0x69 && data[1] == 0x63 && data[2] == 0x6e && data[3] == 0x73, nil
}

type SubImage struct {
	Offset int
	Length int
}

func ReadIcns(reader *bufio.Reader) (map[string]SubImage, error) {
	type IcnsIconEntry struct {
		Type   [4]byte
		Length uint32
	}

	_, err := reader.Discard(8)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	typeToImage := make(map[string]SubImage)
	offset := 8
	for {
		icon := IcnsIconEntry{}
		err = binary.Read(reader, binary.BigEndian, &icon)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		offset += 8
		imageDataLength := int(icon.Length) - 8

		osType := string(icon.Type[:])
		if osType != "info" && osType != "TOC" && osType != "icnV" && osType != "name" {
			typeToImage[osType] = SubImage{
				Offset: offset,
				Length: imageDataLength,
			}
		}

		offset += imageDataLength

		_, err = reader.Discard(imageDataLength)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return typeToImage, nil
}


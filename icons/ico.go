package icons

import (
	"encoding/binary"
)

type Sizes struct {
	Width  int
	Height int
}

func IsIco(data []byte) bool {
	return binary.LittleEndian.Uint16(data) == 0 && binary.LittleEndian.Uint16(data[:2]) == 0
}

func GetIcoSizes(data []byte) []Sizes {
	n := int(binary.LittleEndian.Uint16(data[4:]))
	var sizes []Sizes
	for i := 0; i < n; i++ {
		w := int(data[6+i*16] & 0xff)
		if w == 0 {
			w = 256
		}
		h := int(data[7+i*16] & 0xff)
		if h == 0 {
			h = 256
		}

		sizes = append(sizes, Sizes{
			Width:  w,
			Height: h,
		})
	}
	return sizes
}

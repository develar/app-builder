// Code generated for package snap by go-bindata DO NOT EDIT. (@generated)
// sources:
// pkg/package-format/snap/desktop-scripts/desktop-common.sh
// pkg/package-format/snap/desktop-scripts/desktop-gnome-specific.sh
// pkg/package-format/snap/desktop-scripts/desktop-init.sh
package snap

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _desktopScriptsDesktopCommonSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x3b\x69\x73\xdb\xb8\x92\xdf\xf9\x2b\x3a\x94\x2a\xb1\x93\x48\xb4\x9d\x49\x5e\xd6\x63\x65\x9f\x62\x2b\x8e\x2a\xf2\xb1\x96\x9c\xc9\xd4\xcb\xac\x0a\x22\x21\x09\x2b\x0a\xe0\x03\x48\xdb\x1a\xc7\xff\x7d\xab\x01\xf0\x12\xe9\x2b\xc7\xd6\xba\x2a\xb1\x45\x36\xfa\x42\x77\xa3\x0f\xa8\xf1\xc4\x9b\x30\xee\x4d\x88\x9a\x43\x8b\x3a\x4e\xe3\x71\x3f\x4e\x03\x06\x24\xe1\xfe\x9c\x4a\xf0\xc5\x72\x29\x38\xd0\xab\x48\xc8\x58\xc1\x54\x48\x20\x7c\x05\x01\x55\x8b\x58\x44\x40\xa2\x08\x1a\x8f\xc6\xef\x4c\x13\xee\xc7\x4c\x70\x88\x24\x8d\x28\x0f\xc6\x01\x93\x1b\x9b\x70\xed\x00\x84\xc2\x27\x21\x5c\x10\xd9\x71\x9b\xdb\x6e\xf6\x20\x60\xf8\x60\x07\x1f\xb0\x29\xfc\x0b\x5a\x01\xb8\xcd\x80\x49\x17\xfe\xfa\x1d\xe2\x39\xe5\x0e\x00\x00\xbd\x20\x21\xb8\x86\x59\x68\x22\x92\xaf\xee\x57\x04\xfb\xda\xbc\xc6\x8f\xbb\x2f\x76\xbf\x36\xf1\x8f\x9b\xaf\x2e\xa2\x9a\x32\xe7\xa6\xc0\x0d\x89\x7e\x3d\x33\x96\x0f\xc3\xc6\xee\x8d\xe1\xae\x8e\x19\x9f\xf0\xb1\x88\x28\x1f\x4f\x59\x48\x2d\x3f\x73\x4a\x02\x68\xf9\x5b\x80\xec\xc0\xd3\x77\xe0\x05\xf4\xc2\xe3\x49\x18\x96\x96\x26\x51\x40\x62\x3a\xbe\x0a\x66\x28\x8b\x1a\x5f\x90\x30\xa1\xaa\x24\x93\x22\x17\x74\xcc\x38\x8b\x19\x09\xed\xfb\xce\x94\x84\x8a\x66\x10\x5f\x0e\x0e\xc7\x07\xfd\xb3\x61\xc7\x3d\x38\xd9\x3f\x3f\xea\x1d\x8f\x86\x70\xd0\x1b\x7e\x1a\x9d\x9c\xc2\xc1\xc9\x1f\xc7\x83\x93\xee\x01\x1c\x9d\x0f\xfb\xfb\x70\xda\xdf\x1f\x9d\x9f\xf5\x86\xf0\xb9\x7f\xd0\x3b\x19\xc2\xe9\xf9\xfb\x41\x7f\x7f\xf8\xb1\x7b\xd6\x83\x51\xef\xe8\x74\xd0\x1d\xf5\x86\x28\x62\xc2\x15\x8d\x35\xe6\xe1\x69\x6f\xbf\xdf\x1d\x68\x0a\xe3\xd3\xee\xe8\xe3\xd0\xc9\xb4\x39\x05\xb7\x79\x8d\x40\xfb\x27\xc7\x1f\xfa\x87\xe3\x8f\x27\x47\xbd\xdd\x56\x13\x7f\x79\x6d\x5f\xf0\x29\x9b\xdd\x78\x89\xa2\xb2\x85\xd2\xb5\xf1\xbf\xb2\xe6\x1b\xa0\xe6\x34\x0c\xfd\x39\xf5\x17\xa0\x44\x22\x7d\xda\xc9\x35\x85\x10\xe6\xe1\x77\xd0\x31\xfb\x94\xb1\xfa\x37\x18\x0c\x45\x71\x5e\x5c\xdd\x94\xb8\xa9\x53\x75\x2c\x13\x9a\xa1\x42\xa7\x0a\x80\x71\x68\xa6\x3a\xff\x1d\x02\xa1\xd7\x6a\xd3\xc3\xa7\xcd\xeb\xe0\x06\x5f\xb9\xf6\x71\x98\xd0\x8e\xdb\xbc\x7e\x82\xb6\xec\x3a\xfa\xa1\xe6\xc8\x6d\xd6\x90\x73\xa1\x03\x48\xb1\xc4\x16\x54\xf6\xe1\x45\x67\xc3\x45\xab\x74\x37\x2d\x80\x91\x91\x03\x3e\x0d\x13\xea\xae\xad\xaf\x62\x18\xf7\x8f\xfb\x23\xfc\xa0\x77\xd4\xe2\xc3\x95\x29\xc6\x29\x73\xec\x2f\xe7\x01\x04\xea\x0d\xa5\x82\x56\x23\x0d\x04\xa7\x25\x27\x60\x6a\xac\x92\x49\x44\xe2\xb9\x35\x7c\xe3\xb5\x1b\x92\x92\x10\x9f\x6a\x1f\xda\x44\x7d\x46\x44\x52\x1e\xaf\xbd\xdb\x31\xef\x50\xa3\xd7\x01\x93\x8d\x46\xd3\x80\x79\x37\x2e\x3c\xe9\x64\xfe\x0e\x4f\x9f\x82\xa4\x71\x22\x39\x6c\xc1\xb7\x6f\xe9\xdf\xdb\xc8\x4a\x2e\xdd\xf0\xb8\x7b\x3a\xb6\xce\x33\x3e\x3b\x3f\x1e\xf5\x8f\x7a\x45\x61\x1b\x40\x82\x00\x66\x94\x53\x49\x42\x40\x0e\x14\x70\x11\xe3\x53\x1a\xc0\x64\x05\x8a\x93\xc8\x97\x64\x1a\x43\x90\x50\x88\x05\xc8\x84\xc7\x6c\x49\xf5\x0b\x07\x0a\x91\x0b\x06\x07\xe3\x41\xff\xfd\x59\xf7\xec\x4f\xad\xae\x5b\xa8\x7b\x21\x9b\x78\xe5\x37\xdd\xb3\xfd\x8f\xe3\xd1\x59\xff\x74\xd0\x1b\xb9\xdf\x85\x33\x51\xf2\x71\x78\xef\x43\x36\x61\xdc\x75\xd0\x58\x1a\xf0\xe5\xd3\x7b\x30\x3e\xe9\xd8\x90\xfa\xe5\xd3\xfb\xd4\x73\xcf\x4e\x4e\x46\x9d\x3b\xf0\xa8\x39\x91\xd4\xfb\xb2\xbd\xed\x5d\x2d\x26\x2e\xa2\x3b\x64\x17\x14\xbe\x9c\x44\x94\xf7\x8f\x80\x80\x3f\x27\xdc\xd7\x8a\xc5\xe0\x17\x53\x13\x03\x29\x04\x24\x26\x6d\xa7\x01\xa3\x39\x53\xc0\x14\x48\xfa\xef\x84\x49\x1a\x68\x9f\x8d\xe9\x55\x0c\x8c\x47\x49\x8c\x0b\x2f\x85\x5c\xa0\x13\x0f\x0f\x06\x3b\x30\x23\x4b\xaa\xda\x19\xa7\x83\x93\xfd\xee\xa0\x77\xd0\x3f\x7b\x18\x93\x86\xb8\xe6\x73\x88\xf1\x72\x3f\x91\x4a\x48\xa5\x0d\x23\xc3\xb9\x7f\x7e\x36\x3c\x39\xd3\xfb\x71\x3f\x56\xe6\x0b\xae\x5c\xa7\x70\xdc\x96\x10\xd8\x3d\xf0\x50\x5c\x0c\x76\x29\xbc\xd3\x80\x23\xaa\x08\x0c\xd8\xc4\x9c\xfd\xa8\xb0\xc3\x01\xa8\x24\x42\x26\x9c\x5f\x61\x22\xde\x92\x2a\xe2\xfe\x3a\xd4\x2d\x3a\x0b\xb5\x64\x23\x1a\x86\x10\xb2\xc9\xe1\x00\x2e\xe7\x54\xea\xdd\x9f\x32\x1e\xa0\x53\x42\x20\xd9\x05\x95\x2a\xd5\xf6\xa0\xff\xfe\x70\x30\x3e\x38\xeb\x7f\xee\xd9\x30\x74\x97\xce\xef\xe1\x22\x90\xec\x1e\xf9\xaa\xe4\xdc\x02\x27\x9f\xbb\x65\x4e\xee\x61\x04\x69\x23\x4d\x94\xf9\x0f\x21\x17\x44\x8a\x84\xeb\xf3\x06\x83\x87\x31\xe5\x48\x8a\x48\x32\x1a\x13\xb9\x02\xfe\x99\x05\x8c\xa4\x1a\x80\xa5\x48\x78\xac\x8a\x4a\x01\xc6\x9d\x06\x78\x17\xc4\x10\xd0\x58\xf4\x5f\xb3\x10\xe2\x39\x89\x81\x53\x1a\x28\x54\xe7\x84\x22\x99\x35\xf9\x90\x0d\x16\xcf\x05\xba\x0d\x42\x5b\x9b\x4a\x14\xe3\x33\x0c\x0e\x0a\x02\xa1\xa3\x9f\xf6\xa8\x4b\x16\xcf\x35\xf1\x32\x5b\xe8\x94\x67\x74\xda\xde\x85\x79\x1c\x47\x6a\xd7\xf3\x26\xc9\x4c\xb5\x43\x9d\xb7\x46\x24\x68\x73\x1a\x6b\xce\xa2\x95\xf7\x62\x92\xcc\xbc\xed\xd7\x6f\xdf\x6e\xff\xc7\xce\x5d\x7a\xaf\x95\x08\xb5\x76\xce\x59\xbc\xfa\x87\x4d\x82\x61\xe3\x32\x57\x22\xea\x2e\x65\xa0\x4c\x1b\xd9\xf1\xb6\xdf\xbc\x7a\xfb\xdb\xd6\xeb\xcd\x5f\x63\xcd\x21\x9b\x24\xc8\x99\xb6\xe6\xd3\x24\x54\x94\x24\x01\x13\x96\xcf\x5f\x43\x33\xca\xc8\x68\xaa\xbd\xc3\x01\x5c\x50\x1e\x08\x09\x98\xa6\x2a\x10\x1c\x66\xe1\x05\x0f\x80\x72\x32\x09\x69\x00\x6a\xa5\x62\xba\x54\x8e\xce\x92\x6b\x35\x7c\xc1\x03\x8f\xce\xc2\xb1\xc1\xd3\x0e\xcc\xa1\xfa\x55\x1f\xed\x05\x19\xc6\xe3\xde\xe1\x60\xfc\xb9\x77\x7c\x70\x72\x96\x49\x83\x59\xc1\xc3\xb0\x66\x1e\x7f\x38\x8c\x25\x25\x4b\x2a\xd7\xbc\x9e\xc5\x0a\xa2\x30\x99\x31\x9e\x79\xfd\xe1\x70\x34\x3e\x1d\x9c\x1f\xf6\x8f\x8b\x2e\xff\x10\x35\xcd\x94\x25\xd2\xda\x6e\x6f\xb9\x35\xf8\x86\x7f\x0e\x47\xbd\xa3\x1f\x8d\x24\x6b\x64\x1a\x30\x53\xb1\x15\x02\x94\x4f\x38\xa7\x12\x02\x41\x15\x7f\x86\x07\x95\x8a\x49\x18\xa2\x43\xa2\x3b\xf9\x42\x4a\xea\xc7\xfa\x4c\xc9\x9d\x68\xc6\xe2\x79\x32\x69\xfb\x62\xe9\x25\x93\x84\xc7\x89\x97\x65\x1e\x2d\x5b\xf3\xb5\xe6\x34\x8c\xa8\x54\x1e\x53\x2a\xa1\xca\xfb\xed\x55\x9d\x78\xfb\xdd\xe3\xe3\xde\x9d\xc7\xdd\x43\x45\xdb\x6e\x6f\x95\xe5\xc4\x4f\x2d\x23\x64\xcb\x0a\xa9\x6d\xf1\xcb\xc1\x21\xec\x9b\x0c\xe1\xee\xb4\x0b\xcd\xab\x74\x12\xe6\x25\x80\xb6\xa7\x5b\x98\xa6\xb1\xef\x5d\x05\xb3\xb5\x53\xb4\x76\x6d\x0e\xeb\x34\xe0\x80\x4e\x19\x37\x99\x9a\x7a\x06\xe2\x92\xeb\xac\x02\xd3\xd1\xef\xe0\xf3\xa0\x3b\xea\xde\xc9\x65\x76\xe6\x57\xf9\x5c\x5f\xfb\x28\xd8\x87\xc2\xa5\x29\xc4\xfd\xa0\xe3\xf3\x61\xef\x4c\x3f\xcd\x33\x9d\x14\x0c\x6b\xb0\x34\x1b\x0b\xcb\x79\x4f\x11\xa4\xb3\x8e\xc9\x6b\xeb\x15\x29\xb7\xcb\x05\x12\x6f\x45\xe0\x36\x4b\xeb\xdc\xb5\xb3\x10\xc3\xf8\xe1\x80\x4d\x60\x0f\x76\xda\xaf\x5f\xb5\x77\xf4\xf9\xa3\x28\x91\xfe\x1c\x4f\x25\x7c\xaf\xfc\x39\x5d\x12\x95\xd7\x69\x29\xb2\x5d\xa7\x01\x50\x3c\x86\xfe\x66\x61\x48\xda\x33\x2e\x96\xb4\x2d\xe4\xcc\x53\x73\x71\x39\x9e\x24\xb3\xb6\x3f\x63\xff\xc9\x82\xce\x3f\x7e\xdb\x7e\xf5\xea\xf5\x9d\xfa\xa9\x72\x8b\xea\xf1\x89\x3f\xa7\x30\x15\x61\x40\xe5\xed\xda\xd9\xef\xee\x7f\xec\x55\xd5\xb3\x7f\x72\x74\x74\x72\xec\xb5\x35\x12\x57\x17\x27\xb6\x73\xb1\xae\x42\x03\x81\xd6\xf7\x04\x5a\xd4\x72\x93\x63\x75\xe1\xaf\x62\xe5\x82\xc1\xc4\x2c\x41\x93\xa6\x7e\x2c\xe4\x0a\x12\x45\x03\x9b\x00\xa8\x58\x60\xc2\x9c\x70\x64\x7a\x8d\xd6\x4b\x58\xb2\x99\xc4\x7c\x9b\xc5\x0e\xc0\xf2\xe2\x76\x6e\x6a\x44\xd1\xa5\xc1\xda\x16\x17\xd8\xcc\xb4\xa6\x63\xc2\xfd\x6a\xcb\x9b\x00\x35\x66\x65\x90\x54\x2c\xaa\xb0\x48\xd3\xdb\x97\x14\xa5\xd1\xef\xac\x47\xe2\x96\x62\xa1\x8b\x16\x45\xaf\x98\x8a\x15\x6c\x18\xcd\x48\xba\x14\x17\x34\xc0\x73\x88\xc3\xe0\x74\x17\x1a\xdb\x6f\x5e\xbf\x79\xf5\xdb\x16\x56\x1a\x53\x76\x45\x83\xcd\x42\x0d\xb9\x86\xb2\x58\x3e\xae\x31\x55\x84\x72\x00\xfc\xf9\x52\x04\xf0\x8f\xad\xad\xba\xd7\xa6\xb8\xea\x71\x95\xe0\x51\x38\xa7\xba\xa3\x87\xe7\xa1\xca\x4a\x20\x0c\x5f\x0c\x4b\x6a\x05\x1b\xb6\xfe\x49\x5f\xaa\x16\x9e\x28\x69\xa2\xa7\xcf\x17\x64\xba\x98\x79\x9c\xec\x3f\x24\xe3\x28\x54\x3c\xfd\x29\x10\xbe\x7a\x09\x0b\x4a\x23\x88\x25\xf1\x17\x20\xa6\xf6\xac\xc6\x10\x1f\x30\xa9\xe0\x12\x3f\x29\x01\x97\x14\x7c\xc2\x21\x12\x31\xe5\x31\x23\x61\xb8\xca\x2c\xca\x1c\x72\x1c\x5f\x40\x48\x62\x2a\x9d\xfa\xb6\x98\x35\x93\x24\x82\xbc\xd5\xf3\x1c\x84\xc4\x12\x1b\xae\x82\x59\x2b\x7b\xdc\x32\x08\xf4\x66\x52\x1a\xd0\xc0\xd1\x89\xae\x46\x68\x5e\xd9\x16\x5a\xfe\x58\xd2\x50\x90\xa0\xf2\x38\x64\x7c\x91\xf6\xdb\x1c\xdb\xb7\x31\xae\xa5\x5b\x0b\x6b\xd1\xd1\x9c\x04\x4f\x0a\x1d\x8d\x1a\xeb\x4b\x31\x64\x46\x81\x31\x6b\x8a\xd1\xaa\xdc\xc1\x2a\x7c\x34\x4a\xcf\x7a\x4d\x79\xf3\xad\xd8\x01\xf3\x9a\xd3\xf5\xb6\x8c\x76\xd3\x0a\x4c\x8d\x43\x18\xf0\x8a\x36\x6c\xf3\xab\xd8\xb8\x41\x33\x64\xd3\x72\xcf\x13\xdc\xe6\x59\xaf\x3b\x28\x11\x5a\xef\xfa\x3d\x7d\xfa\x98\x35\xd6\xca\x32\x61\xac\xe6\x4b\x96\xb9\x7f\x72\x74\x7a\x72\xdc\x3b\x1e\x0d\xc7\xc7\xbd\xde\xc1\xf8\xfc\xf4\xa0\x3b\xea\xb9\xd0\x01\x17\xf9\xc6\xcd\xf8\xf6\x0d\xfe\x05\xcd\x75\xb1\xea\x7a\x6c\x18\x05\x5d\xef\xbf\x1b\xde\x13\xd5\xf8\xaa\x55\xd6\x68\x5e\xa7\x0c\xde\x34\x66\xee\x43\x64\x7c\x57\x55\x6d\x4d\x57\x12\xc0\x8f\xa0\x45\x1e\xa4\x80\xdb\xf6\xea\xd1\x26\x03\xb0\x0c\x5e\xab\x64\x09\x7b\x75\x74\xd1\x2a\xea\x98\x6f\x4e\xdb\x66\x99\xa1\xaa\x0d\xe0\x2e\x43\x99\x32\x87\x9a\xce\x74\xc5\xdf\x2c\xc4\xba\x67\xe9\xc7\xd6\xa6\x1e\xb4\x55\xb7\x84\x86\x1a\x9e\x8c\xd3\x9a\xb0\xb9\xaf\x1b\xcc\x6c\x9a\x86\x21\xe2\xc7\x89\x8e\x41\x92\x92\xa0\x18\x80\x14\x86\x30\x4a\xfc\x39\x06\x14\x9d\xf9\xa1\xaa\x37\x36\x18\x74\x60\xeb\x77\x60\xb0\x07\xcd\xeb\x46\x7d\xbb\xf3\x5f\xff\xfc\xeb\xe6\x77\x60\x2f\x5e\x6c\x6e\x5a\xbd\xb3\x29\x3c\xa9\x98\x7d\xa5\x0d\x6d\x57\x37\xd9\x5f\x37\x6e\xc1\x24\x6f\x53\xe1\x2d\x4a\xc4\x17\x13\x49\xc9\xc2\xec\x83\xde\x2b\x13\xa4\x4d\x91\x8f\x11\xd9\x86\x45\xc2\x83\xfa\x78\x69\x4f\x3d\x4c\x9c\x90\xa7\x97\x3a\xb2\xb2\xb8\xb2\x39\x16\x3c\xdd\x1c\xed\xdc\x62\xb9\x44\xbc\xad\x8b\x7a\xd4\xef\xb2\xb6\x7e\x26\x62\x1d\xdc\xed\x3b\x6c\x37\xd2\x1c\xdb\x5a\x6e\xed\x05\x65\x07\xa8\x70\x6a\x00\x2b\x56\xf4\x23\x9b\x0a\x30\x59\xeb\x40\xdf\xb9\xa3\xd0\x6a\x49\x1a\x92\x98\x5d\xd0\x56\x2c\x3a\x36\xf8\x6f\xba\x85\x48\x5e\x0a\x86\xcd\xc9\x7a\x14\xcf\x27\x57\xb7\x00\x00\xc8\x25\x1e\xe1\x05\x80\x9d\xe2\xc4\xa9\xd4\xd1\xb7\xf8\x6c\xc2\x78\x2b\xc6\x90\x43\x4b\xad\xf3\x95\xc3\x57\xa6\x04\x36\x3a\x58\xf7\xd7\x76\x77\x49\x81\x48\x8a\x15\xad\x8f\xbb\x86\xe9\x39\xa7\x97\x66\xef\x5e\x82\x5f\x70\xc9\x39\xb9\xc8\x53\x00\x45\x30\xd5\x62\x1c\x44\x18\x98\x26\xaf\xce\x68\xd0\xb8\x30\x0b\x33\x09\xe8\x43\x36\xb0\x6e\xeb\x44\x18\x74\xea\x36\xac\x34\x0e\x31\x1b\x67\x5d\xed\xb2\x16\x7e\x1d\xce\xec\xd1\x00\xdc\xa6\x08\x03\x9b\x07\xd8\x4d\xe3\xf4\x32\x7b\x60\xec\x26\x40\x15\x58\xd0\x4d\x9b\x49\x64\x50\xd9\x0e\xc0\x46\x29\x8d\xd0\x78\xab\x19\xc7\xb7\x6f\x70\x37\x98\x49\xc1\xdd\xcd\xf5\xc4\xa0\x75\xc1\x2d\xb8\xf7\xdc\xd2\xf7\x60\xe7\x5d\xd9\x6a\x68\x98\x1b\xdf\x3d\x82\xe9\xd7\xff\x8f\x45\x29\xa7\x30\xda\x40\x03\x1a\x53\x3f\x86\x4b\xb2\x0a\xd1\xba\x14\x95\x17\x54\x82\x12\xfe\x82\xc6\x2f\x35\x0d\x50\x34\x06\xca\x2f\x98\x14\x7c\xa9\x8d\x53\x60\xde\x1d\xb2\xd4\x2a\x23\x49\xa7\x54\x3a\x8d\x14\xc9\x4b\x30\x98\x30\x4b\xf5\xc5\x32\x22\x31\xa8\xd5\x52\xef\xf7\x46\xc2\x63\x16\xa2\xb9\x27\x4a\xff\x93\xb6\x63\xdb\x86\x33\xba\xa4\xcb\x09\x95\x2f\x4d\x73\xa4\x5c\x90\x28\xf0\x64\xc2\x75\x6e\xe0\xed\x25\x2c\x78\xa7\x3b\x3d\x6d\xad\x17\xe4\x27\x14\x62\x91\xf6\x8a\xcc\xe4\x2b\x2f\xef\x9c\x86\x99\x7f\xcc\xa9\x15\xab\x0d\x1f\x84\x44\xc1\x09\x0b\x15\xd6\xc3\x69\x35\x3c\x15\x32\x59\xb6\xb3\x1e\x52\x9b\x09\x2f\xf6\xac\x50\xad\x00\xf3\x84\x16\xfe\x85\x31\xdb\x4e\xb4\xf4\xf8\x61\xfb\xed\x1b\x6f\x7b\xcb\x69\x40\xb7\xa8\x15\xdd\x2d\x0e\x04\xba\xbe\x1d\x40\x64\x4a\xe6\x3a\x12\x62\xf6\x4f\x56\xa6\x62\xa1\x70\xd0\x1f\x76\xdf\x0f\x7a\xe3\x3f\xba\x7f\x0e\xba\xc7\x07\x4e\x43\xd7\x5d\x84\xaf\x80\x0b\xde\xa2\xcb\x28\x5e\x99\x41\xea\x26\x56\x30\x6a\xc1\x22\x43\x21\x14\x33\xe6\x03\xd6\x13\x92\x86\xab\xb6\x53\x4a\x0f\x2d\xb2\x71\xf7\x73\xb7\x3f\x40\xf4\x36\x21\xb0\xe5\x74\x6d\xa1\xf6\xf4\x29\xb4\xfe\x06\xb7\xb9\xc6\x50\xa9\x8e\x06\xb8\x0c\x98\x8a\x42\xb2\xea\xb8\xa9\x7e\xb6\xdc\xb5\x21\x69\x4a\xfc\xa0\x3f\x3c\x1d\x74\xff\xac\x86\xd7\x1c\x47\x05\xd6\x29\x04\x56\x4b\x60\x8c\x9b\x87\x5e\xd1\xa9\x30\xed\xb5\xdb\x5e\x33\xc5\xe6\x96\x17\x71\x12\xdd\xb2\x68\x6d\x85\x61\x7c\x08\x6e\x73\x9d\x60\x95\xf3\x06\x42\xcb\x84\x73\x8c\xe7\xa6\x57\x90\xd9\x3e\x1a\xb6\x0e\xcf\x16\xd4\x16\xed\x99\x84\xbd\xf7\xe7\x87\x9d\xed\xec\xfd\x3d\xfb\x95\xa5\x35\x86\xac\x6f\x8e\x7e\x93\xac\x95\x3c\x0b\x4d\x9c\x8b\xcb\x0c\xb6\x78\xc0\xad\x2b\xa3\x2a\x50\xf1\xb4\xab\x8a\x5f\x83\x21\x5b\x99\x9f\x7e\x26\xa4\x1c\x91\x05\x35\xed\xfd\xae\x6e\xef\x1b\x97\x03\x72\x41\x58\x48\x26\xa1\xae\xbc\x59\x60\x44\x40\x6c\x2d\x15\x51\x9f\x4d\x99\x5f\x69\x43\x3c\xac\x9b\x00\xa0\x9b\xfc\xc6\xa5\x3a\xae\xfe\xe0\x99\x4f\x6e\xfe\x5a\xcf\x00\xee\xb3\xa0\x02\xa2\x75\x93\xa8\x41\x52\x55\xa2\xdd\xe9\xd3\xf3\xc1\xb0\x37\x1e\xf6\xce\x3e\xf7\xce\x3a\x6e\xc2\xd9\xd5\x6e\xf3\xba\x06\xc1\x8d\x5b\xd6\xdc\x61\x1f\x24\x8d\x84\x62\x3a\x68\x3d\xae\xd9\x7a\xd8\x1f\x8f\xfe\x3c\xed\x0d\xfa\xef\x7f\x74\x64\x32\x63\x39\x13\xa6\x4f\xff\x4b\x39\xa9\x92\xbb\x1f\xd9\xf7\x89\xf1\x08\xbc\x3f\xb6\xf8\x7f\x54\x0d\x02\xa7\x01\x9f\x28\x8d\x00\xeb\x2e\x29\xc9\x0a\x4b\xac\xb4\xb1\xae\x5e\x6a\xef\x0d\x85\x88\x30\x9e\xc4\x73\x29\x92\x99\x1e\x23\x2e\x9d\xfe\x87\x61\xe7\xd9\xee\x33\x53\xa3\xb5\x24\xd6\xcc\xb8\xcc\x54\x05\x06\xd3\xde\xde\x5e\xb1\x07\x8b\xa9\x99\xa6\xf7\x41\xf0\xd8\x8e\x18\xf4\x81\x8c\xf8\x68\x36\x2d\xfa\x70\x72\x3c\xb2\x65\xee\x5d\x63\x1d\x1a\xfb\xde\x54\xf0\x58\xb9\x35\x0b\x3f\xf4\x07\xbd\x7b\x17\x9a\xff\x75\x95\xed\x16\x2e\xbc\x2c\xc9\x82\x8e\xf1\x3c\x1f\xe3\x7b\xdb\xf5\xbc\x76\x00\xa8\x3f\x17\xe0\xee\xe5\x4f\xdf\x95\x6f\xae\xe5\x85\x7b\xa1\x79\x6e\x79\x2c\xdf\x68\xd3\x88\x00\xf6\x02\x26\xdf\xdd\xb5\x6c\xcf\x43\x08\x7b\x59\xaa\x9e\xd4\x23\xd0\xd7\xa1\xac\xcb\xd4\xd7\xb6\xb1\x2e\x51\xcf\x39\xb9\x5e\x87\xc6\xac\xbb\x4e\xe8\x75\xbe\xee\x58\x98\xf3\x58\x4c\x0c\x2d\x82\x67\x00\x7b\x8c\xfb\x61\x12\x50\x60\x33\x2e\x24\x1d\x2f\x99\x52\x8c\xcf\x3a\xee\x8a\x2a\xf7\x1d\xee\x4d\x3b\xd8\xf3\x2c\xd0\xbb\x67\xba\xe2\xf9\x83\xea\x4a\x1b\x53\x94\x74\x75\x3c\x67\x0a\x93\x1c\x92\x84\x76\x26\x80\x4e\x34\x65\x52\xe9\x4c\x52\x27\x31\xf8\x98\xf1\x99\x46\x71\x29\xe4\x42\xed\xea\x49\xba\x48\x62\x60\x31\x7a\x47\x66\x21\x97\x2c\x0c\x21\x96\x2b\x7d\x87\x45\xb2\x58\x4f\x44\xf1\x2c\xc1\x52\xd4\x24\x94\x73\xb1\x34\xe5\x57\x46\x0c\xad\x7f\x42\x61\x12\xe2\x59\xa4\x6f\x28\x75\xa3\xa8\x2b\x97\x42\xb6\x8b\xf2\x66\xf0\x98\xd0\xb2\xab\x8e\x7b\x15\xcc\xdc\x77\x39\xf1\x3d\x2f\x85\xd0\xd2\xd6\xd9\x89\x06\xf0\xf2\x25\xb7\x98\x4c\x86\xe7\x8e\xa5\x05\x6a\x99\x1d\x59\xdf\xf0\x4a\xce\x71\x93\x35\x65\x1f\xd9\x1a\xcc\x38\x93\x4b\x68\xc9\x69\x65\x90\xe3\x5d\xe7\x74\x5e\x6a\x93\x31\xff\xb7\x9e\xbf\x34\xb1\xe4\x65\xdb\xfc\xbe\x71\xb4\xbe\xf5\xb5\xa3\xc2\x5e\x4d\x25\x99\xe9\x82\x81\xa9\xbc\xdf\x8e\x59\x3a\xc9\x4a\x5a\xb3\xfd\x4c\xe9\xf5\xd6\x62\xf4\x06\xe9\xf4\x40\x0f\xe3\x8b\x18\xcd\xaf\x44\xea\xb5\xbb\xc0\xc9\x12\xb3\xe8\x78\x6e\xf6\x3b\x0f\x39\xc6\x38\xbd\xd7\x5b\xba\xb7\xa2\x63\x8f\x9e\xf8\xb7\xab\xe3\x87\x62\xab\xaf\xb0\x6f\x08\x58\x17\xa1\xea\x3a\x84\xf9\xeb\x72\xac\x4b\xe7\x4d\x46\x49\x59\xa6\x46\x24\xb5\x1d\x79\x33\xc9\x1b\x7d\x82\x57\xed\xed\xb7\x66\xb0\xa2\x0b\x18\x6d\x7e\xc0\xf4\x20\xdc\x9f\x13\x3e\xa3\x81\x46\x66\x40\x77\xb6\x74\xb9\x83\x3a\x8d\x45\x69\x16\x07\x97\x73\xe6\xcf\x61\x4e\x14\x6a\x8a\x53\x3f\xa6\x69\xd8\x47\xc9\x75\x82\x37\xad\xdc\x9e\x32\x00\x6e\x75\x8e\x97\xae\xe0\x77\x2c\x59\x9f\x3c\xd9\x37\x36\xa5\x79\x9f\xb0\x30\x80\x25\x5b\x52\x63\xde\x4e\xa3\x28\xf9\x2c\x5e\x68\xd7\xfc\x77\x0c\xcc\x17\xfc\xd7\x58\xb1\x87\xd4\xb3\x13\xe4\x89\x19\x19\xdc\x73\xff\x0c\x97\x78\x39\xd7\x65\x27\x66\xd3\x62\x9f\xcf\x34\xea\x5a\x4b\x5d\x0b\x92\x98\x4c\x88\xaa\xed\xf3\x81\x6d\x79\xb7\x22\x49\x75\x8d\xdd\xc1\xf2\x51\xc5\x64\x19\x29\x68\x05\x67\x0f\x63\xaa\x76\x93\xf0\xa7\x96\x8d\xdb\x34\x91\x1e\x4a\x59\xe2\xc9\x04\x2c\x45\x90\x84\xd4\x74\x98\xcc\x78\x74\xc3\xb8\x23\x66\x24\x33\x45\xe3\x98\xf1\x99\xb2\x60\x9b\xd9\x1d\x8a\xfe\xc9\xf8\xe8\xe4\xe0\x7c\xa0\x93\xe8\xce\xfa\x50\xd3\x9b\x31\xd1\xb2\x98\xdd\x1f\xdd\xde\x7c\xe0\xf3\xdd\xa9\x6d\xc8\x26\xad\x9d\xf6\x96\x66\xeb\xdf\x09\x95\xab\x94\xb7\xd2\x06\x67\x76\x54\x96\xce\x68\xae\x10\x3d\xea\x5e\x97\x5c\xec\xfb\xd2\x6f\xe1\x59\xae\x3c\xf7\x79\x5b\x89\x87\x10\xfa\x99\x88\x7f\x89\x7a\xeb\x48\x65\xe6\x37\xcc\x1a\x45\x0c\xcf\x87\xdc\xd8\xcc\x1d\x06\xe7\x70\x38\x1e\xee\x7f\xec\x1d\x75\x0b\x36\x96\x1b\x75\x46\xd5\xde\x78\x70\x0b\xdf\x49\x30\x28\xc7\xe9\x5d\x88\xeb\x92\x15\x6d\xdf\xb6\xeb\x45\x72\xd5\x4d\xaf\xbe\xfd\xce\xe4\x0e\xac\x7c\x63\x73\xd3\xbb\x3e\x51\xab\x4a\x57\xa8\xee\xb5\x18\x39\x12\x6f\x66\x81\xda\xa9\x2e\xeb\x1a\x17\xfa\x90\xce\xaf\x3e\x90\x10\x2b\x8b\x95\x3d\x33\xec\x16\x58\x3c\xd9\x2a\x5f\xf0\x98\xf1\xac\x11\x51\x6e\xa2\xeb\xe2\x70\x23\x54\xd0\xea\x96\xd8\x71\xbd\xe7\xed\xab\x65\x58\xec\x41\x6e\xde\xde\x62\xaf\xae\xab\xd5\xf5\x63\x88\x8b\x0b\x2a\x25\x0b\xe8\x77\x71\x90\x2d\xbe\x9b\x8d\x6c\x1a\xd8\x80\x13\x1e\xae\x52\x15\xe6\xd7\x6f\xcc\xb4\x4d\x44\x8c\x06\x40\xf8\x2a\xb6\x69\x6e\x2d\xfb\x65\x42\xb7\xa8\xe1\xa1\x82\x69\x03\xaf\xb7\x57\x96\x7e\x7d\xe7\x07\x23\xf2\xba\x7f\xfd\x8c\xd0\x81\x7f\x58\xbc\xad\xdc\xa3\xed\xb5\x0f\xdd\x3f\xca\xe3\x83\xce\xf6\x4d\x7e\xa4\x9c\x46\x96\x5f\xe9\xbb\xbf\x59\xca\x87\x8a\xc6\x8a\x40\x5f\x34\xc4\x4c\x34\xcf\x8c\xd0\x71\xd1\xfa\x9d\x03\xcc\xe6\x90\xb3\x91\x4d\x65\xf2\xab\xef\x95\x4b\x35\x5e\x60\xb2\xbb\x62\x3a\x51\xb3\x5e\xb7\xaa\xeb\x8f\xaf\xca\xe4\x59\x63\xac\x2c\x28\x85\x9d\x1a\x0a\x85\x73\xe0\x3e\xa4\xb7\xae\xcf\x62\xf0\x08\x53\x91\x09\x0b\x59\xbc\x7a\xf8\x75\x79\x2f\xce\x57\x3d\xe0\x0a\x7c\x11\xfc\xce\x2f\x5d\xfc\x2c\x4c\xf6\x6e\x3e\x66\x38\xc1\xa2\x15\xb1\xab\x49\x32\x85\x50\x90\xa0\x70\x55\xfe\xf0\xe0\xd3\xf8\xb4\xff\xe5\xfd\xf9\x87\xf4\x78\xd2\x4d\x8d\x4a\x22\x93\x21\x68\x59\x04\x36\xa1\xbd\x0d\x0d\x9a\xd0\xf7\xbb\x43\x4e\x0d\x9d\x62\xa7\xbd\xbd\xd5\xde\xf2\x52\xce\x7f\x42\x9a\xac\xcf\xb9\x5a\xc9\xdd\x9f\x93\x69\x95\x05\x28\x7c\xd4\x79\x41\xaa\xc3\xb2\xb9\xff\x1f\x90\x7b\x77\xa7\xdc\x99\x37\xf4\x7d\x5d\x99\xea\xaa\xcd\x6c\xf3\xaf\x29\x4d\xec\xf7\x56\x2a\x35\x69\x0d\xc4\x77\x26\x19\x76\xf0\xb5\xd4\xdf\x6c\xb8\x2d\xc9\xd0\x44\x3c\xf7\x79\x21\x35\xc9\x4d\x40\xaf\xf6\x18\x0f\xe8\x95\x29\xee\xb2\x71\xe7\x93\x12\x80\x2f\x78\x4b\xff\x59\x57\x33\xe1\x8f\x7e\x69\x93\x9d\x1a\x21\xbd\xe6\x06\x16\x2d\x58\xd2\xa7\x48\xd3\x0b\x02\xc5\x71\x46\x90\xbe\x1c\x57\xbe\xc1\x6a\x7e\x0a\xaa\xcc\xe1\x9c\xea\x94\xc3\x50\xd0\xb3\xd2\x5a\xb8\x07\x38\x81\x9a\x30\xee\xd9\xca\x4b\xcb\xaf\x25\xaf\x34\xe4\xd2\xf3\xf8\xf1\x78\x6e\x61\x2d\x1d\x48\x7f\x07\x73\xed\x59\xbc\xd8\xf9\x49\x0c\x5a\x5c\xb7\x30\x99\xe5\x69\x85\x3f\xcb\x49\x53\x61\x0e\x7d\x38\xfa\x64\xcd\xd4\xb4\xea\xe6\xe4\x82\x09\x89\xd5\x26\x9b\x32\x3d\x5a\x1e\xcd\x85\x32\xd7\x96\xd8\x32\x22\x7e\x9c\x37\x56\x80\xf2\x19\xe3\xd4\x5c\xe2\x9d\xac\xe0\xbf\x62\x20\x0a\x2e\x69\x18\x3a\xb3\x78\x31\x36\xc7\xa1\xea\x6c\xcc\xe2\x45\xeb\x15\x66\xd2\x36\x7d\x68\x33\xce\x20\x7d\x38\x11\x62\xb1\x24\x72\xa1\xf4\x13\x1d\x44\xe2\x05\xa6\x0f\xfe\x5c\x08\x45\x25\x02\x6f\x3a\xd9\x8d\x33\xb7\x79\x5d\xc0\x8d\x8e\xe7\x5a\xe7\x09\xa8\x8a\x3b\x75\x97\xc8\x8a\x2d\x88\x01\xb8\x4d\x04\xbc\xf5\xb8\xdf\x08\x98\xb4\x9e\xa0\xe1\x36\xef\x39\xec\xcd\x8d\x46\x0d\x5a\xbe\xf9\x64\xe7\x85\x6a\xb5\x9c\x88\x90\xf9\x60\x12\x24\x01\x6c\x92\xa8\x74\x34\xa7\x6f\x1a\xa0\x68\xfa\x61\x6c\x87\xe9\x49\xa4\xbf\x63\x62\x61\xf4\xd7\x65\x9c\x06\x6c\x28\xaa\x27\x8f\x4b\x7d\x4d\xac\xf1\x4a\xef\x57\xe3\x0d\x08\x7e\xe7\x17\x8b\x5e\xbf\xdd\xfa\xed\xcd\xab\x4d\xa7\xff\xfe\x7c\x38\x2e\x4f\x1d\xd6\x35\x85\x4c\x94\x2e\x2e\xaf\xaf\x71\x1d\xdb\x75\x5d\x7f\xe1\xe1\x4a\xfb\x1d\xd7\x34\xde\xd6\xc2\x38\x79\x5b\xab\xa2\x49\x24\x6f\x10\xd5\x51\x4e\x8f\xfb\xbb\x87\xb6\x08\x46\x7d\x70\x9b\xff\x74\xff\x37\x00\x00\xff\xff\x73\x58\x87\xb1\xdc\x40\x00\x00")

func desktopScriptsDesktopCommonShBytes() ([]byte, error) {
	return bindataRead(
		_desktopScriptsDesktopCommonSh,
		"desktop-scripts/desktop-common.sh",
	)
}

func desktopScriptsDesktopCommonSh() (*asset, error) {
	bytes, err := desktopScriptsDesktopCommonShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "desktop-scripts/desktop-common.sh", size: 16604, mode: os.FileMode(420), modTime: time.Unix(1710469316, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _desktopScriptsDesktopGnomeSpecificSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xbc\x94\xdd\x8f\xd2\x40\x14\xc5\xdf\xfb\x57\x5c\xa7\x3c\x19\xfb\x21\xfb\x68\x6a\x9c\x6d\x67\xa1\xa1\x5f\x5b\x06\x3f\x62\xcc\xa4\x94\x29\x1d\xb7\xb4\x38\x33\xcd\xc2\x7f\x6f\x20\x28\xeb\x02\x6e\x8c\xae\x7d\xbd\xbd\xf7\x9c\xf3\x3b\x69\xcd\x17\xce\x5c\xb4\xce\xbc\x50\x35\x58\xdc\x30\xcc\xdf\x3e\x86\x09\x23\x3a\x81\xa6\xe8\xdb\xb2\xe6\x12\xd4\x9a\x97\xa2\x12\x25\xac\x0b\xa9\xc1\x7c\x6a\xdb\x10\x15\x7c\x06\x34\x98\x26\x38\x63\x01\x99\x4e\x68\x9a\xb1\x0f\xf8\x53\x84\x93\x80\xe1\xf7\x38\x8c\xf0\x75\x44\x10\x78\x80\xb4\xec\x39\x82\x2f\x6f\x40\xd7\xbc\x35\x00\xf8\x66\xdd\x49\x0d\xa3\x60\xc2\xae\xb1\x3f\x21\x49\xe0\xa1\xfb\x62\xdb\x14\xed\x02\x1d\xc7\x7e\x34\xa3\x94\xe4\x67\x5f\x31\x21\xe8\xb8\x82\xb6\xd3\x50\xf7\x52\x83\xee\x0e\xf6\xb7\xa0\x6b\xa1\xa0\x50\x70\xcf\x9b\xe6\x15\x7c\xed\x95\x06\xd1\x42\x59\x28\x7e\x3c\x7d\x4b\xd9\x6d\x86\x59\x16\x61\x7a\x93\xe6\xb1\x77\xb8\x6c\xf1\x65\x63\x54\xc2\x30\x7e\x18\xa4\x13\x96\x61\x3a\xf6\x1e\xa5\xcc\x67\x09\x0d\x63\xe2\xf4\x4a\x3a\x8d\x98\x3b\xbf\x4e\x71\xee\x8f\x19\xcd\xc3\x2c\x22\xd4\x59\xea\x3b\xeb\xca\x76\x91\x61\x98\x20\xe6\xbd\x82\xa2\x5d\x40\x55\x0a\xbd\x01\xd1\x6a\xbe\x94\x85\x16\x5d\x6b\xec\x94\xc2\x98\xc5\x69\x30\x8b\x08\x0b\xc2\xdc\x1b\x7c\x0c\x46\xcc\xc7\xfe\x98\xb0\x71\x1a\x13\x47\xac\x56\xdd\xa2\x6f\xb8\x7a\x68\xee\xb8\x72\x13\x46\xc4\x1b\x9c\x9c\x39\xae\xd9\x65\x51\xd6\xfc\x6c\x69\x7e\x1a\x67\x69\x42\x12\x3a\x65\x09\x21\x01\x9b\x65\x01\xa6\xe7\x9b\x93\x2b\xb0\x64\x05\xe8\x54\x69\xd7\xca\xea\x6e\x21\x24\x58\xeb\x4b\xf3\xbd\xb8\xb5\x79\xac\xff\x07\x38\x1b\x31\xdf\x13\xb5\xdc\x3d\xd9\x6f\x3d\x97\x5b\xeb\x67\xc6\x3d\xe9\x07\x6e\x01\x9a\x16\x2c\x55\xfd\x85\xe0\xa1\x3f\xe7\xca\x76\x6d\xf7\x48\x13\x39\x2f\x6d\xd5\x5d\xca\x09\xcf\x9c\xf0\xed\x89\xf0\xae\xfe\xfd\xa7\xd3\xfc\x2b\xc6\x43\xdb\xbd\x60\x61\xf8\x2c\x90\x87\xb6\xeb\x0c\xed\xd7\xff\x95\xf2\x13\x19\x2f\x63\xae\xc4\xe1\x37\xc1\x4b\x40\x83\x77\xe8\x7b\x00\x00\x00\xff\xff\xd7\xce\x8a\x81\x79\x05\x00\x00")

func desktopScriptsDesktopGnomeSpecificShBytes() ([]byte, error) {
	return bindataRead(
		_desktopScriptsDesktopGnomeSpecificSh,
		"desktop-scripts/desktop-gnome-specific.sh",
	)
}

func desktopScriptsDesktopGnomeSpecificSh() (*asset, error) {
	bytes, err := desktopScriptsDesktopGnomeSpecificShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "desktop-scripts/desktop-gnome-specific.sh", size: 1401, mode: os.FileMode(420), modTime: time.Unix(1710469316, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _desktopScriptsDesktopInitSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x54\x7f\x6b\xdb\x30\x10\xfd\x5f\x9f\xe2\xaa\x98\xb6\x81\x25\x86\xd1\x85\xd1\xcd\x65\xa6\xf6\xda\xb0\x34\x09\x49\x3a\x06\xa5\x18\x45\x3e\xc7\x62\xb2\x15\x24\xb9\xcd\x20\x1f\x7e\xc8\x69\x7e\xf5\x67\xf4\x87\x41\xba\x77\xef\xde\x9d\x9e\xdc\x38\xf2\xa7\xa2\xf4\xa7\xcc\xe4\xd0\x42\x42\x1a\xcf\x17\x69\x40\x8f\x55\x25\xcf\x51\x83\x28\x85\x85\xc6\x2b\x18\x32\xee\x87\xc3\x24\x8a\xc7\xbf\x26\x83\x61\x72\x39\xb8\x19\x0e\xfa\x71\x7f\x32\x4e\xfa\x71\x1c\x25\xb7\xc3\x28\x9c\xc4\x01\xb5\xba\x42\x4a\x48\x03\x4c\x8e\x52\xf2\x1c\xf9\x5f\x30\xaa\xd2\x1c\x03\x3f\xc5\x07\xbf\xac\xa4\x24\x6d\xa0\x5e\xcd\x76\x3b\x8e\x47\x49\x14\x4e\x42\xbf\x2d\x99\xb1\x89\xc6\x07\x61\x84\x2a\x29\x7c\xbe\xd8\xc0\x61\xb9\x04\x47\x4b\x44\x06\x77\xeb\xcc\xb5\x8e\x5e\x38\x9e\x24\xa3\xf8\x77\x77\xdc\x1d\xf4\x29\x04\xeb\xf8\xf6\xe8\xfe\x1b\xd8\x1c\x4b\x02\x70\x50\x03\x19\x93\x06\x29\x41\x69\x90\x00\x20\xcf\x15\xd0\xb7\x2b\x06\xcf\xab\x5d\x7c\xd8\x1a\xc9\x84\x9b\xcf\x18\x2d\x78\xa3\x38\xec\x5d\x0f\x6e\x62\xb0\xca\x89\x84\xca\xa0\x36\xa0\x91\x49\xc8\x55\x81\x90\x0a\x8d\xdc\x2a\xfd\x8f\xac\x91\x01\xf5\x4e\x67\x68\xb1\xb4\x30\x67\xc6\x3c\xa6\xe0\xdd\x76\x23\x58\x02\xaf\x2c\xb4\x52\x38\x39\x3f\x81\x56\x06\x9d\x66\x7d\x09\xdd\x6c\x43\x0b\x39\x33\x50\xa8\x54\x64\x02\x53\x77\x2a\x74\x7d\xde\x4a\x85\x36\x60\xd0\x5a\x51\xce\xcc\x27\xc8\x94\xe6\x08\xac\x84\x6a\x9e\x32\xbb\x1a\xfa\x9d\xa3\xa4\xde\x9f\xe8\x2a\xb9\x1c\xf4\x7f\x76\xaf\x12\x27\xc5\xdf\xa4\xb7\xeb\x4f\x91\x7e\x31\x55\x41\xe1\xf8\xf8\x23\xb8\x54\x9c\x49\xdc\x24\xdc\x6f\xaf\x68\x55\x8d\x7a\xa7\xab\x18\x7c\x07\xba\x19\x92\xdf\xe6\xaa\xcc\xc4\xec\x59\x5d\xda\xa4\x70\xe4\xee\xfd\x94\x33\x7b\xb0\xca\x26\x85\xe5\x92\xc0\xd3\x3a\xb8\xe0\x4a\xf9\xe1\x25\xf7\x3b\x6d\xee\xf5\x7a\xa0\x21\x57\x2f\x0a\x20\x13\xb5\x73\x76\x1f\x41\x38\xba\xbc\xa6\x10\x04\x40\x59\x91\x76\xce\x76\xbd\xee\x42\x01\x5d\x7c\xed\x24\x9d\xb3\x96\x14\x65\xb5\x68\xcd\xca\xca\xf9\xfa\x0d\x02\x5d\xe4\xd9\x4b\x02\xa6\x8b\x6d\x36\xb2\xa9\xc8\xb3\x77\x39\x5e\x13\xc1\x98\xe6\xf9\x61\x2a\xe6\x73\xde\x39\x43\xf9\x92\x63\xae\x1e\x51\xbb\xa0\xc4\x7d\x9e\xfa\x95\xae\x30\x5b\xb2\x5d\x88\x9b\xd9\xde\x9c\x1d\x20\x99\x8c\xba\xc3\x5e\x3c\x09\xa8\x57\x17\x7f\x1a\x6b\x6d\x5b\x07\xf6\xa5\x98\xba\x7f\x66\x6a\x71\x61\x53\x55\x30\x51\xb6\x8d\xda\x95\x85\x8b\xb9\xd2\x16\x7a\x51\x32\x1c\xc5\xbd\x41\x18\x05\xd4\xdb\x6e\xce\xdf\x63\xa9\x25\x3d\xe5\xaf\xad\xb6\xde\x1f\xe0\x88\x57\xa1\xbb\x4d\x39\x72\xe4\x40\xbd\x1f\xf4\x7f\x00\x00\x00\xff\xff\xc0\xd6\xa2\xf0\xfa\x05\x00\x00")

func desktopScriptsDesktopInitShBytes() ([]byte, error) {
	return bindataRead(
		_desktopScriptsDesktopInitSh,
		"desktop-scripts/desktop-init.sh",
	)
}

func desktopScriptsDesktopInitSh() (*asset, error) {
	bytes, err := desktopScriptsDesktopInitShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "desktop-scripts/desktop-init.sh", size: 1530, mode: os.FileMode(420), modTime: time.Unix(1710469316, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"desktop-scripts/desktop-common.sh":         desktopScriptsDesktopCommonSh,
	"desktop-scripts/desktop-gnome-specific.sh": desktopScriptsDesktopGnomeSpecificSh,
	"desktop-scripts/desktop-init.sh":           desktopScriptsDesktopInitSh,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"desktop-scripts": &bintree{nil, map[string]*bintree{
		"desktop-common.sh":         &bintree{desktopScriptsDesktopCommonSh, map[string]*bintree{}},
		"desktop-gnome-specific.sh": &bintree{desktopScriptsDesktopGnomeSpecificSh, map[string]*bintree{}},
		"desktop-init.sh":           &bintree{desktopScriptsDesktopInitSh, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

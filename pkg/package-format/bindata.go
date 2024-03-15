// Code generated for package package_format by go-bindata DO NOT EDIT. (@generated)
// sources:
// pkg/package-format/appimage/templates/AppRun.sh
package package_format

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

var _appimageTemplatesApprunSh = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x57\xdf\x6f\xdb\x38\x12\x7e\xd7\x5f\x31\x65\x74\x7b\xc9\x21\xb2\xec\xde\x4b\xcf\x81\x6f\xd7\x59\x2b\x89\x50\x27\x0d\x6c\x17\xc8\xa1\x28\x04\x5a\x1a\x5b\x84\x29\x52\x2b\x52\x8d\xbd\x5e\xfd\xef\x07\x52\x92\xeb\x5f\xdd\x76\x81\x7d\xb1\x64\xfa\x9b\xf9\x86\x33\x9c\x6f\xe8\x8b\x37\xfe\x9c\x09\x7f\x4e\x55\xea\x28\xd4\xe0\xa1\xe3\xb0\x05\x7c\x82\x37\xe0\xfd\x0e\xc4\x1d\x05\xb7\x1f\xef\x09\x7c\x86\x1b\xd0\x29\x0a\x07\x00\xc5\x17\x07\xc0\x62\xd7\xce\x82\x39\xce\xec\x21\x9c\x0e\x88\xdb\x25\xce\x05\xa4\x5a\xe7\x7d\xdf\x57\x9a\xc6\x2b\xf9\x05\x8b\x05\x97\xaf\x9d\x58\x66\xfe\x6f\x25\x2a\xcd\xa4\x50\xfe\xbf\x7b\xff\xe9\xbe\xeb\xbd\xf3\x1d\x5a\x2c\xd5\xe0\x92\xb8\xbf\x90\x2b\xe7\xe9\xe3\xe3\x6d\x30\x89\x3e\xdc\x45\xc3\xc9\xbd\xf1\x76\x41\x9a\x38\x6c\x14\xc3\xe7\xe7\x51\x38\x39\x08\xe3\x02\xee\x98\x48\xcc\x37\x18\xe6\xf9\x88\x15\x1d\x08\x35\x30\x65\x57\x12\x56\x60\xac\x65\xb1\x01\x9d\x52\x0d\xb1\x14\x9a\x32\xa1\x0c\x72\x52\x8a\x8e\x35\x9f\xa5\x4c\x01\x55\xaa\xcc\x50\xd5\x30\x6d\x56\x54\x5c\xb0\x5c\x43\x81\x8a\x25\xa8\x80\x09\xf3\xdc\xa3\x01\x59\x00\x05\x55\xce\x77\x1c\xb5\xbb\x70\x71\x60\xcf\x14\x14\xa5\x68\xcd\xa9\x30\xd6\x61\x46\x97\x78\x6d\x37\xd0\x3a\xb4\x4b\x06\xa9\x59\x86\xc0\xd9\x0a\xf9\x06\x52\xaa\x80\xf2\x02\x69\xb2\xb1\x79\x6e\xb6\xef\x00\xe4\x54\xa7\x03\xe2\x5e\x26\xac\x10\x34\x43\x20\xee\xa5\x81\x71\x26\x56\xe0\x2d\x80\xb8\x5b\x53\x8c\x8a\x5c\x91\x2b\xe2\x00\xbc\xa6\x8c\x23\x7c\xfa\x04\xc4\x35\x96\x04\xde\x0c\x80\x10\xf8\xe9\x27\x53\x5d\x6c\x56\x7d\xb7\x47\xe0\xf3\xe7\x1b\x48\xa4\x03\xd0\x70\xb8\x5b\xf3\xf8\x87\xff\xaf\xca\x01\x48\xa4\x40\x07\xa0\x8e\x62\xd0\xf8\xb2\xa5\xc7\x75\x2e\x0b\x0d\xcf\xc3\xd9\xc3\x80\xb8\xdb\x1a\x51\xf5\x77\x6f\x7e\xa9\x0a\x5f\xcd\x99\xe8\xbb\x5b\x03\xaa\x48\x6b\xf2\x32\xba\x8f\x46\xc3\xd9\x30\x1a\x85\x93\xe9\x80\x74\x7c\x95\xd2\x02\xfd\x7e\x6d\x61\xdf\x97\x42\x66\x58\x2f\x70\x19\x53\x7e\x06\xd2\x77\xb7\x07\x8e\xbe\xfa\x1f\x8f\xa2\x71\x78\x3b\x19\x4e\xfe\x17\x1d\x45\x57\x3b\x64\xf3\xbe\xbb\x3d\x02\x7d\x33\xba\x9d\x2d\xd9\xe7\x26\x27\xe4\x27\xd1\xfb\x7f\x1e\x7e\x4b\x77\x3f\x0d\x66\xb3\xf0\xe9\x7e\x1a\x4d\x7f\x7d\x08\x1e\xad\xbb\xe3\x88\x1b\xaf\x9c\xcd\xbd\xb7\x9d\xae\xaf\xe2\x14\x33\xaa\xfa\xee\xf6\x9c\x71\x45\x1c\xe7\x36\x7c\x1a\xb4\x9d\xe3\x6f\xb7\x9d\x60\x8d\x71\xa9\xe9\x9c\xe3\x13\xcd\xb0\xaa\x8e\x1b\x2c\x7c\x1c\xde\x07\x51\xf0\x12\xce\xa2\xe1\xdd\x2c\x98\x44\xe1\xd3\x74\x36\x1c\x8f\x0f\xba\x4e\x17\x34\x07\xaa\x71\xcd\x34\x18\xa8\x3d\x05\x4c\x05\x25\xa7\xc3\x38\xc6\x5c\x63\x32\xe8\x39\x4e\x8d\xb8\xbc\x72\xb6\x0e\x80\x65\x71\x0f\x41\x30\x18\x40\x6f\xdf\x71\x0b\x3b\x14\x02\xf0\xf0\x37\xe8\x1e\xe2\x00\x70\x8d\x31\x10\xf7\x36\x7c\x22\x76\x05\xb9\xc2\xd3\x9f\x4c\x37\x18\x89\xf9\xf4\xcb\xe7\xaa\xc6\x2d\x98\x63\x3f\x2a\xc7\xc1\xa2\x90\xc5\x7e\x7c\xde\x1a\x6c\x9a\x8d\x18\xfe\x8e\x82\xe9\xcd\x21\xeb\xc9\x79\x22\xd0\xc0\x3c\xcf\x3a\x03\xcf\xd3\xb8\xd6\x86\xb5\x57\x11\x78\xfb\x5f\x3f\xc1\x2f\xbe\x28\x39\x37\x9a\xc9\x4f\x48\x56\x09\xa3\x5c\x2e\xbf\xcb\xd2\xe2\x3c\x2f\x53\xcb\xb9\x5c\xff\x30\xc1\xcb\x0f\x12\xbc\xfc\x20\x41\x93\x63\x8c\x53\xd9\x40\x9c\x26\xa5\xf6\x34\xf4\x4c\x5a\x37\xa8\x84\x6c\xd2\x3a\x0b\x67\xe3\x60\xe0\xf6\xcc\x6b\xf0\x32\x1b\xb8\x6f\xff\x8e\x5c\xb7\x73\xc4\xa4\x9b\x69\x8e\x03\xe2\x5a\x22\xd2\xe4\xdf\x7c\x0f\x5e\x66\x07\xe1\xc3\x1f\x7f\xd4\x31\x76\xff\xae\x52\x58\x6a\xd8\xa3\xb6\x1b\x87\x96\xfb\x3b\x7c\x7f\xbd\x32\x27\x7c\x31\x47\x5a\x9c\xf0\xf6\xba\xf0\xae\x7b\xc4\x7e\x50\xb6\x3a\x8d\xd7\xed\x4e\xae\x77\x14\x19\x53\x8a\x89\x65\x07\xa6\x2b\x96\xe7\x4c\x2c\xa1\x19\x25\x9d\xa6\xc5\x5a\x87\x75\xfb\xc4\x29\xc6\xab\x28\xc1\xbc\xa9\xf5\x28\x78\xae\x2b\xdd\x2a\x8a\x7b\xf9\x9a\xb2\x38\xb5\x17\x88\x67\x72\x75\xb8\xd5\xe6\x08\x8d\x82\x67\x33\x26\xff\x22\xf5\xa9\x66\x1d\x08\x54\xbb\xf8\x55\xfa\xea\x99\x4f\xec\x90\x16\x52\x9b\x61\x2b\x0c\xcb\xa2\x90\x19\xbc\x32\x9d\x32\xb1\x3f\x9f\x6f\x20\x45\x11\x23\x94\x26\xa8\x76\x4c\x4f\x4a\x01\x0b\x59\x80\xd1\xd0\x81\xd5\xbc\xed\x96\x2d\xa0\x63\x24\xed\x8e\x71\xac\xaa\x73\x5a\x3a\x0d\xc7\xc1\xd3\xec\xac\x8e\x06\x1f\xc7\xc3\xe8\x71\x38\x79\xdf\x4a\xbd\x99\x24\xbf\x7e\x78\xba\x0b\xef\xa3\x87\x0f\x8f\x41\xdf\x73\xcd\xc3\xef\xc4\x52\x2c\xd8\xb2\x32\x02\xfe\x5c\xc8\xa4\x8c\xb5\xe1\x13\x8d\x82\xef\x3b\xba\x0b\xc7\x66\xd7\x07\x9e\x7d\xdc\xd3\xdc\x3a\x05\x2a\x95\xaf\xd6\x0a\xa4\xe0\x1b\x53\xb0\x04\xd5\x4a\xcb\x1c\x16\xe6\xb2\x90\x48\x54\xe2\x9f\xda\x64\x5d\xe9\xb6\xa0\xcd\x6d\xe1\x90\x8b\x9c\x91\xef\xef\x75\xb6\x09\x40\x0a\x88\xa9\x88\x91\x83\x62\x59\xce\x37\xb6\xc0\x0a\xa8\x48\x40\x96\x45\x3d\x61\x52\x2a\x12\x8e\x05\x70\x5a\x8a\x38\x45\x05\x34\xcf\xaf\x41\xc9\xeb\x93\x39\x62\x6e\x5c\xa8\x21\xc5\x02\x41\x4b\xe8\x5a\x3f\xf5\x0d\x4b\x42\xcf\x44\x65\xf7\x4a\x1b\x7c\x13\xc5\xd1\xc4\xea\x36\xcb\x7f\x26\x3c\x46\x5c\x3c\x26\x16\xf2\xab\xf2\x7c\xad\x49\x33\x51\xc1\xf3\x16\x4d\x75\x0e\x47\xef\xee\x98\x18\x88\x5c\x79\x9c\xce\x91\x0f\x86\xcb\x02\xd1\x34\xb3\x4d\x47\xb3\x38\x62\x8a\xda\xf5\xfa\x84\x85\xea\x41\x67\xdc\x38\xa8\x2a\xcf\x4b\x75\xc6\xb7\x5b\x14\x49\x55\xb5\x03\xcf\xf6\x52\xd1\x07\xf7\xe7\x76\x18\xfe\x98\xbe\x99\x52\xc4\x54\x98\x8e\x58\x98\x8b\x34\x15\x1b\x90\xb9\x95\x56\x2d\xcd\x61\x8f\x11\x6c\x80\xfe\x2e\xa2\x79\xa9\xb5\x14\xca\x76\x42\xe3\xb3\x03\x43\x53\xb7\xc5\x82\xc5\x8c\x72\xc0\x35\xcd\x72\x8e\xe6\x19\x6b\xbe\xb1\xed\x05\x1f\xde\x37\x96\xf6\xdf\x81\xea\xfb\xbe\xc6\x38\x9d\x53\x85\x9d\x55\x82\x1d\x59\x2c\xfd\x11\x7e\x41\x2e\xf3\x0c\x85\xf6\x67\xa5\x96\x05\xa3\x5c\xf9\xd3\x14\x39\x8f\xa6\xf6\x42\xcd\xc4\x32\x32\xde\xa2\xf7\xa3\x20\x1a\x59\x6e\x75\x11\xd4\x74\xd1\xdb\x5e\x27\xaa\x2b\x34\x97\xeb\xa8\x8e\x2c\x9a\xcb\xf5\x6e\xa3\xb6\xc5\x37\x10\x53\x85\xf0\x8a\x90\x53\xa5\xc0\xa6\x5b\x81\xb1\xfa\x76\xf9\xf7\x24\xbf\xf6\x0e\xdf\xae\xeb\x06\x55\x5d\x43\x38\x57\x58\x20\x6d\x1e\x77\x97\x11\xfb\xb4\x31\xb9\x3f\x03\x6b\xcb\xd2\xbd\x6a\x5e\xce\x1c\xd4\xde\xde\x4f\x75\xe5\xc7\x2c\x46\xa1\x70\x77\xc0\xc9\x1e\x22\x5b\x25\xac\x00\x2f\x87\x23\x55\xd8\xc7\x68\x59\x5a\x99\x3e\xea\xed\x06\x71\x73\xb3\x83\xf6\xae\xbe\xc9\x6d\xce\xd0\x39\xfe\x9d\x72\x1f\xb9\xf2\xce\xf8\x1a\x0a\x28\x05\xae\x73\x8c\x4d\x57\xd7\x37\x2a\xf3\x3f\x48\xc6\x71\x59\x14\x98\x74\xc8\xf7\xb3\xd2\x50\xa0\xa2\x71\x3d\x2e\x16\xcc\x69\xba\xe5\xff\x01\x00\x00\xff\xff\xfb\xf1\x3d\x89\xe7\x0e\x00\x00")

func appimageTemplatesApprunShBytes() ([]byte, error) {
	return bindataRead(
		_appimageTemplatesApprunSh,
		"appimage/templates/AppRun.sh",
	)
}

func appimageTemplatesApprunSh() (*asset, error) {
	bytes, err := appimageTemplatesApprunShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "appimage/templates/AppRun.sh", size: 3815, mode: os.FileMode(493), modTime: time.Unix(1710469316, 0)}
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
	"appimage/templates/AppRun.sh": appimageTemplatesApprunSh,
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
	"appimage": &bintree{nil, map[string]*bintree{
		"templates": &bintree{nil, map[string]*bintree{
			"AppRun.sh": &bintree{appimageTemplatesApprunSh, map[string]*bintree{}},
		}},
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

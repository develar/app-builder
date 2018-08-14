package zipx

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/alecthomas/kingpin"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
	"github.com/oxtoacart/bpool"
)

func ConfigureUnzipCommand(app *kingpin.Application) {
	command := app.Command("unzip", "")
	src := command.Flag("input", "").Short('i').Required().String()
	dest := command.Flag("output", "").Short('o').Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		// empty dir must be not used to ensure that some dir will be not removed by mistake, client should clean if need
		err := fsutil.EnsureDir(*dest)
		if err != nil {
			return err
		}

		err = Unzip(*src, *dest, nil)
		if err != nil {
			return err
		}

		return nil
	})
}

// limit write, cpu count can be larger but IO in any case cannot handle a lot of write requests
const concurrency = 8

// https://github.com/mholt/archiver/issues/21
// dest should be an empty dir
func Unzip(src string, dest string, excludedFiles map[string]bool) error {
	if len(src) == 0 {
		return errors.New("input zip file name is empty")
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		// return as is without stack to allow client easily compare error with known zip errors
		return err
	}

	defer util.Close(r)

	extractor := &Extractor{
		outputDir:     dest,
		excludedFiles: excludedFiles,

		createdDirs: make(map[string]bool),
		bufferPool:  bpool.NewBytePool(concurrency, 32*1024),
	}

	// create dirs first (not async)
	for _, zipFile := range r.File {
		if !zipFile.FileInfo().IsDir() {
			continue
		}

		err := extractor.extractDir(zipFile, dest)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// create files async
	err = util.MapAsyncConcurrency(len(r.File), concurrency, func(taskIndex int) (func() error, error) {
		zipFile := r.File[taskIndex]
		if zipFile.FileInfo().IsDir() {
			return nil, nil
		}

		return func() error {
			return extractor.extractAndWriteFile(zipFile)
		}, nil
	})
	if err != nil {
		return errors.WithStack(err)
	}

	// create symlinks
	if extractor.links != nil {
		err = fs.CreateLinks(extractor.links)
		if err != nil {
			return errors.WithStack(err)
		}
		extractor.links = nil
	}

	return nil
}

type Extractor struct {
	mutex sync.RWMutex
	links []fs.LinkInfo

	outputDir     string
	excludedFiles map[string]bool

	createdDirs map[string]bool
	bufferPool  *bpool.BytePool
}

func (t *Extractor) createDirIfNeed(dirPath string) error {
	isDirCreated := false
	t.mutex.RLock()
	_, isDirCreated = t.createdDirs[dirPath]
	t.mutex.RUnlock()

	if isDirCreated {
		return nil
	}

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return err
	}

	t.mutex.Lock()
	t.createdDirs[dirPath] = true
	t.mutex.Unlock()

	return nil
}

func (t *Extractor) extractDir(zipFile *zip.File, dest string) error {
	if zipFile.FileInfo().IsDir() {
		filePath := filepath.Join(dest, zipFile.Name)
		err := os.MkdirAll(filePath, 0777)
		if err != nil {
			return err
		}
		err = fs.SetDirPermsIfNeed(filePath, zipFile.Mode())
		if err != nil {
			return err
		}

		t.createdDirs[filePath] = true
	}
	return nil
}

func (t *Extractor) extractAndWriteFile(zipFile *zip.File) error {
	filePath := filepath.Join(t.outputDir, zipFile.Name)

	if t.excludedFiles != nil {
		_, isExcluded := t.excludedFiles[filePath]
		if isExcluded {
			return nil
		}
	}

	file, err := zipFile.Open()
	if err != nil {
		return err
	}

	defer util.Close(file)

	err = t.createDirIfNeed(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	if zipFile.FileInfo().Mode()&os.ModeSymlink != 0 {
		return t.createSymlink(file, zipFile, filePath)
	}

	buffer := t.bufferPool.Get()
	err = fsutil.WriteFile(file, filePath, zipFile.Mode(), buffer)
	t.bufferPool.Put(buffer)
	if err != nil {
		return err
	}
	return nil
}

func (t *Extractor) createSymlink(reader io.ReadCloser, zipFile *zip.File, filePath string) error {
	buffer := make([]byte, zipFile.FileInfo().Size())
	_, err := io.ReadFull(reader, buffer)
	if err != nil {
		return err
	}

	// symlink cannot be created during copy because symlink can point to not yet copied target file
	t.links = append(t.links, fs.LinkInfo{
		File: filePath,
		Link: string(buffer),
	})
	return nil
}

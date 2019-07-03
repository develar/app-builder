package zipx

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

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
const concurrency = 4

// https://github.com/mholt/archiver/issues/21
// dest must be an empty dir
func Unzip(src string, outputDir string, excludedFiles map[string]bool) error {
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
		outputDir:     filepath.Clean(outputDir),
		excludedFiles: excludedFiles,

		createdDirs: make(map[string]bool),
		bufferPool:  bpool.NewBytePool(concurrency, 64*1024),
	}

	extractor.createdDirs[extractor.outputDir] = true

	lastCreatedDir := ""
	// create files async
	err = util.MapAsyncConcurrency(len(r.File), concurrency, func(taskIndex int) (func() error, error) {
		zipFile := r.File[taskIndex]
		if zipFile.FileInfo().IsDir() {
			// create dir (not async)
			err := extractor.extractDir(zipFile)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}

		filePath, err := extractor.computeExtractPath(zipFile)
		if err != nil {
			return nil, err
		}

		if extractor.excludedFiles != nil {
			_, isExcluded := extractor.excludedFiles[filePath]
			if isExcluded {
				return nil, nil
			}
		}

		fileDir := filepath.Dir(filePath)
		if fileDir != lastCreatedDir {
			err = extractor.createDirIfNeed(fileDir)
			if err != nil {
				return nil, err
			}

			lastCreatedDir = fileDir
		}

		return func() error {
			return extractor.extractAndWriteFile(zipFile, filePath)
		}, nil
	})
	if err != nil {
		return err
	}

	return nil
}

type Extractor struct {
	outputDir     string
	excludedFiles map[string]bool

	createdDirs map[string]bool
	bufferPool  *bpool.BytePool
}

func (t *Extractor) createDirIfNeed(dirPath string) error {
	_, isDirCreated := t.createdDirs[dirPath]
	if isDirCreated {
		return nil
	}

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return err
	}

	t.addWithParentsToCreated(dirPath)
	return nil
}

// check t.createdDirs before create parent dir
func (t *Extractor) MkdirAll(path string, perm os.FileMode) error {
	// fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := os.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}

	// avoid string comparison: dir == t.outputDir, since dir is already checked to has prefix, length check is enough
	minLength := len(t.outputDir)

	// slow path: make sure parent exists and then call Mkdir for path.
	i := len(path)
	for i > minLength && !os.IsPathSeparator(path[i-1]) {
		i--
	}

	if i > minLength {
		// create parent
		parentPath := path[:i-1]
		_, isDirCreated := t.createdDirs[parentPath]
		if !isDirCreated {
			err = t.MkdirAll(parentPath, perm)
			if err != nil {
				return err
			}
		}
	}

	// parent now exists
	err = os.Mkdir(path, perm)
	if err != nil {
		return err
	}

	return nil
}

func (t *Extractor) addWithParentsToCreated(dir string) {
	// avoid string comparison: dir == t.outputDir, since dir is already checked to has prefix, length check is enough
	minLength := len(t.outputDir)
	for {
		t.createdDirs[dir] = true

		i := len(dir)
		for i > minLength && !os.IsPathSeparator(dir[i-1]) {
			i--
		}

		if i <= minLength {
			break
		}

		dir = dir[:i-1]
		_, isDirCreated := t.createdDirs[dir]
		if isDirCreated {
			break
		}
	}
}

func (t *Extractor) computeExtractPath(zipFile *zip.File) (string, error) {
	// #nosec G305
	filePath := filepath.Join(t.outputDir, zipFile.Name)
	if strings.HasPrefix(filePath, t.outputDir) {
		return filePath, nil
	} else {
		return "", errors.Errorf("%s: illegal file path", filePath)
	}
}

func (t *Extractor) extractDir(zipFile *zip.File) error {
	filePath, err := t.computeExtractPath(zipFile)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filePath, 0777)
	if err != nil {
		return err
	}

	err = fs.SetNormalDirPermissions(filePath)
	if err != nil {
		return err
	}

	t.addWithParentsToCreated(filePath)
	return nil
}

func (t *Extractor) extractAndWriteFile(zipFile *zip.File, filePath string) error {
	file, err := zipFile.Open()
	if err != nil {
		return errors.WithStack(err)
	}

	defer util.Close(file)

	if (zipFile.FileInfo().Mode() & os.ModeSymlink) != 0 {
		return t.createSymlink(file, zipFile, filePath)
	}

	buffer := t.bufferPool.Get()
	err = fs.WriteFileAndRestoreNormalPermissions(file, filePath, zipFile.Mode(), buffer)
	t.bufferPool.Put(buffer)
	if err != nil {
		return err
	}
	return nil
}

func (t *Extractor) createSymlink(reader io.Reader, zipFile *zip.File, filePath string) error {
	buffer := make([]byte, zipFile.FileInfo().Size())
	_, err := io.ReadFull(reader, buffer)
	if err != nil {
		return err
	}

	return os.Symlink(string(buffer), filePath)
}

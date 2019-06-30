package appimage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/package-format"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
)

const iconDirRelativePath = "usr/share/icons/hicolor"
const mimeTypeDirRelativePath = "usr/share/mime"

type TemplateConfiguration struct {
	EulaFile        string
	ExecutableName  string
	ProductName     string
	ProductFilename string
	ResourceName    string
	DesktopFileName string

	MimeTypeFile string
}

func (t *TemplateConfiguration) IsHtmlEula() bool {
	return strings.HasSuffix(t.EulaFile, ".html")
}

// https://github.com/AppImage/AppImageKit/issues/438#issuecomment-319094239
// expects icons in the /usr/share/icons/hicolor
func copyIcons(options *AppImageOptions) error {
	stageDir := *options.stageDir

	iconCommonDir := filepath.Join(stageDir, iconDirRelativePath)
	err := fsutil.EnsureDir(iconCommonDir)
	if err != nil {
		return errors.WithStack(err)
	}

	icons := options.configuration.Icons
	iconExtWithDot := filepath.Ext(icons[0].File)
	iconFileName := options.configuration.ExecutableName + iconExtWithDot
	maxIconIndex := len(icons) - 1
	var fileCopier fs.FileCopier
	fileCopier.IsUseHardLinks = true
	err = util.MapAsync(len(icons), func(taskIndex int) (func() error, error) {
		icon := icons[taskIndex]
		var iconSizeDir string
		if iconExtWithDot == ".svg" {
			// https://bugs.freedesktop.org/show_bug.cgi?id=91759
			iconSizeDir = "scalable"
		} else {
			iconSizeDir = fmt.Sprintf("%dx%d/apps", icon.Size, icon.Size)
		}
		iconRelativeToStageFile := iconDirRelativePath + "/" + iconSizeDir + "/" + iconFileName

		return func() error {
			iconDir := filepath.Join(iconCommonDir, iconSizeDir)
			err := fsutil.EnsureDir(iconDir)
			if err != nil {
				return errors.WithStack(err)
			}

			fakeFileInfo := &fakeFileInfo{}
			iconFile := filepath.Join(iconDir, iconFileName)
			err = fileCopier.CopyFile(icon.File, iconFile, false, fakeFileInfo)
			if err != nil {
				return errors.WithStack(err)
			}

			if taskIndex == maxIconIndex {
				err = os.Symlink(iconRelativeToStageFile, filepath.Join(stageDir, iconFileName))
				if err != nil {
					return errors.WithStack(err)
				}

				err = os.Symlink(iconRelativeToStageFile, filepath.Join(stageDir, ".DirIcon"))
				if err != nil {
					return errors.WithStack(err)
				}
			}

			return nil
		}, nil
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type fakeFileInfo struct {
	dir      bool
	basename string
	contents string
}

func (f *fakeFileInfo) Name() string       { return f.basename }
func (f *fakeFileInfo) Sys() interface{}   { return nil }
func (f *fakeFileInfo) ModTime() time.Time { return time.Now() }
func (f *fakeFileInfo) IsDir() bool        { return f.dir }
func (f *fakeFileInfo) Size() int64        { return int64(len(f.contents)) }
func (f *fakeFileInfo) Mode() os.FileMode {
	return 0644
}

func copyMimeTypes(options *AppImageOptions) (string, error) {
	var mimeTypes strings.Builder
	for _, fileAssociation := range options.configuration.FileAssociations {
		if fileAssociation.MimeType != "" {
			mimeTypes.WriteString("<mime-type type=\"")
			mimeTypes.WriteString(fileAssociation.MimeType)
			mimeTypes.WriteString("\">\n")

			mimeTypes.WriteString("  <comment>")
			mimeTypes.WriteString(options.configuration.ProductName)
			mimeTypes.WriteString(" document</comment>\n")

			mimeTypes.WriteString("  <glob pattern=\"*.")
			mimeTypes.WriteString(fileAssociation.Ext)
			mimeTypes.WriteString("\"/>\n")

			mimeTypes.WriteString("  <generic-icon name=\"x-office-document\"/>\n")

			mimeTypes.WriteString("</mime-type>\n")
		}
	}

	// if no mime-types specified, return
	if mimeTypes.Len() == 0 {
		return "", nil
	}

	mimeTypeDir := filepath.Join(*options.stageDir, mimeTypeDirRelativePath)
	fileName := options.configuration.ExecutableName + ".xml"
	mimeTypeFile := filepath.Join(mimeTypeDir, fileName)
	err := fsutil.EnsureDir(mimeTypeDir)
	if err != nil {
		return "", errors.WithStack(err)
	}

	err = ioutil.WriteFile(mimeTypeFile, []byte("<?xml version=\"1.0\"?>\n<mime-info xmlns=\"http://www.freedesktop.org/standards/shared-mime-info\">\n"+mimeTypes.String()+"\n</mime-info>"), 0666)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return mimeTypeDirRelativePath + "/" + fileName, nil
}

func writeDesktopFile(options *AppImageOptions) (string, error) {
	fileName := options.configuration.ExecutableName + ".desktop"
	err := ioutil.WriteFile(filepath.Join(*options.stageDir, fileName), []byte(options.configuration.DesktopEntry), 0666)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return fileName, nil
}

func writeAppLauncherAndRelatedFiles(options *AppImageOptions) error {
	var t *template.Template
	if *options.template == "" {
		data, err := package_format.Asset("appimage/templates/AppRun.sh")
		if err != nil {
			return errors.WithStack(err)
		}

		t = template.New("AppRun.sh")
		t, err = t.Parse(string(data))
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		var err error
		t, err = template.ParseFiles(*options.template)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	desktopFileName, err := writeDesktopFile(options)
	if err != nil {
		return errors.WithStack(err)
	}

	configuration := options.configuration
	executableName := configuration.ExecutableName
	templateConfiguration := &TemplateConfiguration{
		DesktopFileName: desktopFileName,
		ExecutableName:  executableName,
		ProductName:     configuration.ProductName,
		ProductFilename: configuration.ProductFilename,
		ResourceName:    "appimagekit-" + executableName,
	}

	licenseFile := *options.license
	if licenseFile != "" {
		templateConfiguration.EulaFile = filepath.Base(licenseFile)
		err := fs.CopyUsingHardlink(licenseFile, filepath.Join(*options.stageDir, templateConfiguration.EulaFile))
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = copyIcons(options)
	if err != nil {
		return err
	}

	mimeTypeFile, err := copyMimeTypes(options)
	if err != nil {
		return errors.WithStack(err)
	}

	templateConfiguration.MimeTypeFile = mimeTypeFile

	templateFilename := filepath.Join(*options.stageDir, "AppRun")
	f, err := os.Create(templateFilename)
	defer util.Close(f)
	if err != nil {
		return err
	}

	err = t.Execute(f, templateConfiguration)
	if err != nil {
		return err
	}

	util.Close(f)

	err = os.Chmod(templateFilename, 0755)
	if err != nil {
		return err
	}

	return nil
}

package asar

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

type ResolvedFileSet struct {
	Src         string `json:"src"`
	Destination string `json:"destination"`

	Files []string `json:"files"`
}

func BuildAsar(outFilePath string) error {
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	defer outFile.Close()

	_, err = io.Copy(outFile, os.Stdin)
	if err != nil {
		return errors.WithStack(err)
	}

	var fileSets []ResolvedFileSet
	//indexToIsUnpacked := make(map[int]string)

	for _, fileSet := range fileSets {
		for _, file := range fileSet.Files {
			//if indexToIsUnpacked
			copyFileContents(file, outFile)
		}
	}

	return nil
}

func copyFileContents(src string, out *os.File) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}

	_, err = io.Copy(out, in)
	if err != nil {
		in.Close()
		return
	}

	err = in.Close()
	if err != nil {
		return err
	}

	return
}

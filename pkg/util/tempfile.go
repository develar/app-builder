// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/develar/errors"
)

// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randMutex sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextPrefix() string {
	randMutex.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randMutex.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

// TempFile creates a new temporary file in the directory dir
// with a name beginning with prefix, opens the file for reading
// and writing, and returns the resulting *os.File.
// If dir is the empty string, TempFile uses the default directory
// for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously
// will not choose the same file. The caller can use f.Name()
// to find the pathname of the file. It is the caller's responsibility
// to remove the file when no longer needed.
func TempFile(dir, suffix string) (string, error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nConflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, nextPrefix()+suffix)
		_, err := os.Lstat(name)
		if os.IsNotExist(err) {
			return name, nil
		}

		if nConflict++; nConflict > 10 {
			randMutex.Lock()
			rand = reseed()
			randMutex.Unlock()
		}
	}
	return "", errors.Errorf("cannot find unique file name")
}

// TempDir creates a new temporary directory in the directory dir
// with a name beginning with prefix and returns the path of the
// new directory. If dir is the empty string, TempDir uses the
// default directory for temporary files (see os.TempDir).
// Multiple programs calling TempDir simultaneously
// will not choose the same directory. It is the caller's responsibility
// to remove the directory when no longer needed.
func TempDir(dir, suffix string) (name string, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nConflict := 0
	for i := 0; i < 10000; i++ {
		try := filepath.Join(dir, nextPrefix()+suffix)
		err = os.Mkdir(try, 0700)
		if os.IsExist(err) {
			if nConflict++; nConflict > 10 {
				randMutex.Lock()
				rand = reseed()
				randMutex.Unlock()
			}
			continue
		}
		if os.IsNotExist(err) {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return "", err
			}
		}
		if err == nil {
			name = try
		}
		break
	}
	return
}

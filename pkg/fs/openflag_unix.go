//go:build !windows

package fs

import "syscall"

// O_NOFOLLOW makes open fail if the final path component is a symlink
const openNoFollowFlag = syscall.O_NOFOLLOW

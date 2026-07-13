//go:build windows

package fs

// windows has no O_NOFOLLOW; creating a symlink there already requires elevation
const openNoFollowFlag = 0

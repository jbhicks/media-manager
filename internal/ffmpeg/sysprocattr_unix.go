//go:build !windows

package ffmpeg

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {
	// No-op on non-Windows platforms
}

//go:build windows

package local_runtime

import (
	"fmt"
	"os/exec"
)

// On Windows, do nothing for process group. Consider Job Objects for full tree control.
func setProcGroup(cmd *exec.Cmd) {}

// tryKillProcessGroup is unsupported on Windows. Return a getpgid-like error for caller to log and fallback.
func tryKillProcessGroup(cmd *exec.Cmd) (int, error, error) {
	return 0, fmt.Errorf("process group unsupported on windows"), nil
}

func killProcess(cmd *exec.Cmd) error {
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Kill()
	}
	return nil
}

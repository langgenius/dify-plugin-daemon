//go:build darwin || linux || freebsd || netbsd || openbsd

package local_runtime

import (
	"os/exec"
	"syscall"
)

// setProcGroup places the child into its own process group (Unix only)
func setProcGroup(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// tryKillProcessGroup tries to kill the process group and returns diagnostic errors.
// Returns: (pgid, getpgidErr, killErr)
func tryKillProcessGroup(cmd *exec.Cmd) (int, error, error) {
	if cmd == nil || cmd.Process == nil {
		return 0, syscall.EINVAL, nil
	}
	pid := cmd.Process.Pid
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return 0, err, nil
	}
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		return pgid, nil, err
	}
	return pgid, nil, nil
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

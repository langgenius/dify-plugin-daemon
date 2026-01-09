//go:build !(darwin || linux || freebsd || netbsd || openbsd || windows)

package local_runtime

import "os/exec"

func setProcGroup(cmd *exec.Cmd) {}

func tryKillProcessGroup(cmd *exec.Cmd) (int, error, error) { return 0, nil, nil }

func killProcess(cmd *exec.Cmd) error { if cmd != nil && cmd.Process != nil { return cmd.Process.Kill() }; return nil }
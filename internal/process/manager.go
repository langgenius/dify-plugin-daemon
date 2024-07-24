package process

import (
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

var (
	subprocesses map[int]*exec.Cmd
	l            *sync.Mutex
)

func Init() {
	subprocesses = make(map[int]*exec.Cmd)
	l = &sync.Mutex{}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		TerminateAll()
		os.Exit(0)
	}()
}

func WrapProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func NewProcess(cmd *exec.Cmd) {
	l.Lock()
	defer l.Unlock()
	subprocesses[cmd.Process.Pid] = cmd
}

func RemoveProcess(cmd *exec.Cmd) {
	l.Lock()
	defer l.Unlock()

	delete(subprocesses, cmd.Process.Pid)
}

func TerminateAll() {
	l.Lock()
	defer l.Unlock()

	for _, cmd := range subprocesses {
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	}
}

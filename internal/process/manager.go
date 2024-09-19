package process

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"sync"
	"syscall"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

var (
	l               *sync.Mutex
	subprocess_path string
)

func subprocesses() []int {
	if _, err := os.Stat(subprocess_path); err != nil {
		if err == os.ErrNotExist {
			os.MkdirAll(path.Dir(subprocess_path), 0755)
			os.WriteFile(subprocess_path, []byte{}, 0644)
		} else {
			log.Error("Error checking subprocesses file")
			return []int{}
		}
	}

	data, err := os.ReadFile(subprocess_path)
	if err != nil {
		if err != os.ErrNotExist {
			log.Error("Error reading subprocesses file")
		}
		return []int{}
	}
	nums := bytes.Split(data, []byte("\n"))
	procs := make([]int, 0)
	for _, num := range nums {
		if len(num) == 0 {
			continue
		}
		proc, err := strconv.Atoi(string(num))
		if err != nil {
			log.Error("Error parsing subprocesses file")
			return []int{}
		}
		procs = append(procs, proc)
	}

	return procs
}

func addSubprocess(pid int) {
	l.Lock()
	defer l.Unlock()

	procs := subprocesses()
	procs = append(procs, pid)
	data := []byte{}
	for _, proc := range procs {
		data = append(data, []byte(strconv.Itoa(proc)+"\n")...)
	}
	os.WriteFile(subprocess_path, data, 0644)
}

func removeSubprocess(pid int) {
	l.Lock()
	defer l.Unlock()

	procs := subprocesses()
	new_procs := []int{}
	for _, proc := range procs {
		if proc == pid {
			continue
		}
		new_procs = append(new_procs, proc)
	}
	data := []byte{}
	for _, proc := range new_procs {
		data = append(data, []byte(strconv.Itoa(proc)+"\n")...)
	}
	os.WriteFile(subprocess_path, data, 0644)
}

func clearSubprocesses() {
	os.WriteFile(subprocess_path, []byte{}, 0644)
}

func Init(config *app.Config) {
	l = &sync.Mutex{}
	subprocess_path = config.ProcessCachingPath

	sig_exit := make(chan os.Signal, 1)
	signal.Notify(sig_exit, os.Interrupt, syscall.SIGTERM)
	sig_reload := make(chan os.Signal, 1)
	signal.Notify(sig_reload, syscall.SIGUSR2)

	// kill all subprocesses
	TerminateAll()

	go func() {
		for {
			select {
			case <-sig_reload:
				TerminateAll()
			case <-sig_exit:
				TerminateAll()
				os.Exit(0)
			}
		}
	}()
}

func WrapProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func NewProcess(cmd *exec.Cmd) {
	addSubprocess(cmd.Process.Pid)
}

func RemoveProcess(cmd *exec.Cmd) {
	removeSubprocess(cmd.Process.Pid)
}

func TerminateAll() {
	l.Lock()
	defer l.Unlock()

	for _, pid := range subprocesses() {
		log.Info("Killing uncleaned subprocess %d", pid)
		syscall.Kill(-pid, syscall.SIGKILL)
	}

	clearSubprocesses()
}

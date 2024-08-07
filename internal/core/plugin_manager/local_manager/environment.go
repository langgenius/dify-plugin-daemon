package local_manager

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/checksum"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (r *LocalPluginRuntime) InitEnvironment() error {
	if _, err := os.Stat(path.Join(r.State.AbsolutePath, ".installed")); err == nil {
		return nil
	}

	// execute init command
	handle := exec.Command("bash", r.Config.Execution.Install)
	handle.Dir = r.State.AbsolutePath
	handle.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// get stdout and stderr
	stdout, err := handle.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	stderr, err := handle.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	// start command
	if err := handle.Start(); err != nil {
		return err
	}
	defer func() {
		if handle.Process != nil {
			handle.Process.Kill()
		}
	}()

	var err_msg strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)

	last_active_at := time.Now()

	routine.Submit(func() {
		defer wg.Done()
		// read stdout
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			log.Info("installing %s - %s", r.Config.Identity(), string(buf[:n]))
			last_active_at = time.Now()
		}
	})

	routine.Submit(func() {
		defer wg.Done()
		// read stderr
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil && err != os.ErrClosed {
				last_active_at = time.Now()
				err_msg.WriteString(string(buf[:n]))
				break
			} else if err == os.ErrClosed {
				break
			}

			if n > 0 {
				err_msg.WriteString(string(buf[:n]))
				last_active_at = time.Now()
			}
		}
	})

	routine.Submit(func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if handle.ProcessState != nil && handle.ProcessState.Exited() {
				break
			}

			if time.Since(last_active_at) > 60*time.Second {
				handle.Process.Kill()
				err_msg.WriteString("init process exited due to long time no activity")
				break
			}
		}
	})

	wg.Wait()

	if err_msg.Len() > 0 {
		return fmt.Errorf("install failed: %s", err_msg.String())
	}

	if err := handle.Wait(); err != nil {
		return err
	}

	// create .installed file
	f, err := os.Create(path.Join(r.State.AbsolutePath, ".installed"))
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}

func (r *LocalPluginRuntime) calculateChecksum() string {
	plugin_decoder, err := decoder.NewFSPluginDecoder(r.CWD)
	if err != nil {
		return ""
	}

	checksum, err := checksum.CalculateChecksum(plugin_decoder)
	if err != nil {
		return ""
	}

	return checksum
}

func (r *LocalPluginRuntime) Checksum() string {
	if r.checksum == "" {
		r.checksum = r.calculateChecksum()
	}

	return r.checksum
}

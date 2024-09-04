package local_manager

import "os/exec"

func (p *LocalPluginRuntime) InitPythonEnvironment(requirements_txt string) error {
	// create virtual env
	identity, err := p.Identity()
	if err != nil {
		return err
	}

	cmd := exec.Command("python", "-m", "venv", identity.String())

	// set working directory
	cmd.Dir = p.WorkingPath

	// TODO
	return nil
}

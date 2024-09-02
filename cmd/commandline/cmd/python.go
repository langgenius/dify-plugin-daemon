package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed templates/python/main.py
var PYTHON_ENTRYPOINT_TEMPLATE []byte

//go:embed templates/python/requirements.txt
var PYTHON_REQUIREMENTS_TEMPLATE []byte

func createPythonEnvironment(root string, entrypoint string) error {
	// create the python environment
	entrypoint_file_path := filepath.Join(root, fmt.Sprintf("%s.py", entrypoint))
	if err := os.WriteFile(entrypoint_file_path, PYTHON_ENTRYPOINT_TEMPLATE, 0o644); err != nil {
		return err
	}

	requirements_file_path := filepath.Join(root, "requirements.txt")
	if err := os.WriteFile(requirements_file_path, PYTHON_REQUIREMENTS_TEMPLATE, 0o644); err != nil {
		return err
	}

	return nil
}

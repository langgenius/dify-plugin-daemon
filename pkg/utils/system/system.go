package system

import (
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var DelimiterFLag string

func init() {
	if runtime.GOOS == "windows" {
		DelimiterFLag = "#"
	} else {
		DelimiterFLag = ":"
	}
}

func ConvertPath(input string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(input, "\\", "/")
	}
	return input
}

func GetZipReadPath(dir string, filename string) string {
	if runtime.GOOS == "windows" {
		return path.Join(dir, filename)
	}
	return filepath.Join(dir, filename)
}

func GetEnvPythonPath(envPath string) string {
	if runtime.GOOS == "windows" {
		return envPath + "/Scripts/python.exe"
	}
	return envPath + "/bin/python"
}

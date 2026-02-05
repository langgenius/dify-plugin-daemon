package system

import (
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
		return strings.ReplaceAll(input, "\\", "/")[1:]
	}
	return input
}

func GetZipReadPath(dir string, filename string) string {
	if runtime.GOOS == "windows" {
		return dir + filename
	}
	return filepath.Join(dir, filename)
}

func GetEnvPythonPath(envPath string) string {
	if runtime.GOOS == "windows" {
		return envPath + "\\Scripts\\python.exe"
	}
	return envPath + "/bin/python"
}

func GetEnvValidFlagFile(envPath string) string {
	if runtime.GOOS == "windows" {
		return envPath + "\\dify\\plugin.json"
	}
	return envPath + "/dify/plugin.json"
}

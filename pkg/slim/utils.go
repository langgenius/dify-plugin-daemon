package slim

import (
	"os"
	"strconv"
)

func env(name string, d string) string {
	v := os.Getenv(name)
	if v == "" {
		return d
	}
	return v
}

func envInt(name string, d int) int {
	v := os.Getenv(name)
	i, err := strconv.Atoi(v)
	if err != nil || v == "" {
		return d
	}
	return i
}

package tests

import "fmt"

func ReadableBytes(l int) string {
	// convert l bytes to a readable string
	if l < 1024 {
		return fmt.Sprintf("%d B", l)
	}

	if l < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(l)/1024)
	}

	if l < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(l)/(1024*1024))
	}

	return fmt.Sprintf("%.2f GB", float64(l)/(1024*1024*1024))
}

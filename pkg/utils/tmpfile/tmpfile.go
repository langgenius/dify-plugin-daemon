package tmpfile

import "os"

// CreateTempFile creates a temp file with the given prefix
// and returns the path to the temp file and a function to clean up the temp file
func CreateTempFile(prefix string) (*os.File, func(), error) {
	file, err := os.CreateTemp(os.TempDir(), prefix)
	if err != nil {
		return nil, nil, err
	}
	return file, func() { os.Remove(file.Name()) }, nil
}

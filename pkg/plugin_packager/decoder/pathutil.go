package decoder

import (
	"fmt"
	"path"
	"strings"
)

// normalizeLogicalPath converts any manifest-provided path (which may contain
// either '/' or '\\', and may include redundant segments) into a normalized
// forward-slash, relative form suitable for OS-independent matching (path.Match)
// and for decoders (Zip requires '/', FS will convert at boundary).
// It returns an error if the input is an absolute path, has a Windows drive letter,
// or attempts to traverse outside the package (.. prefix).
func normalizeLogicalPath(p string) (string, error) {
	// Trim surrounding spaces
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	// Treat backslash as a path separator from manifest authors (Windows habit)
	p = strings.ReplaceAll(p, "\\", "/")
	orig := p

	// Reject Windows drive-letter paths like C:/... or C:...
	if len(p) >= 2 && p[1] == ':' {
		c := p[0]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			return "", fmt.Errorf("invalid manifest path (drive letter): %q", p)
		}
	}

	// Clean collapses //, ./, ../ etc using POSIX semantics over '/'
	p = path.Clean(p)

	// Reject absolute paths explicitly
	if strings.HasPrefix(p, "/") {
		return "", fmt.Errorf("invalid manifest path (absolute): %q", p)
	}

	// Prevent path traversal: after cleaning, any leading ".." means escaping base
	if p == ".." || strings.HasPrefix(p, "../") {
		return "", fmt.Errorf("invalid manifest path (traversal): %q", p)
	}

	// path.Clean("") => "."; if it collapses to root due to parent refs, reject
	if p == "." {
		// detect if original contained parent traversal segments
		if strings.Contains(orig, "../") || strings.HasSuffix(orig, "/..") || strings.HasPrefix(orig, "../") || orig == ".." || strings.Contains(orig, "/../") {
			return "", fmt.Errorf("invalid manifest path (collapses to root via traversal): %q", orig)
		}
		p = ""
	}

	// Keep it relative: drop leading './'
	p = strings.TrimPrefix(p, "./")

	return p, nil
}

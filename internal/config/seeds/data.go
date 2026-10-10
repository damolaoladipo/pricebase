package seeds

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// dataFile resolves the absolute path to a seed JSON file by walking up from
// the current working directory and checking known _data layout locations.
//
// Seed loaders (readJSON and the Seed* functions) depend on this helper to find
// permissions.json, roles.json, and users.json regardless of whether the process
// was started from the repo root, pricebase/, or a nested package directory.
// The search is filesystem-based and does not embed assets in the binary.
//
// name is the filename only (for example "permissions.json"). The function
// records each candidate path in a deduplication map, walks at most eight parent
// directories from cwd, and for each level tries _data/name,
// internal/_data/name, src/_data/name, pricebase/_data/name, and
// pricebase/internal/_data/name relative to that directory.
//
// Candidates are checked in discovery order with os.Stat; the first path that
// exists is returned. If cwd cannot be read, that error is returned immediately.
// When no candidate exists after the walk completes, the function returns an
// empty string and os.ErrNotExist. It does not validate JSON or file permissions
// beyond existence.
func dataFile(name string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	seen := map[string]bool{}
	var candidates []string

	dir := cwd
	for range 8 {
		for _, rel := range []string{
			filepath.Join("_data", name),
			filepath.Join("internal", "_data", name),
			filepath.Join("src", "_data", name),
			filepath.Join("pricebase", "_data", name),
			filepath.Join("pricebase", "internal", "_data", name),
		} {
			candidate := filepath.Join(dir, rel)
			if seen[candidate] {
				continue
			}
			seen[candidate] = true
			candidates = append(candidates, candidate)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", os.ErrNotExist
}

// readJSON loads a seed JSON file from disk and unmarshals its contents into dest.
//
// It is the shared I/O layer for all Seed* functions. Callers pass a filename
// (such as "roles.json") and a pointer to a slice or struct matching the file
// schema; readJSON handles path resolution, file read, and JSON decode in one step.
//
// name is forwarded to dataFile to locate the file under _data. dest must be a
// non-nil pointer suitable for encoding/json.Unmarshal (typically *[]T or *T).
// The entire file is read into memory before unmarshaling; there is no streaming
// or partial decode.
//
// When dataFile fails (including os.ErrNotExist), that error is returned unchanged.
// os.ReadFile errors (permissions, I/O) propagate as-is. Invalid JSON returns the
// unmarshaling error from encoding/json. On success, dest is populated and nil is
// returned.
func readJSON(name string, dest any) error {
	path, err := dataFile(name)
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, dest)
}

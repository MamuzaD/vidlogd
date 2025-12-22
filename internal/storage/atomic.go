package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func WriteFileAtomic(path string, data []byte, perm os.FileMode) (retErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	f, err := os.CreateTemp(dir, "tmp-"+filepath.Base(path))
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	tmpName := f.Name()
	defer func() {
		if retErr != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if err := f.Chmod(perm); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing data: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("syncing file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := atomicRename(tmpName, path); err != nil {
		return fmt.Errorf("renaming: %w", err)
	}

	if runtime.GOOS != "windows" {
		df, err := os.Open(dir)
		if err != nil {
			return fmt.Errorf("opening directory for sync: %w", err)
		}

		defer df.Close()
		if err := df.Sync(); err != nil {
			return fmt.Errorf("syncing directory: %w", err)
		}
	}

	return nil
}

func atomicRename(tmpName, path string) error {
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}
	return os.Rename(tmpName, path)
}

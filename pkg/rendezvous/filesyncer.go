package rendezvous

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/pkg/errors"
)

// The FileSyncer creates a file when Ready is called. It then waits until the file is removed,
// which it treats as a signal that all other processes are ready as well, and returns from Ready.
// The FileSyncer is intended to be use together with another external process
// that waits until all files have been created and then removes them at once.
type FileSyncer struct {
	fileName     string
	pollInterval time.Duration
}

// NewFileSyncer returns a new FileSyncer using a file with the given fileName.
func NewFileSyncer(fileName string, pollInterval time.Duration) *FileSyncer {
	return &FileSyncer{
		fileName:     fileName,
		pollInterval: pollInterval,
	}
}

// Ready creates a file and periodically checks for the existence of the file with the configured pollInterval.
// Ready returns when that file has been removed, the context has been canceled, or an error occurred.
// Also, if the file already exists, prior to invocation, Ready immediately returns an error.
func (fs *FileSyncer) Ready(ctx context.Context) error {

	// Initially, the file must not exist.
	// Return an error if it does.
	exists, err := fileExists(fs.fileName)
	if err != nil {
		return err
	} else if exists {
		return fmt.Errorf("file already exists: %s", fs.fileName)
	}

	// Create the synchronization file.
	if f, err := os.Create(fs.fileName); err != nil {
		return fmt.Errorf("cannot create file %s: %w", fs.fileName, err)
	} else if err = f.Close(); err != nil {
		return fmt.Errorf("cannot close file %s: %w", fs.fileName, err)
	}

	// Periodically poll the file system while the file still exists and no error occurs.
	for exists, err = fileExists(fs.fileName); exists && err == nil; exists, err = fileExists(fs.fileName) {
		select {
		case <-time.After(fs.pollInterval):
		case <-ctx.Done():
			return fmt.Errorf("context canceled")
		}
	}

	// If the loop finished without an error occurring, the file has been removed and err is nil.
	// Otherwise, we still return the error.
	return err
}

// fileExists returns true if the given file already exists and is not a directory.
// If the file does not exist, fileExists returns false.
// If the file is a directory or an error occurs, fileExists returns false and an error.
func fileExists(fileName string) (bool, error) {
	s, err := os.Stat(fileName)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	} else if s.IsDir() {
		return false, fmt.Errorf("file exists, but is a directory: %s", fileName)
	}
	return true, nil
}

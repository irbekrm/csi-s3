package filesystem

//go:generate mockgen -source=main.go -destination=../../mocks/mock_filesystem.go -package=mocks
import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/google/fscrypt/filesystem"
)

// FS contains high level methods for interacting with filesystem
type FS interface {
	FindMount(string) (Matcher, error)
	EnsureMountRemoved(string) error
	EnsureDirExists(string) error
}

// NewFS returns an FS implementation that will interact with actual filesystem
func NewFS() FS {
	return fs{
		sys: sys{},
	}
}

type fs struct {
	sys Sys
}

// FindMount looks for a mount at path, returns mount (nil if it doesn't exist) and error
func (f fs) FindMount(path string) (Matcher, error) {
	_, err := f.sys.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// TODO: find a more reliable way to check if mount exists
	mnt, err := f.sys.GetMount(path)
	if err != nil && strings.Contains(err.Error(), "is not a mountpoint") {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mount{readonly: mnt.ReadOnly, fsType: mnt.FilesystemType}, nil
}

// EnsureMountRemoved idempotently removes mounted filesystem
func (f fs) EnsureMountRemoved(path string) error {
	if _, err := f.sys.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_, err := f.sys.GetMount(path)
	if err != nil && strings.Contains(err.Error(), "is not a mountpoint") {
		return f.sys.Remove(path)
	}
	if err := f.sys.Unmount(path); err != nil {
		return err
	}
	return f.sys.Remove(path)
}

// EnsureDirExists idempotently makes a directory with os.ModePerm at path
func (f fs) EnsureDirExists(path string) error {
	finfo, err := f.sys.Stat(path)
	// if the directory does not exist, make it
	if os.IsNotExist(err) {
		return f.sys.Mkdir(path, os.ModePerm)
	}
	if err != nil {
		return err
	}
	// a file exists at the path, check that it's a directory
	//TODO: check permissions
	if finfo.Mode().IsDir() {
		return nil
	}
	// should never get to this line
	return fmt.Errorf("file with permissions %s found at path %s", finfo.Mode().Perm().String(), path)
}

// TODO: Match should check for volume capabilities
type Matcher interface {
	Match(string, bool) bool
}

type mount struct {
	readonly bool
	fsType   string
}

// Match checks if mount has the given properties
func (m mount) Match(fsType string, readonly bool) bool {
	return m.readonly == readonly && fsType == m.fsType
}

// Sys contains low level methods for interacting with filesystem
type Sys interface {
	Stat(string) (os.FileInfo, error)
	Unmount(string) error
	Remove(string) error
	GetMount(string) (*filesystem.Mount, error)
	Mkdir(string, os.FileMode) error
}

type sys struct{}

// Stat is a wrapper around os.Stat
func (s sys) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Unmount is a wrapper around syscall.Unmount
func (s sys) Unmount(path string) error {
	return syscall.Unmount(path, 0)
}

// Remove is a wrapper around os.Remove
func (s sys) Remove(path string) error {
	return os.Remove(path)
}

// GetMount is a wrapper around filesystem.GetMount
func (s sys) GetMount(path string) (*filesystem.Mount, error) {
	return filesystem.GetMount(path)
}

// Mkdir is a wrapper around os.Mkdir
func (s sys) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

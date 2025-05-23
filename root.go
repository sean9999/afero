package afero

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

var ErrInvalidRoot = errors.New("invalid root")

var _ Root = (*rootedFs)(nil)

// a *rootedFs is an Fs that can satisfy [Root]
type rootedFs struct {
	*BasePathFs
}

// NewRootedFs creates a rootedFs on top of filesystem, rooted at rootDir
func NewRootedFs(fileSystem Fs, rootDir string) (*rootedFs, error) {
	info, err := fileSystem.Stat(rootDir)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrInvalidRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w. not a directory", ErrInvalidRoot)
	}

	b := &BasePathFs{
		source: fileSystem,
		path:   rootDir,
	}
	return &rootedFs{b}, nil
}

// Name returns the directory that this *rootedFs is rooted at
func (ms *rootedFs) Name() string {
	return ms.path
}

func (ms *rootedFs) Open(name string) (File, error) {
	if !filepath.IsLocal(name) {
		return nil, fmt.Errorf("%w. %s is not local", ErrInvalidRoot, name)
	}
	return ms.BasePathFs.Open(name)
}

// Close closes a Root, and makes it stop working
func (m *rootedFs) Close() error {
	if m.BasePathFs == nil {
		return fmt.Errorf("could not close. %w", ErrInvalidRoot)
	}
	m.BasePathFs = nil
	return nil
}

func (m *rootedFs) FS() Fs {
	// luckily, *rootedFs already is an Fs, so just return it
	return m
}

func (m *rootedFs) Lstat(name string) (fs.FileInfo, error) {
	info, _, err := m.BasePathFs.LstatIfPossible(name)
	return info, err
}

// OpenRoot opens a [Root] rooted in *rootedFs
func (m *rootedFs) OpenRoot(name string) (Root, error) {

	//	the new root is the old root with a new path
	subFs, err := NewRootedFs(m.source, filepath.Join(m.path, name))
	if err != nil {
		return nil, fmt.Errorf("could not open root. %w", err)
	}

	return subFs, nil
}

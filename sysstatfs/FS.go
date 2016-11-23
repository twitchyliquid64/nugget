package sysstatfs

import "bazil.org/fuse/fs"

// Make creates a minimal FS node.
func Make() *FS {
	ret := &FS{
		InodeCount: 1,
		Variables:  map[string]*Variable{},
	}
	ret.SetVariable("ok", "1")
	return ret
}

// FS represents the structure which implements the sysstat filesystem.
type FS struct {
	rootDir    *Dir
	Variables  map[string]*Variable
	InodeCount uint64
}

// Root returns the root Node for this file system.
func (fs *FS) Root() (fs.Node, error) {
	if fs.rootDir == nil {
		fs.rootDir = &Dir{
			fs: fs,
		}
	}
	return fs.rootDir, nil
}

func (fs *FS) Inode() uint64 {
	fs.InodeCount++
	return fs.InodeCount
}

func (fs *FS) SetVariable(name, value string) {
	if v, ok := fs.Variables[name]; ok {
		v.Value = value
	}
	fs.Variables[name] = &Variable{
		Inode: fs.Inode(),
		Value: value,
	}
}

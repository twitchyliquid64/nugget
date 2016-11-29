package sysstatfs

import (
	"errors"
	"strconv"
	"time"

	"bazil.org/fuse/fs"
)

// MaxSYSSTATFSInodes is a hard limit on the number of inodes.
const MaxSYSSTATFSInodes = 1000

// ErrNoInodes is returned if the FS is out of inodes.
var ErrNoInodes = errors.New("Out of Inodes")

// Make creates a minimal FS node.
func Make() *FS {
	ret := &FS{
		InodeCount: 1,
		Variables:  map[string]Variable{},
	}
	ret.SetVariable("ok", "1")
	ret.SetComputedVariable("time", func() []byte {
		return []byte(time.Now().String())
	})
	ret.SetComputedVariable("statfs_inodes_used", func() []byte {
		return []byte(strconv.Itoa(int(ret.InodeCount)))
	})
	return ret
}

// FS represents the structure which implements the sysstat filesystem.
type FS struct {
	rootDir    *Dir
	Variables  map[string]Variable
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

// Inode issues an unused Inode, or returns an error.
func (fs *FS) Inode() (uint64, error) {
	fs.InodeCount++
	if fs.InodeCount > MaxSYSSTATFSInodes {
		return 0, ErrNoInodes
	}
	return fs.InodeCount, nil
}

// SetVariable sets a variable to a specific value.
func (fs *FS) SetVariable(name, value string) error {
	if v, ok := fs.Variables[name]; ok {
		if fv, ok := v.(*FixedVariable); ok {
			fv.Value = value
			return nil
		}
		fs.Variables[name] = &FixedVariable{
			Inode: v.GetInode(),
			Value: value,
		}
	}

	inode, err := fs.Inode()
	if err != nil {
		return err
	}
	fs.Variables[name] = &FixedVariable{
		Inode: inode,
		Value: value,
	}
	return nil
}

// SetComputedVariable sets a variable to the output of the given function. The function
// will be called every time the file is read.
func (fs *FS) SetComputedVariable(name string, value func() []byte) error {
	if v, ok := fs.Variables[name]; ok {
		if cv, ok := v.(*ComputedVariable); ok {
			cv.Callback = value
			return nil
		}
		fs.Variables[name] = &ComputedVariable{
			Inode:    v.GetInode(),
			Callback: value,
		}
	}

	inode, err := fs.Inode()
	if err != nil {
		return err
	}
	fs.Variables[name] = &ComputedVariable{
		Inode:    inode,
		Callback: value,
	}
	return nil
}

package sysstatfs

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/twitchyliquid64/nugget/inodeFactory"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// ErrNoInodes is returned if the FS is out of inodes.
var ErrNoInodes = errors.New("Out of Inodes")

// Make creates a minimal FS node.
func Make(inodeSource inodeFactory.InodeFactory) *FS {
	ret := &FS{
		Variables:   map[string]Variable{},
		InodeSource: inodeSource,
	}
	ret.rootInode = ret.InodeSource.GetInode()
	ret.SetVariable("ok", "1")
	ret.SetComputedVariable("time", func() []byte {
		return []byte(time.Now().String())
	})
	ret.SetComputedVariable("statfs_inodes_used", func() []byte {
		return []byte(strconv.Itoa(ret.InodeSource.GetIssued()))
	})
	return ret
}

// FS represents the structure which implements the sysstat filesystem.
type FS struct {
	rootInode   uint64
	lock        sync.Mutex
	Variables   map[string]Variable
	InodeSource inodeFactory.InodeFactory
}

// Root returns the root Node for this file system.
func (fs *FS) Root() (fs.Node, error) {
	return fs, nil
}

// Inode issues an unused Inode, or returns an error.
// Assumes lock is held.
func (fs *FS) Inode() (uint64, error) {
	return fs.InodeSource.GetInode(), nil
}

// SetVariable sets a variable to a specific value.
func (fs *FS) SetVariable(name, value string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
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
	fs.lock.Lock()
	defer fs.lock.Unlock()
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

//Lookup implements fs.NodeRequestLookuper, basically mapping paths to nodes.
func (fs *FS) Lookup(ctx context.Context, name string) (fs.Node, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if v, ok := fs.Variables[name]; ok {
		return v, nil
	}
	return nil, fuse.ENOENT
}

// ReadDirAll implements fs.HandleReadDirAller for listing directories.
func (fs *FS) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	var out []fuse.Dirent
	count := 0
	for name, vari := range fs.Variables {
		out = append(out, fuse.Dirent{Inode: vari.GetInode(), Name: name, Type: fuse.DT_File})
		count++
	}
	return out, nil
}

// Attr implements fs.Node - so this structure represents the readonly directory.
func (fs *FS) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = fs.rootInode
	a.Mode = os.ModeDir | 0555
	return nil
}

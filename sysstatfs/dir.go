package sysstatfs

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// Dir implements both Node and Handle for the root directory - returning system files.
type Dir struct {
	fs *FS
}

// Attr implements fs.Node - so this structure represents the readonly directory.
func (dir *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

//Lookup implements fs.NodeRequestLookuper, basically mapping paths to nodes.
func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	return dir.fs.Lookup(ctx, name)
}

// ReadDirAll implements fs.HandleReadDirAller for listing directories.
func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return dir.fs.ReadDirAll(ctx)
}

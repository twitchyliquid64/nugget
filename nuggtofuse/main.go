package nuggtofuse

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/inodeFactory"
	"github.com/twitchyliquid64/nugget/nuggdb"
)

// FS represents the structure which talks fuse to a nugget.DataSource or nugget.DataSink.
type FS struct {
	rootInode   uint64
	lock        sync.Mutex
	InodeSource inodeFactory.InodeFactory
	provider    nugget.DataSourceSink
}

// Make creates wraps a provider in a structure that can represent a FUSE filesystem.
func Make(provider nugget.DataSourceSink, inodeSource *inodeFactory.PathAwareFactory) *FS {
	r := &FS{
		InodeSource: inodeSource,
		provider:    provider,
	}
	r.rootInode = inodeSource.GetInode()
	return r
}

// Root returns the root Node for this file system.
func (fs *FS) Root() (fs.Node, error) {
	return fs, nil
}

//Lookup implements fs.NodeRequestLookuper, basically mapping paths to nodes.
func (fs *FS) Lookup(ctx context.Context, name string) (fs.Node, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fmt.Fprintln(os.Stderr, "Lookup", name)
	//return v, nil
	return nil, fuse.ENOENT
}

// ReadDirAll implements fs.HandleReadDirAller for listing directories.
func (fs *FS) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fmt.Fprintln(os.Stderr, "ReadDirAll")
	var out []fuse.Dirent
	_, _, data, err := fs.provider.Fetch("/")
	if err == nuggdb.ErrPathNotFound {
		return out, nil
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return out, fuse.EIO
	}
	for _, entry := range strings.Split(string(data), "\n") { //TODO: Refactor the fetch n split to occur in the data source/sink
		out = append(out, fuse.Dirent{Inode: uint64(fs.getInode(entry)), Name: entry, Type: fuse.DT_File})
	}
	return out, nil
}

func (fs *FS) getInode(path string) uint64 {
	pathInodeFactory, ok := fs.InodeSource.(*inodeFactory.PathAwareFactory)
	if ok {
		return pathInodeFactory.GetByPath(path)
	}
	return fs.InodeSource.GetInode()
}

// Attr implements fs.Node - so this structure represents the root directory.
func (fs *FS) Attr(ctx context.Context, a *fuse.Attr) error {
	//fmt.Fprintln(os.Stderr, "Attr")
	a.Inode = fs.rootInode
	a.Mode = os.ModeDir | 0555 //TODO: Make not read only
	return nil
}

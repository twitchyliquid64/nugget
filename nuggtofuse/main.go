package nuggtofuse

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/inodeFactory"
	"github.com/twitchyliquid64/nugget/logger"
	"github.com/twitchyliquid64/nugget/nuggdb"
)

// FS represents the structure which talks fuse to a nugget.DataSource or nugget.DataSink.
type FS struct {
	rootInode   uint64
	lock        sync.Mutex
	InodeSource inodeFactory.InodeFactory
	provider    nugget.DataSourceSink
	logger      *logger.Logger
}

// Make creates wraps a provider in a structure that can represent a FUSE filesystem.
func Make(provider nugget.DataSourceSink, inodeSource *inodeFactory.PathAwareFactory, l *logger.Logger) *FS {
	r := &FS{
		InodeSource: inodeSource,
		provider:    provider,
		logger:      l,
	}
	r.rootInode = inodeSource.GetInode()
	return r
}

func (fs *FS) getFile(fullPath string) *File {
	return &File{
		fs:       fs,
		inode:    fs.getInode(fullPath),
		fullPath: fullPath,
	}
}

// Root returns the root Node for this file system.
func (fs *FS) Root() (fs.Node, error) {
	return fs, nil
}

//Lookup implements fs.NodeRequestLookuper, basically mapping paths to nodes.
func (fs *FS) Lookup(ctx context.Context, name string) (fs.Node, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.logger.Info("fuse-lookup", "Query for: ", name)
	_, err := fs.provider.Lookup("/" + name)
	if err == nuggdb.ErrPathNotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "Err:", err)
		return nil, fuse.EIO
	}

	return fs.getFile("/" + name), nil
}

// ReadDirAll implements fs.HandleReadDirAller for listing directories.
func (fs *FS) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.logger.Info("fuse-readdirall", "Got root request")

	var out []fuse.Dirent
	_, _, data, err := fs.provider.Fetch("/")
	if err == nuggdb.ErrPathNotFound {
		return out, nil
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return out, fuse.EIO
	}
	for _, entry := range strings.Split(string(data), "\n") { //TODO: Refactor the fetch n split to occur in the data source/sink
		if entry != "" {
			out = append(out, fuse.Dirent{Inode: uint64(fs.getInode(entry)), Name: path.Base(entry), Type: fuse.DT_File})
		}
	}
	fmt.Println(out)
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
	fs.logger.Info("fuse-attr", "Got root request")
	a.Inode = fs.rootInode
	a.Mode = os.ModeDir | 0777
	return nil
}

// Create implements fs.NodeCreater. It is called to create and open a new
// file. The kernel will first try to Lookup the name, and this method will only be called
// if the name didn't exist.
func (fs *FS) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if strings.Contains(req.Name, "/") {
		fs.logger.Error("fuse-create", "Cannot create node which contains slashes: ", req.Name)
		return nil, nil, fuse.EPERM
	}
	fs.logger.Info("fuse-create", "Name: ", req.Name)
	fs.provider.Store("/"+req.Name, []byte{})
	f := fs.getFile("/" + req.Name)
	return f, f, fuse.EIO
}

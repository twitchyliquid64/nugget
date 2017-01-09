package nuggtofuse

import (
	"context"
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
	overrides   map[string]fs.Node
}

// Make creates wraps a provider in a structure that can represent a FUSE filesystem.
func Make(provider nugget.DataSourceSink, inodeSource *inodeFactory.PathAwareFactory, l *logger.Logger) *FS {
	r := &FS{
		InodeSource: inodeSource,
		provider:    provider,
		logger:      l,
		overrides:   map[string]fs.Node{},
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

func (fs *FS) getDir(fullPath string) *Dir {
	return &Dir{
		fs:       fs,
		inode:    fs.getInode(fullPath),
		fullPath: fullPath,
	}
}

// SetOverride allows you to add another directory to the root of the filesystem, which will be exposed via FUSE.
func (fs *FS) SetOverride(name string, override fs.Node) {
	fs.lock.Lock()
	fs.lock.Unlock()
	fs.overrides[name] = override
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

	if override, overrideExists := fs.overrides[name]; overrideExists {
		return override, nil
	}

	eID, err := fs.provider.Lookup("/" + name)
	if err == nuggdb.ErrPathNotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
		fs.logger.Error("fuse-lookup", "Lookup for "+name+" failed: ", err)
		return nil, fuse.EIO
	}

	meta, err := fs.provider.ReadMeta(eID)
	if err != nil {
		fs.logger.Error("fuse-lookup", "ReadMeta for "+name+" failed: ", err)
		return nil, fuse.EIO
	}

	if meta.IsDirectory() {
		return fs.getDir("/" + name), nil
	}
	return fs.getFile("/" + name), nil
}

// ReadDirAll implements fs.HandleReadDirAller for listing directories.
func (fs *FS) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	fs.logger.Info("fuse-readdirall", "Got root request")

	var out []fuse.Dirent
	entries, err := fs.provider.List("/")
	if err == nuggdb.ErrPathNotFound {
		return out, nil
	} else if err != nil {
		fs.logger.Error("fuse-readdirall", "provider.List(/) Failed: ", err)
		return out, fuse.EIO
	}
	for _, entry := range entries {
		if entry.IsDirectory() {
			out = append(out, fuse.Dirent{Inode: uint64(fs.getInode(entry.Identifier())), Name: path.Base(entry.Identifier()), Type: fuse.DT_Dir})
		} else {
			out = append(out, fuse.Dirent{Inode: uint64(fs.getInode(entry.Identifier())), Name: path.Base(entry.Identifier()), Type: fuse.DT_File})
		}
	}

	for name := range fs.overrides {
		out = append(out, fuse.Dirent{Name: name, Type: fuse.DT_Dir})
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
	return f, f, nil
}

// Mkdir implements the NodeMkdirer interface. It is called to make a new directory.
func (fs *FS) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	fs.logger.Info("fuse-mkdir", "Got root request for: ", req.Name)
	if strings.Contains(req.Name, "/") {
		fs.logger.Error("fuse-mkdir", "Cannot create node which contains slashes: ", req.Name)
		return nil, fuse.EPERM
	}

	_, _, err := fs.provider.Mkdir("/" + req.Name)
	if err == nil {
		return fs.getDir("/" + req.Name), nil
	}
	fs.logger.Error("fuse-mkdir", "provider.Mkdir(/"+req.Name+") failed: ", err)
	return nil, fuse.EIO
}

// Remove implements NodeRemover, which allows the removal of files.
func (fs *FS) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	fs.logger.Info("fuse-remove", "Got root request for: ", req.Name)
	if strings.Contains(req.Name, "/") {
		fs.logger.Error("fuse-remove", "Cannot remove node which contains slashes: ", req.Name)
		return fuse.EPERM
	}

	err := fs.provider.Delete("/" + req.Name)
	if err != nil {
		fs.logger.Error("fuse-remove", "provider.Delete(/"+req.Name+") Failed: ", err)
		return fuse.EIO
	}
	return nil
}

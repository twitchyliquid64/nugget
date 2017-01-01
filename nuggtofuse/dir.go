package nuggtofuse

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/twitchyliquid64/nugget/nuggdb"
)

// Dir represents is a FUSE wrapper around a directory entity stored in the system.
// Dir MUST exist.
type Dir struct {
	fs       *FS
	fullPath string
	inode    uint64
}

// Attr implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	d.fs.logger.Info("fuse-attr", "Got request for ", d.fullPath)
	a.Inode = d.inode
	a.Mode = os.ModeDir | 0777
	return nil
}

// ReadDirAll implements fs.HandleReadDirAller for listing directories.
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.fs.logger.Info("fuse-readdirall", "Got request on ", d.fullPath)

	var out []fuse.Dirent
	entries, err := d.fs.provider.List(d.fullPath)
	if err == nuggdb.ErrPathNotFound {
		return out, nil
	} else if err != nil {
		d.fs.logger.Error("fuse-readdirall", "provider.List("+d.fullPath+") Failed: ", err)
		return out, fuse.EIO
	}
	for _, entry := range entries {
		fmt.Println(entry.Identifier())
		if entry.IsDirectory() {
			out = append(out, fuse.Dirent{Inode: uint64(d.fs.getInode(entry.Identifier())), Name: path.Base(entry.Identifier()), Type: fuse.DT_Dir})
		} else {
			out = append(out, fuse.Dirent{Inode: uint64(d.fs.getInode(entry.Identifier())), Name: path.Base(entry.Identifier()), Type: fuse.DT_File})
		}
	}
	fmt.Println(out)
	return out, nil
}

//Lookup implements fs.NodeRequestLookuper, basically mapping paths to nodes.
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.fs.logger.Info("fuse-lookup", "Query for: ", path.Join(d.fullPath, name))
	eID, err := d.fs.provider.Lookup(path.Join(d.fullPath, name))
	if err == nuggdb.ErrPathNotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
		d.fs.logger.Error("fuse-lookup", "Lookup for "+path.Join(d.fullPath, name)+" failed: ", err)
		return nil, fuse.EIO
	}

	meta, err := d.fs.provider.ReadMeta(eID)
	if err != nil {
		d.fs.logger.Error("fs-lookup", "ReadMeta for "+path.Join(d.fullPath, name)+" failed: ", err)
		return nil, fuse.EIO
	}

	if meta.IsDirectory() {
		return d.fs.getDir(path.Join(d.fullPath, name)), nil
	}
	return d.fs.getFile(path.Join(d.fullPath, name)), nil
}

// Create implements fs.NodeCreater. It is called to create and open a new
// file. The kernel will first try to Lookup the name, and this method will only be called
// if the name didn't exist.
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if strings.Contains(req.Name, "/") {
		d.fs.logger.Error("fuse-create", "Cannot create node which contains slashes: ", req.Name)
		return nil, nil, fuse.EPERM
	}
	d.fs.logger.Info("fuse-create", "Name: ", path.Join(d.fullPath, req.Name))
	d.fs.provider.Store(path.Join(d.fullPath, req.Name), []byte{})
	f := d.fs.getFile(path.Join(d.fullPath, req.Name))
	return f, f, fuse.EIO
}

// Mkdir implements the NodeMkdirer interface. It is called to make a new directory.
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	d.fs.logger.Info("fuse-mkdir", "Got request for: ", path.Join(d.fullPath, req.Name))
	if strings.Contains(req.Name, "/") {
		d.fs.logger.Error("fuse-mkdir", "Cannot create node which contains slashes: ", req.Name)
		return nil, fuse.EPERM
	}

	_, _, err := d.fs.provider.Mkdir(path.Join(d.fullPath, req.Name))
	if err == nil {
		return d.fs.getDir(path.Join(d.fullPath, req.Name)), nil
	}
	d.fs.logger.Error("fuse-mkdir", "provider.Mkdir("+path.Join(d.fullPath, req.Name)+") failed: ", err)
	return nil, fuse.EIO
}

// Remove implements NodeRemover, which allows the removal of files.
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	d.fs.logger.Info("fuse-remove", "Got request for: ", path.Join(d.fullPath, req.Name))
	if strings.Contains(req.Name, "/") {
		d.fs.logger.Error("fuse-remove", "Cannot remove node which contains slashes: ", req.Name)
		return fuse.EPERM
	}

	err := d.fs.provider.Delete(path.Join(d.fullPath, req.Name))
	if err != nil {
		d.fs.logger.Error("fuse-remove", "provider.Delete("+path.Join(d.fullPath, req.Name)+") Failed: ", err)
		return fuse.EIO
	}
	return nil
}

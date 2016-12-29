package nuggtofuse

import (
	"context"

	"bazil.org/fuse"
)

// File represents is a FUSE wrapper around a file entity stored in the system.
type File struct {
	fs       *FS
	fullPath string
	inode    uint64
}

// Attr implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.inode
	a.Mode = 0444

	entryID, err := f.fs.provider.Lookup(f.fullPath)
	if err != nil {
		return err
	}

	meta, err := f.fs.provider.ReadMeta(entryID)
	if err != nil {
		return err
	}
	a.Size = meta.GetSize()

	return nil
}

// ReadAll implements fs.HandleReadAller, allowing file reads.
func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	_, _, data, err := f.fs.provider.Fetch(f.fullPath)
	return data, err
}

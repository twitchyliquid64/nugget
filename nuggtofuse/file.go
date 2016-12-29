package nuggtofuse

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// File represents is a FUSE wrapper around a file entity stored in the system.
// File MUST exist.
type File struct {
	fs       *FS
	fullPath string
	inode    uint64
}

// Open implements  fuse.NodeOpener. It is called the first time a file is opened
// by any process. Further opens or FD duplications will reuse this handle. When all FDs
// have been closed, Release() will be called.
func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return f, nil
}

// Attr implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	f.fs.logger.Info("fuse-attr", "Got request for ", f.fullPath)
	a.Inode = f.inode
	a.Mode = os.ModePerm | 0777

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
	f.fs.logger.Info("fuse-readall", "Got request for ", f.fullPath)
	_, _, data, err := f.fs.provider.Fetch(f.fullPath)
	return data, err
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.fs.logger.Info("fuse-write", "Got request for ", f.fullPath)

	//TODO: Have a write interface that passes it through to the server, where it can do the splicing.
	_, _, data, err := f.fs.provider.Fetch(f.fullPath)
	if err != nil {
		f.fs.logger.Error("fuse-write", "Failed fetch operation: ", err)
		return fuse.EIO
	}

	newData := doWrite(req.Offset, req.Data, data)
	_, _, err = f.fs.provider.Store(f.fullPath, newData)
	if err != nil {
		f.fs.logger.Error("fuse-write", "Failed store operation: ", err)
		return fuse.EIO
	}
	return nil
}

// doWrite does the buffer manipulation to perform a write. Data buffers are kept
// contiguous.
// Credit: bwester (consulfs)
func doWrite(offset int64, writeData []byte, fileData []byte) []byte {
	fileEnd := int64(len(fileData))
	writeEnd := offset + int64(len(writeData))
	var buf []byte
	if writeEnd > fileEnd {
		buf = make([]byte, writeEnd)
		if fileEnd <= offset {
			copy(buf, fileData)
		} else {
			copy(buf, fileData[:offset])
		}
	} else {
		buf = fileData
	}
	copy(buf[offset:writeEnd], writeData)
	return buf
}

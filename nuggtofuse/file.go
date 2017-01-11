package nuggtofuse

import (
	"context"

	"github.com/twitchyliquid64/nugget"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
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
	a.Mode = 0777

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

// Read implements fs.HandleRead, allowing file reads.
func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	f.fs.logger.Info("fuse-read", "Got request for ", f.fullPath, " with size=", req.Size, " and offset=", req.Offset)

	if optimizedProvider, ok := f.fs.provider.(nugget.OptimisedDataSourceSink); ok {
		data, err := optimizedProvider.Read(f.fullPath, req.Offset, int64(req.Size))
		if err != nil {
			f.fs.logger.Error("fuse-read", "Failed Read operation: ", err)
			return fuse.EIO
		}
		resp.Data = data
		return nil
	}

	f.fs.logger.Warning("fuse-read", "Provider is not optimized, falling back to Fetch/slice strategy.")
	_, _, data, err := f.fs.provider.Fetch(f.fullPath)
	fuseutil.HandleRead(req, resp, data)
	return err
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.fs.logger.Info("fuse-write", "Got request for ", f.fullPath)

	if optimizedProvider, ok := f.fs.provider.(nugget.OptimisedDataSourceSink); ok {
		written, _, _, err := optimizedProvider.Write(f.fullPath, req.Offset, req.Data)
		if err != nil {
			f.fs.logger.Error("fuse-write", "Failed Write operation: ", err)
			return fuse.EIO
		}
		resp.Size = int(written)
		return nil
	}

	f.fs.logger.Warning("fuse-write", "Provider is not optimized, falling back to Fetch/store strategy.")
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
	resp.Size = len(req.Data)
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

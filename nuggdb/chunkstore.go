package nuggdb

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/twitchyliquid64/nugget"
)

const chunkBucket = "ChunkIDToData"

// ErrChunkNotFound is returned if the chunk requested was not found in the chunkstore.
var ErrChunkNotFound = errors.New("Could not find chunk in chunkstore")

// Chunkstore is the concrete instance responsible
// for storing / fetching chunks by chunk ID.
type Chunkstore struct {
	path string
}

// OpenChunkStore opens a chunkstore backed by the file at path.
func OpenChunkStore(path string) (*Chunkstore, error) {
	if !fileExists(path) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			return nil, err
		}
	}

	chunkStore := &Chunkstore{
		path: path,
	}
	return chunkStore, nil
}

// Lookup returns the data associated with a chunkID. ErrChunkNotFound
// is returned if no such chunk exists.
func (cs *Chunkstore) Lookup(chunkID nugget.ChunkID) ([]byte, error) {
	fPath := path.Join(cs.path, cs.dirPrefix(chunkID), cs.fileName(chunkID))

	if !fileExists(fPath) {
		return []byte(""), ErrChunkNotFound
	}

	return ioutil.ReadFile(fPath)
}

func (cs *Chunkstore) Write(chunkID nugget.ChunkID, offset int64, data []byte) (int, int64, error) {
	fPath := path.Join(cs.path, cs.dirPrefix(chunkID), cs.fileName(chunkID))
	fHandle, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, ErrChunkNotFound
		}
		return 0, 0, err
	}
	defer fHandle.Close()
	written, err := fHandle.WriteAt(data, offset)
	if err != nil {
		return written, 0, err
	}
	stat, err := fHandle.Stat()
	if err != nil {
		return written, 0, err
	}
	return written, stat.Size(), err
}

func (cs *Chunkstore) Read(chunkID nugget.ChunkID, offset int64, size int64) ([]byte, error) {
	fPath := path.Join(cs.path, cs.dirPrefix(chunkID), cs.fileName(chunkID))
	fHandle, err := os.Open(fPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte(""), ErrChunkNotFound
		}
		return []byte(""), err
	}
	defer fHandle.Close()
	buff := make([]byte, size)
	n, err := fHandle.ReadAt(buff, offset)
	return buff[:n], err
}

func (cs *Chunkstore) dirPrefix(chunkID nugget.ChunkID) string {
	return hex.EncodeToString(chunkID[:2])
}

func (cs *Chunkstore) fileName(chunkID nugget.ChunkID) string {
	return hex.EncodeToString(chunkID[2:])
}

// Commit sets the data for a chunkID.
func (cs *Chunkstore) Commit(chunkID nugget.ChunkID, data []byte) error {
	dirPath := path.Join(cs.path, cs.dirPrefix(chunkID))

	if !fileExists(dirPath) {
		err := os.Mkdir(dirPath, 0777)
		if err != nil {
			return err
		}
	}

	fPath := path.Join(cs.path, cs.dirPrefix(chunkID), cs.fileName(chunkID))
	return ioutil.WriteFile(fPath, data, 0755)
}

// Forge saves data with a new chunk ID and returns it.
func (cs *Chunkstore) Forge(data []byte) (nugget.ChunkID, error) {
	var id nugget.ChunkID
	rand.Read(id[:])

	return id, cs.Commit(id, data)
}

// Close closes the underlying database. This should be called before shutdown.
func (cs *Chunkstore) Close() error {
	return nil
}

// Delete removes a chunk from the chunkstore. Nil is returned if the chunk does not exist.
func (cs *Chunkstore) Delete(chunkID nugget.ChunkID) error {
	fPath := path.Join(cs.path, cs.dirPrefix(chunkID), cs.fileName(chunkID))
	return os.Remove(fPath)
}

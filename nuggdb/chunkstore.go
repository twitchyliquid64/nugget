package nuggdb

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/twitchyliquid64/nugget"
)

const chunkBucket = "ChunkIDToData"

// ErrChunkNotFound is returned if the chunk requested was not found in the chunkstore.
var ErrChunkNotFound = errors.New("Could not find chunk in chunkstore")

// Chunkstore is the concrete instance responsible
// for storing / fetching chunks by chunk ID.
type Chunkstore struct {
	path string
	db   *bolt.DB
}

// OpenChunkStore opens a chunkstore backed by the file at path.
func OpenChunkStore(path string) (*Chunkstore, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists([]byte(chunkBucket))
		return err2
	})
	if err != nil {
		return nil, err
	}

	chunkStore := &Chunkstore{
		path: path,
		db:   db,
	}
	return chunkStore, nil
}

// Lookup returns the data associated with a chunkID. ErrChunkNotFound
// is returned if no such chunk exists.
func (cs *Chunkstore) Lookup(chunkID nugget.ChunkID) ([]byte, error) {
	var result []byte
	var notFound bool
	err := cs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(chunkBucket))
		v := b.Get([]byte(chunkID[:]))
		if v != nil {
			result = make([]byte, len(v))
			copy(result, v)
		} else {
			notFound = true
		}
		return nil
	})
	if notFound {
		return result, ErrChunkNotFound
	}
	return result, err
}

// Commit sets the data for a chunkID.
func (cs *Chunkstore) Commit(chunkID nugget.ChunkID, data []byte) error {
	return cs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(chunkBucket))
		err := b.Put([]byte(chunkID[:]), data)
		return err
	})
}

// Forge saves data with a new chunk ID and returns it.
func (cs *Chunkstore) Forge(data []byte) (nugget.ChunkID, error) {
	var id nugget.ChunkID
	rand.Read(id[:])

	return id, cs.Commit(id, data)
}

// Close closes the underlying database. This should be called before shutdown.
func (cs *Chunkstore) Close() error {
	return cs.db.Close()
}

// Delete removes a chunk from the chunkstore. Nil is returned if the chunk does not exist.
func (cs *Chunkstore) Delete(chunkID nugget.ChunkID) error {
	return cs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(chunkBucket))
		err := b.Delete(chunkID[:])
		return err
	})
}

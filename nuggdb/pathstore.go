package nuggdb

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/twitchyliquid64/nugget"
)

const pathEntryIDBucket = "PathToEntryID"

// ErrPathNotFound is returned if the path requested was not found in the pathstore.
var ErrPathNotFound = errors.New("Could not find path in pathstore")

// Pathstore is the concrete instance responsible
// for storing / fetching the mapping between paths
// and entity IDs.
type Pathstore struct {
	path string
	db   *bolt.DB
}

// OpenPathStore opens a pathstore backed by the file at path.
func OpenPathStore(path string) (*Pathstore, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists([]byte(pathEntryIDBucket))
		return err2
	})
	if err != nil {
		return nil, err
	}

	pathstore := &Pathstore{
		path: path,
		db:   db,
	}
	return pathstore, nil
}

// Lookup finds a entryID which is mapped to a given path. ErrPathNotFound
// is returned if no such mapping exists.
func (ps *Pathstore) Lookup(path string) (nugget.EntryID, error) {
	var result nugget.EntryID
	var notFound bool
	err := ps.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pathEntryIDBucket))
		v := b.Get([]byte(path))
		if v != nil {
			copy(result[:], v)
		} else {
			notFound = true
		}
		return nil
	})
	if notFound {
		return result, ErrPathNotFound
	}
	return result, err
}

// Commit sets the entryID for path.
func (ps *Pathstore) Commit(path string, entryID nugget.EntryID) error {
	return ps.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(pathEntryIDBucket))
		err := b.Put([]byte(path), []byte(entryID[:]))
		return err
	})
}

// Forge commits a random EntryID for path, and returns it.
func (ps *Pathstore) Forge(path string) (nugget.EntryID, error) {
	var id nugget.EntryID
	rand.Read(id[:])

	return id, ps.Commit(path, id)
}

// Close closes the underlying database. This should be called before shutdown.
func (ps *Pathstore) Close() error {
	return ps.db.Close()
}

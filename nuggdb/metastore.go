package nuggdb

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/twitchyliquid64/nugget"
)

const entryIDToMetaBucket = "EntryIDToMeta"

// ErrMetaNotFound is returned if the entryID requested was not found in the metastore.
var ErrMetaNotFound = errors.New("Could not find entry in metastore")

// Metastore is the concrete instance responsible
// for storing / fetching the mapping between EntryIDs
// and EntryMetadata's.
type Metastore struct {
	path string
	db   *bolt.DB
}

// OpenMetaStore opens a metastore backed by the file at path.
func OpenMetaStore(path string) (*Metastore, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err2 := tx.CreateBucketIfNotExists([]byte(entryIDToMetaBucket))
		return err2
	})
	if err != nil {
		return nil, err
	}

	metastore := &Metastore{
		path: path,
		db:   db,
	}
	return metastore, nil
}

// Lookup finds a Metadata entry which is mapped to a EntryID. ErrMetaNotFound
// is returned if no such mapping exists.
func (ps *Metastore) Lookup(entryID nugget.EntryID) (EntryMetadata, error) {
	var result EntryMetadata
	var notFound bool
	err := ps.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(entryIDToMetaBucket))
		v := b.Get([]byte(entryID[:]))
		if v != nil {
			result = MakeMetadata(v)
		} else {
			notFound = true
		}
		return nil
	})
	if notFound {
		return result, ErrMetaNotFound
	}
	return result, err
}

// Commit sets the Meta for a given EntryID.
func (ps *Metastore) Commit(meta EntryMetadata) error {
	return ps.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(entryIDToMetaBucket))
		err := b.Put([]byte(meta.EntryID[:]), meta.Serialize())
		return err
	})
}

// Close closes the underlying database. This should be called before shutdown.
func (ps *Metastore) Close() error {
	return ps.db.Close()
}

// Delete removes a metadata entry from the metastore. Nil is returned if the entryID is not mapped.
func (ps *Metastore) Delete(entryID nugget.EntryID) error {
	return ps.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(entryIDToMetaBucket))
		err := b.Delete(entryID[:])
		return err
	})
}

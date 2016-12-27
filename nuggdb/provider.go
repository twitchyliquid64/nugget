package nuggdb

import (
	"crypto/rand"
	"errors"
	"os"
	"path"

	"github.com/twitchyliquid64/nugget"
)

const (
	pathStoreFilename  = "paths.db"
	metaStoreFilename  = "meta.db"
	chunkStoreFilename = "data.db"
)

// Provider represents a nugget database, reading and storing file information backed by boltDB databases.
type Provider struct {
	pathstore  *Pathstore
	metastore  *Metastore
	chunkstore *Chunkstore
	basedir    string
}

// Create initializes the backend of a nugget filesystem, returning an object that implements
// nugget.DataSource & nugget.DataSink.
func Create(baseDir string) (*Provider, error) {
	var err error
	ret := &Provider{
		basedir: baseDir,
	}
	if !fileExists(baseDir) {
		return nil, errors.New("Could not stat base directory")
	}
	ret.pathstore, err = OpenPathStore(path.Join(baseDir, pathStoreFilename))
	if err != nil {
		return nil, err
	}
	ret.metastore, err = OpenMetaStore(path.Join(baseDir, metaStoreFilename))
	if err != nil {
		return nil, err
	}
	ret.chunkstore, err = OpenChunkStore(path.Join(baseDir, chunkStoreFilename))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Lookup looks up a specific path, returning the EntryID of the path if one exists.
func (p *Provider) Lookup(path string) (nugget.EntryID, error) {
	return p.pathstore.Lookup(path)
}

// ReadMeta returns the metadata for a given entryID.
func (p *Provider) ReadMeta(entry nugget.EntryID) (nugget.NodeMetadata, error) {
	meta, err := p.metastore.Lookup(entry)
	return &meta, err
}

//ReadData returns the data stored at the given chunkID
func (p *Provider) ReadData(chunkID nugget.ChunkID) ([]byte, error) {
	return p.chunkstore.Lookup(chunkID)
}

//Store completely overwrites a file at fPath.
func (p *Provider) Store(fPath string, data []byte) (nugget.EntryID, nugget.NodeMetadata, error) {
	existingEntryID, pathSearchError := p.Lookup(fPath)
	if pathSearchError != nil && pathSearchError != ErrPathNotFound {
		return existingEntryID, nil, pathSearchError
	} // return an error for all path errors, except where file does not exist.

	//commit data first
	chunkID, chunkWriteError := p.chunkstore.Forge(data)
	if chunkWriteError != nil {
		return existingEntryID, nil, chunkWriteError
	} //return error if we could not write the raw data

	//now we make a brand new metadata entry
	var newEntryID nugget.EntryID
	rand.Read(newEntryID[:])
	meta := EntryMetadata{
		EntryID: newEntryID,
		IsDir:   false,
		Lname:   path.Base(fPath),
		Size:    uint64(len(data)),
		Locality: LocalityInfo{
			ChunkID: chunkID,
		},
	}

	metaWriteError := p.metastore.Commit(meta)
	if metaWriteError != nil {
		p.chunkstore.Delete(chunkID) // Undo our only change - new chunk
		return existingEntryID, &meta, metaWriteError
	}

	pathWriteError := p.pathstore.Commit(fPath, newEntryID)
	if pathWriteError != nil {
		p.chunkstore.Delete(chunkID)   // Undo our changes: new chunk
		p.metastore.Delete(newEntryID) // Undo our changes: new Entry
		//TODO: Report leaks for errors when rolling back
	}

	if pathSearchError == nil { //path already exists, need to delete crap
		swapErr := p.deleteGracefullyOrRollbackNewFile(fPath, chunkID, existingEntryID, &meta)
		if swapErr != nil {
			return newEntryID, nil, swapErr
		}
	}

	return newEntryID, &meta, pathWriteError
}

func (p *Provider) deleteGracefullyOrRollbackNewFile(fPath string, newChunkID nugget.ChunkID, oldEntryID nugget.EntryID, newMeta *EntryMetadata) error {
	abort := func() {
		p.chunkstore.Delete(newChunkID)       // Undo our changes: new chunk
		p.metastore.Delete(newMeta.ID())      // Undo our changes: new Entry
		p.pathstore.Commit(fPath, oldEntryID) // Undo our changes: write back pointer to old entryMetadata
		//TODO: Report or handle errors when rolling back
	}
	oldMeta, err := p.metastore.Lookup(oldEntryID)
	if err != nil {
		abort()
		return err
	}
	err = p.metastore.Delete(oldEntryID)
	if err != nil {
		abort()
		return err
	}
	err = p.chunkstore.Delete(oldMeta.GetDataLocality().Chunks()[0])
	if err != nil {
		p.metastore.Commit(oldMeta) //Undo our delete: write back old MetaEntry
		abort()
		return err
	}
	return nil
}

// Close closes all underlying files and makes the provider unusable.
func (p *Provider) Close() error {
	e := p.pathstore.Close()
	if e != nil {
		p.metastore.Close()
		p.chunkstore.Close()
		return e
	}
	e = p.metastore.Close()
	if e != nil {
		p.chunkstore.Close()
		return e
	}
	return p.chunkstore.Close()
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
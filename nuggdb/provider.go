package nuggdb

// nuggdb implements the data storage backend for nugget/nuggFS.
// put simply, it takes our programming interface and stores it to
// disk using boltdb (key-value store).

import (
	"crypto/rand"
	"errors"
	"os"
	"path"

	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/logger"
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
func Create(baseDir string, l *logger.Logger) (*Provider, error) {
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

// Mkdir creates and commits a new directory, returning the entryID and metadata of the directory file.
func (p *Provider) Mkdir(fPath string) (nugget.EntryID, nugget.NodeMetadata, error) {
	eID, meta, newFile, err := p.store(fPath, dirEntries{}.Serialize(), true)
	if err == nil && newFile {
		err = p.appendDirectoryEntry(fPath, true)
	}
	return eID, meta, err
}

//Delete deletes a file or directory.
func (p *Provider) Delete(fPath string) error {
	eID, errLookup := p.Lookup(fPath)
	if errLookup != nil {
		return errLookup
	}

	meta, errMeta := p.ReadMeta(eID)
	if errMeta != nil {
		return errMeta
	}

	//Remove the link first, we can always put it back if we need to
	pathErr := p.pathstore.Delete(fPath)
	if pathErr != nil {
		return pathErr //failure without affecting consistency
	}

	//Remove the metadata next, we can always put it back if we need to
	entryIDErr := p.metastore.Delete(eID)
	if entryIDErr != nil {
		p.pathstore.Commit(fPath, eID) //Rollback pathstore deletion
		return entryIDErr              //failure without affecting consistency
	}

	//Delete the data, if this works we are done
	chunkErr := p.chunkstore.Delete(meta.GetDataLocality().Chunks()[0])
	if chunkErr != nil {
		p.metastore.Commit(EntryMetadata{
			EntryID:  eID,
			IsDir:    meta.IsDirectory(),
			Lname:    meta.LocalName(),
			Size:     meta.GetSize(),
			Locality: LocalityInfo{ChunkID: meta.GetDataLocality().Chunks()[0]},
		}) //Rollback
		p.pathstore.Commit(fPath, eID) //Rollback
		return chunkErr                //failure without affecting consistency
	}
	return p.removeDirectoryEntry(fPath)
}

func (p *Provider) removeDirectoryEntry(fPath string) error {
	dirPath := path.Dir(fPath)
	_, meta, data, err := p.Fetch(dirPath)
	if err != nil && err != ErrPathNotFound {
		return err
	} else if err == nil {
		if !meta.IsDirectory() {
			return errors.New("Cannot make file on top of a non-directory path")
		}
	}

	var entries []DirEntry
	if len(data) > 0 {
		entries, err = deserializeDirEntries(data)
		if err != nil {
			return err
		}
	}

	var outEntries []DirEntry
	for _, entry := range entries {
		if entry.Name != fPath {
			outEntries = append(outEntries, entry)
		}
	}

	_, _, _, err = p.store(dirPath, dirEntries(outEntries).Serialize(), true)
	return err
}

func (p *Provider) appendDirectoryEntry(fPath string, isDir bool) error {
	dirPath := path.Dir(fPath)
	_, meta, data, err := p.Fetch(dirPath)
	if err != nil && err != ErrPathNotFound {
		return err
	} else if err == nil {
		if !meta.IsDirectory() {
			return errors.New("Cannot make file on top of a non-directory path")
		}
	}

	var entries []DirEntry
	if len(data) > 0 {
		entries, err = deserializeDirEntries(data)
		if err != nil {
			return err
		}
	}

	for _, entry := range entries {
		if entry.Name == fPath {
			return nil
		}
	}
	//doesnt exist, add it
	entries = append(entries, DirEntry{Name: fPath, IsDir: isDir})
	_, _, _, err = p.store(dirPath, dirEntries(entries).Serialize(), true)
	return err
}

//Store completely overwrites a file at fPath.
func (p *Provider) Store(fPath string, data []byte) (nugget.EntryID, nugget.NodeMetadata, error) {
	eID, meta, newFile, err := p.store(fPath, data, false)
	if err == nil && newFile {
		err = p.appendDirectoryEntry(fPath, false)
	}
	return eID, meta, err
}

func (p *Provider) Write(fPath string, offset int64, data []byte) (int64, nugget.EntryID, nugget.NodeMetadata, error) {
	eID, meta, existingData, err := p.Fetch(fPath)
	if err != nil {
		return 0, eID, meta, err
	}

	newData := doWrite(offset, data, existingData)
	eID, meta, err = p.Store(fPath, newData)
	return int64(len(data)), eID, meta, err
}

func (p *Provider) Read(fPath string, offset int64, size int64) ([]byte, error) {
	_, _, data, err := p.Fetch(fPath)
	if err != nil {
		return []byte(""), err
	}

	if offset > int64(len(data)) {
		data = nil
	} else {
		data = data[offset:]
	}
	if int64(len(data)) > size {
		data = data[:size]
	}
	return data, nil
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

// Fetch returns the full tree of information about a file.
func (p *Provider) Fetch(fPath string) (eID nugget.EntryID, meta nugget.NodeMetadata, data []byte, err error) {
	eID, err = p.Lookup(fPath)
	if err != nil {
		return
	}
	meta, err = p.ReadMeta(eID)
	if err != nil {
		return
	}
	data, err = p.ReadData(meta.GetDataLocality().Chunks()[0])
	return
}

func (p *Provider) store(fPath string, data []byte, isDir bool) (nugget.EntryID, nugget.NodeMetadata, bool, error) {
	existingEntryID, pathSearchError := p.Lookup(fPath)
	if pathSearchError != nil && pathSearchError != ErrPathNotFound {
		return existingEntryID, nil, false, pathSearchError
	} // return an error for all path errors, except where file does not exist.

	//commit data first
	chunkID, chunkWriteError := p.chunkstore.Forge(data)
	if chunkWriteError != nil {
		return existingEntryID, nil, false, chunkWriteError
	} //return error if we could not write the raw data

	//now we make a brand new metadata entry
	var newEntryID nugget.EntryID
	rand.Read(newEntryID[:])
	meta := EntryMetadata{
		EntryID: newEntryID,
		IsDir:   isDir,
		Lname:   path.Base(fPath),
		Size:    uint64(len(data)),
		Locality: LocalityInfo{
			ChunkID: chunkID,
		},
	}

	metaWriteError := p.metastore.Commit(meta)
	if metaWriteError != nil {
		p.chunkstore.Delete(chunkID) // Undo our only change - new chunk
		return existingEntryID, &meta, false, metaWriteError
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
			return newEntryID, nil, false, swapErr
		}
	}

	return newEntryID, &meta, pathSearchError == ErrPathNotFound, pathWriteError
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

// List returns information about the directory at fPath.
func (p *Provider) List(fPath string) ([]nugget.DirEntry, error) {
	_, meta, data, err := p.Fetch(fPath)
	if err != nil {
		return nil, err
	}
	if !meta.IsDirectory() {
		return nil, errors.New("Cannot List on non-directory file")
	}
	entries, err := deserializeDirEntries(data)
	if err != nil {
		return nil, err
	}

	b := make([]nugget.DirEntry, len(entries))
	for i := range entries {
		b[i] = &entries[i]
	}
	return b, err
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

package remote

// Remote supports the representation of remote filesystems, and implements data exchange

// ChunkID uniquely represents a data chunk
type ChunkID [16]byte

// EntryID uniquely represents a file (file/directory)
type EntryID [12]byte

// DataSource represents entities who can be queried about filesystem objects.
type DataSource interface {
	Lookup(path string) (EntryID, error)
	ReadMeta(entry EntryID) (NodeMetadata, error)
	ReadData(node ChunkID) ([]byte, error)
}

// DataSink represents entities who can accept data writes.
type DataSink interface {
}

// NodeMetadata represents the metadata of a file/directory.
type NodeMetadata interface {
	ID() EntryID
	IsDirectory() bool
	LocalName() string //No path information
	GetSize() uint64
	GetDataLocality() LocalityInfo //represents where the data is actually stored
}

// LocalityInfo represents information about the concrete location of data.
type LocalityInfo interface {
	IsChunked() bool
	ChunkSize() uint64
	Chunks() []ChunkID
	ChunkAtIndex(int) ChunkID
}
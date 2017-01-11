package nugget

// Remote supports the representation of remote filesystems, and implements data exchange

// ChunkID uniquely represents a data chunk
type ChunkID [16]byte

// EntryID uniquely represents a file (file/directory)
type EntryID [12]byte

// OptimisedDataSourceSink implements optional methods which can speed up most file systems.
type OptimisedDataSourceSink interface {
	Write(fPath string, offset int64, data []byte) (int64, EntryID, NodeMetadata, error)
}

// DataSource represents entities who can be queried about filesystem objects.
type DataSource interface {
	Lookup(path string) (EntryID, error)
	ReadMeta(entry EntryID) (NodeMetadata, error)
	ReadData(node ChunkID) ([]byte, error)
	Fetch(path string) (EntryID, NodeMetadata, []byte, error)
	List(path string) ([]DirEntry, error)
}

// DataSink represents entities who can accept data writes.
type DataSink interface {
	Store(path string, data []byte) (EntryID, NodeMetadata, error)
	Mkdir(path string) (EntryID, NodeMetadata, error)
	Delete(path string) error
	Close() error
}

// DataSourceSink represents an entity which can both accept data writes
// and be queried about filesystem objects.
type DataSourceSink interface {
	DataSink
	DataSource
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
	Chunks() []ChunkID
	ChunkAtIndex(int) ChunkID
}

// DirEntry represents a file/directory stored in a directory.
type DirEntry interface {
	Identifier() string
	IsDirectory() bool
}

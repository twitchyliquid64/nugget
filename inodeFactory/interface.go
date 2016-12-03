package inodeFactory

// inodeFactory keeps track of issued inodes.

// InodeFactory represents a factory which issues unique inodes.
type InodeFactory interface {
	GetInode() uint64
	GetIssued() int // number of inodes out in the wild via this factory.
}

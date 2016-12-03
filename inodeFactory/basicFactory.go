package inodeFactory

import "sync"

// BasicFactory implements InodeFactory, issuing unique inodes without recycling.
type BasicFactory struct {
	LastIssuedInode uint64
	Lock            sync.Mutex
}

// GetInode returns a inode number unique to the factory.
func (f *BasicFactory) GetInode() uint64 {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	f.LastIssuedInode++
	return f.LastIssuedInode
}

// GetIssued returns the total number of issued inodes
func (f *BasicFactory) GetIssued() int {
	return int(f.LastIssuedInode)
}

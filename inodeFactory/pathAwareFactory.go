package inodeFactory

import "sync"

// PathAwareFactory implements InodeFactory, issuing unique inodes where the path has not been seen before.
type PathAwareFactory struct {
	LastIssuedInode uint64
	Paths           map[string]uint64
	Lock            sync.Mutex
}

// GetInode returns a inode number unique to the factory.
func (f *PathAwareFactory) GetInode() uint64 {
	f.Lock.Lock()
	defer f.Lock.Unlock()
	f.LastIssuedInode++
	return f.LastIssuedInode
}

// GetIssued returns the total number of issued inodes
func (f *PathAwareFactory) GetIssued() int {
	return int(f.LastIssuedInode)
}

// GetByPath returns the same inode for a given path, or a unique Inode otherwise.
func (f *PathAwareFactory) GetByPath(path string) uint64 {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	i, ok := f.Paths[path]
	if ok {
		return i
	}
	f.LastIssuedInode++
	f.Paths[path] = f.LastIssuedInode
	return f.LastIssuedInode
}

// MakePathAwareFactory returns an initialized structure ready to be used.
func MakePathAwareFactory() *PathAwareFactory {
	return &PathAwareFactory{
		Paths: map[string]uint64{},
	}
}

//TODO: Need tests!

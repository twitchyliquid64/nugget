package sysstatfs

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// Variable represents a file in sysstatFS which exposes internal information
type Variable interface {
	fs.Node
	fs.Handle
	GetInode() uint64
}

// FixedVariable implements both Node and Handle to allow the variables to be read.
// the variable is static however can be set with FS.SetVariable()
type FixedVariable struct {
	Inode uint64
	Value string
}

// Attr implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (v *FixedVariable) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = v.Inode
	a.Mode = 0444
	a.Size = uint64(len(v.Value))
	return nil
}

// ReadAll implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (v *FixedVariable) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(v.Value), nil
}

// GetInode returns the inode number this Variable will present
func (v *FixedVariable) GetInode() uint64 {
	return v.Inode
}

// ComputedVariable implements both Node and Handle to allow the variables to be read.
// ComputedVariable calls a function to resolve the file contents when it is read.
type ComputedVariable struct {
	Inode    uint64
	Callback func() []byte
}

// Attr implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (v *ComputedVariable) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = v.Inode
	a.Mode = 0444
	a.Size = uint64(len(v.Callback()))
	return nil
}

// ReadAll implements fs.Node, allowing the Variable to masquerade as a fuse file.
func (v *ComputedVariable) ReadAll(ctx context.Context) ([]byte, error) {
	return v.Callback(), nil
}

// GetInode returns the inode number this Variable will present
func (v *ComputedVariable) GetInode() uint64 {
	return v.Inode
}

package sysstatfs

import (
	"context"

	"bazil.org/fuse"
)

// Variable implements both Node and Handle to allow the variables to be read.
type Variable struct {
	Inode uint64
	Value string
}

func (v *Variable) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = v.Inode
	a.Mode = 0444
	a.Size = uint64(len(v.Value))
	return nil
}

func (v *Variable) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(v.Value), nil
}

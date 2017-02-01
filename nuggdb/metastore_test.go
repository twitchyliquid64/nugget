package nuggdb

import (
	"os"
	"testing"

	"github.com/twitchyliquid64/nugget"
)

func TestOpenMetaStoreSucceeds(t *testing.T) {
	p, err := OpenMetaStore("testmetastore.db")
	defer func() {
		p.Close()
		os.Remove("testmetastore.db")
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestMetaCommitSucceeds(t *testing.T) {
	p, err := OpenMetaStore("testmetastore.db")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		p.Close()
		os.Remove("testmetastore.db")
	}()

	meta := EntryMetadata{
		EntryID: nugget.EntryID{'1', '\xA7'},
		Size:    4553,
		Lname:   "bro",
	}

	err = p.Commit(meta)
	if err != nil {
		t.Error(err)
	}
}

func TestMetaChangeTakesEffect(t *testing.T) {
	p, err := OpenMetaStore("testmetastore.db")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		p.Close()
		os.Remove("testmetastore.db")
	}()

	meta := EntryMetadata{
		EntryID: nugget.EntryID{'1', '\xA7'},
		Size:    4553,
		Lname:   "bro",
		IsDir:   true,
		Locality: LocalityInfo{
			ChunkID: nugget.ChunkID{'\x42'},
		},
	}

	err = p.Commit(meta)
	if err != nil {
		t.Error(err)
	}

	v, err := p.Lookup(meta.EntryID)
	if err != nil {
		t.Error(err)
	}

	if v.GetSize() != 4553 {
		t.Error("Size incorrect")
	}

	v.Size = 1
	err = p.Commit(v)
	if err != nil {
		t.Error(err)
	}

	v2, err := p.Lookup(meta.EntryID)
	if err != nil {
		t.Error(err)
	}
	if v2.GetSize() != 1 {
		t.Error("Size incorrect")
	}
}

func TestMetaCommitReadSucceeds(t *testing.T) {
	p, err := OpenMetaStore("testmetastore.db")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		p.Close()
		os.Remove("testmetastore.db")
	}()

	meta := EntryMetadata{
		EntryID: nugget.EntryID{'1', '\xA7'},
		Size:    4553,
		Lname:   "bro",
		IsDir:   true,
		Locality: LocalityInfo{
			ChunkID: nugget.ChunkID{'\x42'},
		},
	}

	err = p.Commit(meta)
	if err != nil {
		t.Error(err)
	}

	v, err := p.Lookup(meta.EntryID)
	if err != nil {
		t.Error(err)
	}
	if v.Lname != meta.Lname {
		t.Error("Name mismatch")
	}
	if v.Size != meta.Size {
		t.Error("Size mismatch")
	}
	if v.IsDir != meta.IsDir {
		t.Error("IsDir mismatch")
	}
	if v.EntryID != meta.EntryID {
		t.Error("EntryID mismatch")
	}
	if v.Locality.ChunkID != meta.Locality.ChunkID {
		t.Error("ChunkID mismatch")
	}
}

func TestMetaLookupNotExistErrorsCorrectly(t *testing.T) {
	p, err := OpenMetaStore("testmetastore.db")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		p.Close()
		os.Remove("testmetastore.db")
	}()

	_, err = p.Lookup(nugget.EntryID{'1', '\xA7'})
	if err != ErrMetaNotFound {
		t.Error("Expected ErrMetaNotFound, got", err)
	}
}

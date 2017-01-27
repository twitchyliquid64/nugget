package nuggdb

import (
	"bytes"
	"os"
	"testing"

	"github.com/twitchyliquid64/nugget"
)

func TestOpenChunkstoreSucceeds(t *testing.T) {
	cs, err := OpenChunkStore("testchunkstore.db")
	defer func() {
		cs.Close()
		os.RemoveAll("testchunkstore.db")
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestLookupForNonexistantChunkErrorsCorrectly(t *testing.T) {
	cs, err := OpenChunkStore("testchunkstore.db")
	defer func() {
		cs.Close()
		os.RemoveAll("testchunkstore.db")
	}()
	if err != nil {
		t.Error(err)
	}

	result, err := cs.Lookup(nugget.ChunkID{})
	if err != ErrChunkNotFound {
		t.Error("Expected ErrChunkNotFound")
	}

	if len(result) != 0 {
		t.Error("Expected len(result) == 0, got", len(result))
	}
}

func TestCommitChunkstoreSucceedsAndLookup(t *testing.T) {
	cs, err := OpenChunkStore("testchunkstore.db")
	defer func() {
		cs.Close()
		os.RemoveAll("testchunkstore.db")
	}()
	if err != nil {
		t.Error(err)
	}

	err = cs.Commit(nugget.ChunkID{13, 22, 11}, []byte{1, 2, 3, 4, 5})
	if err != nil {
		t.Error(err)
	}

	data, err2 := cs.Lookup(nugget.ChunkID{13, 22, 11})
	if err2 != nil {
		t.Error(err2)
	}
	if bytes.Compare(data, []byte{1, 2, 3, 4, 5}) != 0 {
		t.Error("Mismatch. Wanted", []byte{1, 2, 3, 4, 5}, "got", data)
	}
}

func TestForgeChunkstoreSucceedsAndLookup(t *testing.T) {
	cs, err := OpenChunkStore("testchunkstore.db")
	defer func() {
		cs.Close()
		os.RemoveAll("testchunkstore.db")
	}()
	if err != nil {
		t.Error(err)
	}

	var chunkID nugget.ChunkID
	chunkID, err = cs.Forge([]byte{5, 2, 3, 4, 5})
	if err != nil {
		t.Error(err)
	}

	if (chunkID == nugget.ChunkID{}) {
		t.Error("ChunkID should not be empty")
	}

	data, err2 := cs.Lookup(chunkID)
	if err2 != nil {
		t.Error(err2)
	}
	if bytes.Compare(data, []byte{5, 2, 3, 4, 5}) != 0 {
		t.Error("Mismatch. Wanted", []byte{5, 2, 3, 4, 5}, "got", data)
	}
}

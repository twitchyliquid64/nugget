package nuggdb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/twitchyliquid64/nugget"
)

func TestCreateProvider(t *testing.T) {
	baseDir, err := ioutil.TempDir("", "nuggdb_provider_test")
	defer func() {
		os.RemoveAll(baseDir)
	}()
	if err != nil {
		t.Error("Setup error:", err)
		t.FailNow()
	}

	p, err := Create(baseDir)
	if err != nil {
		t.Error(err)
	}
	p.Close()
}

func TestProviderStore(t *testing.T) {
	baseDir, err := ioutil.TempDir("", "nuggdb_provider_test")
	defer func() {
		os.RemoveAll(baseDir)
	}()
	if err != nil {
		t.Error("Setup error:", err)
		t.FailNow()
	}

	p, err := Create(baseDir)
	if err != nil {
		t.Error(err)
	}
	defer p.Close()

	entryID, meta, err := p.Store("/mate/waht", []byte("yolo"))
	if err != nil {
		t.Error(err)
	}
	if (entryID == nugget.EntryID{}) {
		t.Error("EntryID is empty:", entryID)
	}
	if (meta.GetDataLocality().Chunks()[0] == nugget.ChunkID{}) {
		t.Error("ChunkID is empty:", meta.GetDataLocality().Chunks())
	}
	foundEntry, err := p.Lookup("/mate/waht")
	if err != nil {
		t.Error(err)
	}
	if foundEntry != entryID {
		t.Error("Expected matching entryID's", foundEntry, entryID)
	}
	foundMeta, err := p.ReadMeta(foundEntry)
	if err != nil {
		t.Error(err)
	}
	if foundMeta.LocalName() != "waht" {
		t.Error("Local name incorrect")
	}
	if foundMeta.ID() != foundEntry {
		t.Error("Meta entryID incorrect", foundMeta.ID(), foundEntry)
	}
	if foundMeta.GetDataLocality().Chunks()[0] != meta.GetDataLocality().Chunks()[0] {
		t.Error("Chunk ID mismatch")
	}
	foundData, err := p.ReadData(foundMeta.GetDataLocality().Chunks()[0])
	if err != nil {
		t.Error(err)
	}
	if string(foundData) != "yolo" {
		t.Error("Data incorrect")
	}
}

//TODO: Tests for each of the error conditions in Provider.Store()

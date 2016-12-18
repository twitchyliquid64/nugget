package nuggdb

import (
	"os"
	"testing"

	"github.com/twitchyliquid64/nugget"
)

func TestOpenPathStoreSucceeds(t *testing.T) {
	p, err := OpenPathStore("testpathstore.db")
	defer func() {
		p.Close()
		os.Remove("testpathstore.db")
	}()
	if err != nil {
		t.Error(err)
	}
}

func TestCommitSucceeds(t *testing.T) {
	p, err := OpenPathStore("testpathstore.db")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		p.Close()
		os.Remove("testpathstore.db")
	}()

	err = p.Commit("/new", nugget.EntryID{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'})
	if err != nil {
		t.Error(err)
	}
}

func TestLookupForNonexistantPathErrorsCorrectly(t *testing.T) {
	p, err := OpenPathStore("testpathstore.db")
	defer func() {
		p.Close()
		os.Remove("testpathstore.db")
	}()
	if err != nil {
		t.Error(err)
	}

	result, err := p.Lookup("/does/not/exist")
	if err != ErrPathNotFound {
		t.Error("Expected ErrPathNotFound")
	}
	if (result != nugget.EntryID{'\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00'}) {
		t.Error("Expected empty result, got ", result)
	}
}

func TestCommitLookupWorks(t *testing.T) {
	p, err := OpenPathStore("testpathstore.db")
	defer func() {
		p.Close()
		os.Remove("testpathstore.db")
	}()
	if err != nil {
		t.Error(err)
	}

	err = p.Commit("/thing", nugget.EntryID{'0', '3', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'})
	if err != nil {
		t.Error(err)
	}

	result, err := p.Lookup("/thing")
	if err != nil {
		t.Error(err)
	}
	if (result != nugget.EntryID{'0', '3', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}) {
		t.Error("Expected specific EntryID, got ", result)
	}
}

func TestForgeWorks(t *testing.T) {
	p, err := OpenPathStore("testpathstore.db")
	defer func() {
		p.Close()
		os.Remove("testpathstore.db")
	}()

	if err != nil {
		t.Error(err)
	}

	id, err := p.Forge("/forge")
	if err != nil {
		t.Error(err)
	}

	if (id == nugget.EntryID{'\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00', '\x00'}) {
		t.Error("Was not expecting empty EntryID")
	}

	result, err := p.Lookup("/forge")
	if err != nil {
		t.Error(err)
	}
	if result != id {
		t.Error("Expected specific EntryID, got ", result)
	}
}

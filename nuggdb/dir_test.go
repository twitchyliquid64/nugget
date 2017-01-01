package nuggdb

import (
	"encoding/binary"
	"testing"
)

func TestSerializeDirEntryCorrectLen(t *testing.T) {
	d := dirEntry{Name: "yolo.swag/"}
	s := d.Serialize()
	if len(s) != (5 + len(d.Name) + 1) {
		t.Error("Incorrect len")
	}
}

func TestSerializeDirEntryProducesVersion1(t *testing.T) {
	d := dirEntry{Name: "yolo.sw434634gdfsdfds zfdgdsag/"}
	s := d.Serialize()
	if binary.LittleEndian.Uint16(s[0:2]) != 1 {
		t.Error("Expected 1")
	}
}

func TestSerializeDirEntryProducesCorrectNameLen(t *testing.T) {
	d := dirEntry{Name: "yolo.sw434634gdfsdfds zfdgdsag/"}
	s := d.Serialize()
	if binary.LittleEndian.Uint16(s[3:5]) != uint16(len(d.Name)) {
		t.Error("Expected ", len(d.Name))
	}
}

func TestDirEntryDeserialize(t *testing.T) {
	d := dirEntry{Name: "yolo.sw434634gdfsdfdszfdgdsag/", IsDir: true}
	s := d.Serialize()
	d2, err := deserializeDirEntry(s)
	if err != nil {
		t.Error(err)
	}
	if d.Name != d2.Name {
		t.Error("Name mismatch: ", d.Name, d2.Name)
	}
	if d.IsDir != d2.IsDir {
		t.Error("IsDir mismatch: ", d.IsDir, d2.IsDir)
	}
}

func TestEntriesSerializeLen(t *testing.T) {
	d1 := dirEntry{Name: "1 lola"}
	d2 := dirEntry{Name: "kek", IsDir: true}
	d3 := dirEntry{Name: "wut"}
	entries := dirEntries{d1, d2, d3}
	b := entries.Serialize()
	if len(b) != (2 + dirEntrySize(len(d1.Name)) + dirEntrySize(len(d2.Name)) + dirEntrySize(len(d3.Name))) {
		t.Error("Incorrect len")
	}

	if binary.LittleEndian.Uint16(b[0:2]) != 3 {
		t.Error("Expected 3")
	}
}

func TestEntriesSerializeDeserialize(t *testing.T) {
	d1 := dirEntry{Name: "1 lola"}
	d2 := dirEntry{Name: "kek", IsDir: true}
	d3 := dirEntry{Name: "wut"}
	entries := dirEntries{d1, d2, d3}
	b := entries.Serialize()
	reversedEntries, err := deserializeDirEntries(b)

	if err != nil {
		t.Error(err)
	}

	if len(reversedEntries) != 3 {
		t.Error("Incorrect len")
	}

	if reversedEntries[0].Name != d1.Name || reversedEntries[0].IsDir != d1.IsDir {
		t.Error("First entry incorrect")
	}
	if reversedEntries[1].Name != d2.Name || reversedEntries[1].IsDir != d2.IsDir {
		t.Error("Second entry incorrect")
	}
	if reversedEntries[2].Name != d3.Name || reversedEntries[2].IsDir != d3.IsDir {
		t.Error("Third entry incorrect")
	}
}

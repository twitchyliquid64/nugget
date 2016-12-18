package nuggdb

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/twitchyliquid64/nugget"
)

func TestSerialize(t *testing.T) {
	a := EntryMetadata{
		EntryID:   nugget.EntryID{'1', '2', '0', '0', '0', '0', '6', '0', '\xFF', '0', '0', 'A'},
		LocalName: "Lolz",
		Size:      54634532544,
		IsDir:     true,
		Locality: LocalityInfo{
			ChunkID: nugget.ChunkID{'1', '2'},
		},
	}

	if len(a.Serialize()) != (12 + 100 + 8 + 2 + 16) {
		t.Error("Len incorrect")
	}

	var test nugget.EntryID
	copy(test[:], a.Serialize()[:12])
	if test != a.EntryID {
		t.Error("EntryID does not match")
	}

	localName := string(bytes.Trim(a.Serialize()[12:12+100], "\x00"))
	if localName != a.LocalName {
		t.Error("Names do not match, got", localName, "len", len(localName))
	}

	size := binary.LittleEndian.Uint64(a.Serialize()[12+100 : 12+100+8])
	if size != a.Size {
		t.Error("Expected sizes to match, got", size)
	}
	isDir := (a.Serialize()[12+100+8] & 1) == 1
	if isDir != a.IsDir {
		t.Error("IsDir does not match, got", isDir)
	}

	var chunk nugget.ChunkID
	copy(chunk[:], a.Serialize()[12+100+8+2:12+100+8+2+16])
	if chunk != a.Locality.ChunkID {
		t.Error("Expected chunkID to match")
	}
}

func TestSerializeDeserialize(t *testing.T) {
	a := EntryMetadata{
		EntryID:   nugget.EntryID{'1', '2', '0', '0', '0', '0', '6', '0', '\xFF', '0', '0', 'A'},
		LocalName: "Lolz",
		Size:      54634532544,
		IsDir:     true,
		Locality: LocalityInfo{
			ChunkID: nugget.ChunkID{'1', '2'},
		},
	}
	buff := a.Serialize()
	out := MakeMetadata(buff)
	if out.IsDir != a.IsDir {
		t.Error("Expected IsDir to match")
	}
	if out.LocalName != a.LocalName {
		t.Error("Expected LocalName to match")
	}
	if out.Size != a.Size {
		t.Error("Expected Size to match")
	}
	if out.EntryID != a.EntryID {
		t.Error("Expected Size to match")
	}
	if out.Locality.ChunkID != a.Locality.ChunkID {
		t.Error("Expected Locality.ChunkID to match")
	}
}

package nuggdb

import (
	"bytes"
	"encoding/binary"

	"github.com/twitchyliquid64/nugget"
)

// EntryMetadata is a concrete implementation of nugget.NodeMetadata
type EntryMetadata struct {
	IsDir    bool
	Lname    string
	EntryID  nugget.EntryID
	Size     uint64
	Locality LocalityInfo
}

// ID returns the EntryID.
func (meta *EntryMetadata) ID() nugget.EntryID {
	return meta.EntryID
}

// IsDirectory returns true if the metadata entry represents a directory
func (meta *EntryMetadata) IsDirectory() bool {
	return meta.IsDir
}

// LocalName returns the localised name of the metadata entry - ie without path info.
func (meta *EntryMetadata) LocalName() string {
	return meta.Lname
}

// GetSize returns the size of the file in bytes. Returns number of files for directories.
func (meta *EntryMetadata) GetSize() uint64 {
	return meta.Size
}

// GetDataLocality returns information about the chunks where the actual data is stored.
func (meta *EntryMetadata) GetDataLocality() nugget.LocalityInfo {
	return &meta.Locality
}

// Serialize returns a byte slice which represents the EntryMetadata structure.
func (meta *EntryMetadata) Serialize() []byte {
	buff := make([]byte, 12+100+8+2+16) //EntryID + LocalName + Size + flags + len(LocalityInfo)
	copy(buff[:12], meta.EntryID[:])
	copy(buff[12:100+12], meta.Lname)
	binary.LittleEndian.PutUint64(buff[12+100:12+100+8], meta.Size)
	if meta.IsDir {
		buff[12+100+8] |= (1 << 0)
	}
	copy(buff[12+100+8+2:], meta.Locality.Serialize())
	return buff
}

// LocalityInfo is a concrete implementation of nugget.LocalityInfo
// TODO: Update to support multiple chunks, remember to update tests meta.Serialize as well
type LocalityInfo struct {
	ChunkID nugget.ChunkID
}

// IsChunked always returns false, multiple chunks per file is not yet implemented.
func (l *LocalityInfo) IsChunked() bool {
	return false
}

// Chunks returns an ordered slice of all the chunks which make up the file.
func (l *LocalityInfo) Chunks() []nugget.ChunkID {
	return []nugget.ChunkID{l.ChunkID}
}

// ChunkAtIndex returns the chunkID at the index of the array of chunks which make up the file.
func (l *LocalityInfo) ChunkAtIndex(pos int) nugget.ChunkID {
	return l.ChunkID
}

// Serialize returns a byte slice which represents the LocalityInfo structure.
func (l *LocalityInfo) Serialize() []byte {
	buff := make([]byte, 16)
	copy(buff, l.ChunkID[:])
	return buff
}

// MakeMetadata constructs a EntryMetadata from the byte slice.
func MakeMetadata(data []byte) EntryMetadata {
	if len(data) != (12 + 100 + 8 + 2 + 16) {
		panic("Len incorrect")
	}
	ret := EntryMetadata{}

	copy(ret.EntryID[:], data[:12])
	ret.Lname = string(bytes.Trim(data[12:12+100], "\x00"))

	ret.Size = binary.LittleEndian.Uint64(data[12+100 : 12+100+8])

	ret.IsDir = (data[12+100+8] & 1) == 1
	ret.Locality = MakeLocality(data[12+100+8+2:])
	return ret
}

// MakeLocality constructs a LocalityInfo struct from the byte slice.
func MakeLocality(data []byte) LocalityInfo {
	var chunkID nugget.ChunkID
	copy(chunkID[:], data)
	return LocalityInfo{
		ChunkID: chunkID,
	}
}

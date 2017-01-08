package nuggdb

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type DirEntry struct {
	EntryVersion uint16
	IsDir        bool
	Name         string
}

func DirEntrySize(nameLen int) int {
	return 2 + 1 + 2 + nameLen + 1
}

func (e *DirEntry) Identifier() string {
	return e.Name
}

func (e *DirEntry) IsDirectory() bool {
	return e.IsDir
}

// Serialize returns a byte slice which represents the DirEntry structure.
func (e *DirEntry) Serialize() []byte {
	buff := make([]byte, DirEntrySize(len(e.Name))) //EntryVersion + Flags(IsDir) + nameSize(uint16) + Name + nullbyte
	e.EntryVersion = 1

	//Version
	binary.LittleEndian.PutUint16(buff[:2], e.EntryVersion)

	//Flags
	if e.IsDir {
		buff[2] |= (1 << 0)
	}

	//Name length + Name
	binary.LittleEndian.PutUint16(buff[3:5], uint16(len(e.Name)))
	copy(buff[6:6+len(e.Name)], []byte(e.Name))
	return buff
}

func deserializeDirEntry(data []byte) (DirEntry, error) {
	var out DirEntry
	out.EntryVersion = binary.LittleEndian.Uint16(data[0:2])
	if out.EntryVersion == 1 {
		out.IsDir = (data[2] & 1) == 1
		nameSize := binary.LittleEndian.Uint16(data[3:5])
		out.Name = string(data[6 : 6+nameSize])
		return out, nil
	}

	return DirEntry{}, errors.New("Unknown version")
}

type dirEntries []DirEntry

func (dir dirEntries) Serialize() []byte {
	b := new(bytes.Buffer)

	lenBuff := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBuff, uint16(len([]DirEntry(dir))))
	b.Write(lenBuff)

	for _, entry := range []DirEntry(dir) {
		b.Write(entry.Serialize())
	}

	return b.Bytes()
}

func deserializeDirEntries(data []byte) ([]DirEntry, error) {
	size := int(binary.LittleEndian.Uint16(data[0:2]))
	out := make([]DirEntry, size)
	var err error

	cursor := 2
	for i := 0; i < size; i++ {
		out[i], err = deserializeDirEntry(data[cursor:])
		if err != nil {
			return out, err
		}
		cursor += DirEntrySize(len(out[i].Name))
	}
	return out, nil
}

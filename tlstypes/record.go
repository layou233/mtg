package tlstypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const recordMaxChunkSize = 16384 + 24

type Record struct {
	Type    RecordType
	Version Version
	Data    Byter
}

func (r Record) WriteBytes(writer io.Writer) {
	writer.Write([]byte{byte(r.Type)})                           //nolint: errcheck
	writer.Write(r.Version.Bytes())                              //nolint: errcheck
	binary.Write(writer, binary.BigEndian, uint16(r.Data.Len())) //nolint: errcheck
	r.Data.WriteBytes(writer)
}

func (r Record) Len() int {
	return 1 + 2 + 2 + r.Data.Len()
}

func ReadRecord(reader io.Reader) (Record, error) {
	buf := [2]byte{}
	rec := Record{}

	if _, err := io.ReadFull(reader, buf[:1]); err != nil {
		return rec, fmt.Errorf("cannot read record type: %w", err)
	}

	rec.Type = RecordType(buf[0])

	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return rec, fmt.Errorf("cannot read version: %w", err)
	}

	switch {
	case bytes.Equal(buf[:], Version13Bytes):
		rec.Version = Version13
	case bytes.Equal(buf[:], Version12Bytes):
		rec.Version = Version12
	case bytes.Equal(buf[:], Version11Bytes):
		rec.Version = Version11
	case bytes.Equal(buf[:], Version10Bytes):
		rec.Version = Version10
	}

	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return rec, fmt.Errorf("cannot read data length: %w", err)
	}

	data := make([]byte, binary.BigEndian.Uint16(buf[:]))
	if _, err := io.ReadFull(reader, data); err != nil {
		return rec, fmt.Errorf("cannot read data: %w", err)
	}

	rec.Data = RawBytes(data)

	return rec, nil
}

func MakeRecords(raw []byte) []Record {
	var arr []Record

	for len(raw) > 0 {
		chunkSize := recordMaxChunkSize
		if chunkSize > len(raw) {
			chunkSize = len(raw)
		}

		arr = append(arr, Record{
			Type:    RecordTypeApplicationData,
			Version: Version12,
			Data:    RawBytes(raw[:chunkSize]),
		})
		raw = raw[chunkSize:]
	}

	return arr
}

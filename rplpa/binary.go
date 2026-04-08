package rplpa

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/bnch/uleb128"
)

// this file includes functions (BString, RBString) from https://github.com/bnch/bancho credits goes to thehowl. under MIT license.

// BString returns a Binary array of an Osu! Encoded string!
func wbString(s string) []byte {
	if s == "" {
		return []byte{0}
	}
	b := []byte{11}
	b = append(b, uleb128.Marshal(len(s))...)
	b = append(b, []byte(s)...)
	return b
}

// RBString reads an Osu! Encoded string of the Given io.Reader returns an String else Error
func rBString(value io.Reader) (s string, err error) {
	bufferSlice := make([]byte, 1)
	value.Read(bufferSlice)
	if bufferSlice[0] != 11 {
		return "", nil
	}
	length := uleb128.UnmarshalReader(value)
	bufferSlice = make([]byte, length)
	b, err := value.Read(bufferSlice)
	if b < length {
		err = errors.New("Unexpected end of string")
	}
	s = string(bufferSlice)
	return
}

// Int returns an Binary encoded Int
func wInt(value int) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RInt Reads an Binary encoded int with the given io.Reader returns int else error
func rInt(value io.Reader) (i int, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// UInt returns an Binary encoded unsigned Int
func wuInt(value uint) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RUInt Reads an Binary encoded unsigned Int and returns a uint else an error
func rUInt(value io.Reader) (i uint, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Int8 returns a Binary encoded Int8
func wInt8(value int8) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RInt8 Reads an Binary encoded int8 with the given io.Reader returns int8 else error
func rInt8(value io.Reader) (i int8, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// UInt8 returns an Binary encoded unsigned Int8
func wUInt8(value uint8) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RUInt8 Reads an Binary encoded unsigned Int8 and returns a uint8 else an error
func rUInt8(value io.Reader) (i uint8, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Int16 returns a Binary encoded Int16
func wInt16(value int16) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RInt16 Reads an Binary encoded int16 with the given io.Reader returns int16 else error
func rInt16(value io.Reader) (i int16, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// UInt16 returns an Binary encoded unsigned Int16
func wUInt16(value uint16) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RUInt16 Reads an Binary encoded unsigned Int16 and returns a uint16 else an error
func rUInt16(value io.Reader) (i uint16, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Int32 returns a Binary encoded Int32
func wInt32(value int32) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RInt32 Reads an Binary encoded int32 with the given io.Reader returns int32 else error
func rInt32(value io.Reader) (i int32, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// UInt32 returns an Binary encoded unsigned Int32
func wUInt32(value uint32) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RUInt32 Reads an Binary encoded unsigned Int32 and returns a uint32 else an error
func rUInt32(value io.Reader) (i uint32, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Int64 returns a Binary encoded Int64
func wInt64(value int64) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RInt64 Reads an Binary encoded int64 with the given io.Reader returns int64 else error
func rInt64(value io.Reader) (i int64, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// UInt64 returns an Binary encoded unsigned Int64
func wUInt64(value uint64) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RUInt64 Reads an Binary encoded unsigned Int64 and returns a uint64 else an error
func rUInt64(value io.Reader) (i uint64, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Float32 Returns an Binary encoded Float32
func wFloat32(value float32) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RFloat32 Reads an float32 of the given io.Reader, returns float32 else error
func rFloat32(value io.Reader) (i float32, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Float64 Returns an Binary encoded Float64
func wFloat64(value float64) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, value)
	return b.Bytes()
}

// RFloat64 Reads an float64 of the given io.Reader, returns float64 else error
func rFloat64(value io.Reader) (i float64, err error) {
	err = binary.Read(value, binary.LittleEndian, &i)
	return
}

// Bool returns an Binary encoded Boolean using int8
func wBool(value bool) []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, int8(func() int8 {
		if value {
			return int8(1)
		}
		return int8(0)
	}()))
	return b.Bytes()
}

// RBool reads a Binary encoded boolean using int8, returns bool else error
func rBool(value io.Reader) (i bool, err error) {
	var m int8
	err = binary.Read(value, binary.LittleEndian, &m)
	i = m > 0
	return
}

func rSlice(value io.Reader, length int32) (s []byte, err error) {
	s = make([]byte, length)
	_, err = value.Read(s)
	return
}

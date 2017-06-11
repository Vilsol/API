package utils

import (
	"bufio"
	"time"
	"encoding/binary"
)

type Buffer struct {
	Data []byte
}

type PacketParser struct {
	Reader *bufio.Reader
	Buffer *Buffer
}

func NewParser(reader *bufio.Reader) (PacketParser) {
	buffer := Buffer{make([]byte, 0)}

	parser := PacketParser{
		Reader: reader,
		Buffer: &buffer,
	}

	return parser
}

/**
Read an integer from n amount of bytes
 */
func (parser PacketParser) ReadInt(bytes int) int {
	if !parser.WaitUntilBuffered(bytes) {
		return 0
	}

	var integer int
	b := parser.ReadBytes(bytes)

	if bytes == 1 {
		integer = int(b[0])
	} else if bytes == 2 {
		integer = int(binary.LittleEndian.Uint16(b))
	} else if bytes == 4 {
		integer = int(binary.LittleEndian.Uint32(b))
	} else if bytes == 8 {
		integer = int(binary.LittleEndian.Uint64(b))
	}

	return integer
}

/**
Peek an integer from n amount of bytes
 */
func (parser PacketParser) PeekInt(bytes int) int {
	if !parser.WaitUntilBuffered(bytes) {
		return 0
	}

	var integer int
	b, _ := parser.Reader.Peek(bytes)

	if bytes == 1 {
		integer = int(b[0])
	} else if bytes == 2 {
		integer = int(binary.LittleEndian.Uint16(b))
	} else if bytes == 4 {
		integer = int(binary.LittleEndian.Uint32(b))
	} else if bytes == 8 {
		integer = int(binary.LittleEndian.Uint64(b))
	}

	return integer
}

/**
Read a string from an integer of n amount of bytes
 */
func (parser PacketParser) ReadString(bytes int) (int, string) {
	var length = parser.ReadInt(bytes)

	if !parser.WaitUntilBuffered(length) {
		return 0, ""
	}

	return length, string(parser.ReadBytes(length))
}

/**
Wait until n amount of bytes are buffered or timeout after 50ms
 */
func (parser PacketParser) WaitUntilBuffered(bytes int) bool {
	if parser.Reader.Buffered() < bytes {
		for i := 0; i < 50; i++ {
			time.Sleep(1 * time.Millisecond)
			if parser.Reader.Buffered() >= bytes {
				return true
			}
		}

		return false
	}

	return true
}

/**
Read n amount of bytes
 */
func (parser PacketParser) ReadBytes(bytes int) []byte {
	if !parser.WaitUntilBuffered(bytes) {
		return nil
	}

	result := make([]byte, bytes)

	for i := 0; i < bytes; i++ {
		result[i], _ = parser.Reader.ReadByte()
	}

	appended := append(parser.Buffer.Data, result...)
	parser.Buffer.Data = appended

	return result
}

/**
Discard n amount of bytes
 */
func (parser PacketParser) DiscardBytes(bytes int) {
	if !parser.WaitUntilBuffered(bytes) {
		return
	}

	parser.ReadBytes(bytes)
}

/**
Skip bytes that match provided byte
 */
func (parser PacketParser) SkipWhile(byte byte) {
	for true {
		if !parser.WaitUntilBuffered(1) {
			return
		}

		b, _ := parser.Reader.Peek(1)
		if b[0] != byte {
			break
		}

		parser.ReadBytes(1)
	}
}

/**
Get all bytes that this parser processed
 */
func (parser PacketParser) ProcessAllBytes() {
	parser.ReadBytes(parser.Reader.Buffered())
}

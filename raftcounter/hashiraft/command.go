package hashiraft

import "bytes"

type CommandType uint8

const (
	IncrCounterCommand CommandType = iota
	DecrCounterCommand
	GetCurrentCounterCommand
)

func (c CommandType) Encode() []byte {
	var buf bytes.Buffer
	buf.WriteByte(uint8(c))
	return buf.Bytes()

}

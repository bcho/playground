package hashiraft

import (
	"bytes"
	"encoding/binary"
	"io"
)

type counter struct {
	value int64
}

func newCounter() *counter {
	return &counter{
		value: 0,
	}
}

func (c *counter) Incr() int64 {
	c.value = c.value + 1
	return c.value
}

func (c *counter) Decr() int64 {
	c.value = c.value - 1
	return c.value
}

func (c *counter) Current() int64 {
	return c.value
}

func (c *counter) EncodeValue() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, c.value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *counter) DecodeValue(rc io.ReadCloser) error {
	var value int64
	err := binary.Read(rc, binary.LittleEndian, &value)
	if err == nil {
		c.value = value
	}
	return err
}

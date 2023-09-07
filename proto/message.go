package proto

import "bytes"

type Header struct {
	PackLength uint32
	HeadLength uint16
	Version    uint16
	Operation  uint32
	Sequence   uint32
}

func (h Header) ToBytes() []byte {
	var buffer bytes.Buffer

	buffer.Write(bigEndianUint32(h.PackLength))
	buffer.Write(bigEndianUint16(h.HeadLength))
	buffer.Write(bigEndianUint16(h.Version))
	buffer.Write(bigEndianUint32(h.Operation))
	buffer.Write(bigEndianUint32(h.Sequence))

	return buffer.Bytes()
}

type Message struct {
	header  Header
	payload []byte
}

func (message Message) ToBytes() []byte {
	var buffer bytes.Buffer

	buffer.Write(message.header.ToBytes())
	buffer.Write(message.payload)

	return buffer.Bytes()
}

func (message Message) Operation() uint32 {
	return message.header.Operation
}

func (message Message) Payload() []byte {
	return message.payload
}

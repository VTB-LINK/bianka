/*
 * MIT License
 *
 * Copyright (c) 2023 vlink and runstp.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS," WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE, AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
 * OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES, OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT, OR OTHERWISE, ARISING FROM, OUT OF,
 * OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

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

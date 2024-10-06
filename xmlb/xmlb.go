/*
Copyright (c) 2024, Nao Yonashiro
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// Package xmlb provides a high-performance bridge between the gosax library and encoding/xml.
// It is designed to facilitate the rewriting of code that uses encoding/xml, offering a more efficient
// and memory-conscious approach to XML parsing.
//
// While gosax provides a low-level bridge with encoding/xml through various utility functions,
// xmlb offers a higher-performance bridge intended for rewriting.
package xmlb

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"

	"github.com/orisano/gosax"
)

const (
	StartElement = iota + 1
	EndElement
	CharData
	ProcInst
	Comment
	Directive
)

type Decoder struct {
	r *gosax.Reader
}

func NewDecoder(r io.Reader, buf []byte) *Decoder {
	gr := gosax.NewReaderBuf(r, buf)
	gr.EmitSelfClosingTag = true
	return &Decoder{gr}
}

func (d *Decoder) Token() (Token, error) {
	ev, err := d.r.Event()
	if err == nil && ev.Type() == gosax.EventEOF {
		err = io.EOF
	}
	if err != nil {
		return Token{}, err
	}
	return Token(ev), nil
}

func (d *Decoder) Skip() error {
	return gosax.Skip(d.r)
}

type Token gosax.Event

func (t Token) Type() uint8 {
	switch gosax.Event(t).Type() {
	case gosax.EventStart:
		return StartElement
	case gosax.EventEnd:
		return EndElement
	case gosax.EventText:
		return CharData
	case gosax.EventCData:
		return CharData
	case gosax.EventProcessingInstruction:
		return ProcInst
	case gosax.EventComment:
		return Comment
	case gosax.EventDocType:
		return Directive
	case gosax.EventEOF:
		return 0
	default:
		panic("unreachable")
	}
}

type NameBytes struct {
	b []byte
	p int
}

func (n NameBytes) Space() []byte {
	if n.p == 0 {
		return nil
	}
	return n.b[:n.p-1]
}

func (n NameBytes) Local() []byte {
	return n.b[n.p:]
}

var ErrNoAttributes = errors.New("no attributes")

type AttributesBytes []byte

func (a AttributesBytes) Get(key string) ([]byte, error) {
	b := []byte(a)
	for len(b) > 0 {
		attr, b2, err := gosax.NextAttribute(b)
		if err != nil {
			return nil, err
		}
		b = b2
		if string(attr.Key) != key {
			continue
		}
		v, err := gosax.Unescape(attr.Value[1 : len(attr.Value)-1])
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, ErrNoAttributes
}

type StartElementBytes struct {
	Name  NameBytes
	Attrs AttributesBytes
}

func (t Token) Name() NameBytes {
	name, _ := gosax.Name(t.Bytes)
	p := bytes.IndexByte(name, ':')
	if p < 0 {
		return NameBytes{name, 0}
	}
	return NameBytes{name, p + 1}
}

func (t Token) StartElement() (xml.StartElement, error) {
	return gosax.StartElement(t.Bytes)
}

func (t Token) StartElementBytes() StartElementBytes {
	name, attrs := gosax.Name(t.Bytes)
	p := bytes.IndexByte(name, ':')
	if p < 0 {
		p = 0
	} else {
		p += 1
	}
	return StartElementBytes{NameBytes{name, p}, attrs}
}

func (t Token) EndElement() xml.EndElement {
	return gosax.EndElement(t.Bytes)
}

func (t Token) CharData() (xml.CharData, error) {
	switch t.Type() {
	case gosax.EventText:
		return gosax.CharData(t.Bytes)
	case gosax.EventCData:
		return bytes.TrimSuffix(bytes.TrimPrefix(t.Bytes, []byte("<![CDATA[")), []byte("]]>")), nil
	default:
		panic("unreachable")
	}
}

func (t Token) ProcInst() xml.ProcInst {
	return gosax.ProcInst(t.Bytes)
}

func (t Token) Comment() xml.Comment {
	return gosax.Comment(t.Bytes)
}

func (t Token) Directive() xml.Directive {
	return gosax.Directive(t.Bytes)
}

package xmlb

import (
	"bytes"
	"encoding/xml"
	"errors"

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

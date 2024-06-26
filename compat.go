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

// This file contains utility functions to bridge gosax with encoding/xml.
// These functions provide convenient ways to convert gosax types to encoding/xml types,
// facilitating interoperability between the two packages.

package gosax

import (
	"bytes"
	"encoding/xml"
	"io"
)

// StartElement converts a byte slice to an xml.StartElement.
func StartElement(b []byte) (xml.StartElement, error) {
	name, b := Name(b)
	e := xml.StartElement{
		Name: xmlName(name),
	}
	for len(b) > 0 {
		var attr Attribute
		var err error
		attr, b, err = NextAttribute(b)
		if err != nil {
			return xml.StartElement{}, err
		}
		e.Attr = append(e.Attr, xml.Attr{
			Name:  xmlName(attr.Key),
			Value: string(attr.Value[1 : len(attr.Value)-1]),
		})
	}
	return e, nil
}

// EndElement converts a byte slice to an xml.EndElement.
func EndElement(b []byte) xml.EndElement {
	name, _ := Name(b)
	return xml.EndElement{
		Name: xmlName(name[1:]),
	}
}

// CharData converts a byte slice to xml.CharData.
func CharData(b []byte) (xml.CharData, error) {
	return Unescape(b)
}

// Comment converts a byte slice to an xml.Comment.
func Comment(b []byte) xml.Comment {
	return trim(b, "<!--", "-->")
}

// ProcInst converts a byte slice to an xml.ProcInst.
func ProcInst(b []byte) xml.ProcInst {
	name, b := Name(b)
	return xml.ProcInst{
		Target: string(name[1:]),
		Inst:   b[:len(b)-1],
	}
}

// Directive converts a byte slice to an xml.Directive.
func Directive(b []byte) xml.Directive {
	return trim(b, "<!", ">")
}

// Token converts an Event to an xml.Token.
// This function is provided for convenience, but it may allocate memory.
//
// Note: For performance-critical applications, it's recommended to use
// the direct conversion functions (StartElement, EndElement, CharData, etc.)
// instead of Token, as they allow better control over memory allocations.
func Token(e Event) (xml.Token, error) {
	switch e.Type() {
	case EventStart:
		return StartElement(e.Bytes)
	case EventEnd:
		return EndElement(e.Bytes), nil
	case EventText:
		return CharData(e.Bytes)
	case EventCData:
		return xml.CharData(trim(e.Bytes, "<![CDATA[", "]]>")), nil
	case EventComment:
		return Comment(e.Bytes), nil
	case EventProcessingInstruction:
		return ProcInst(e.Bytes), nil
	case EventDocType:
		return Directive(e.Bytes), nil
	case EventEOF:
		return nil, io.EOF
	default:
		panic("unknown event type")
	}
}

func xmlName(b []byte) xml.Name {
	if i := bytes.IndexByte(b, ':'); i >= 0 {
		return xml.Name{
			Space: string(b[:i]),
			Local: string(b[i+1:]),
		}
	} else {
		return xml.Name{
			Local: string(b),
		}
	}
}

func trim(b []byte, prefix, suffix string) []byte {
	return bytes.TrimSuffix(bytes.TrimPrefix(b, []byte(prefix)), []byte(suffix))
}

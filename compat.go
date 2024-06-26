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

package gosax

import (
	"bytes"
	"encoding/xml"
)

func StartElement(b []byte) xml.StartElement {
	name, b := Name(b)
	e := xml.StartElement{
		Name: xmlName(name),
	}
	for len(b) > 0 {
		var attr Attribute
		attr, b = NextAttribute(b)
		e.Attr = append(e.Attr, xml.Attr{
			Name:  xmlName(attr.Key),
			Value: string(attr.Value[1 : len(attr.Value)-1]),
		})
	}
	return e
}

func EndElement(b []byte) xml.EndElement {
	name, _ := Name(b)
	return xml.EndElement{
		Name: xmlName(name[1:]),
	}
}

func CharData(b []byte) (xml.CharData, error) {
	return Unescape(b)
}

func Comment(b []byte) xml.Comment {
	return trim(b, "<!--", "-->")
}

func ProcInst(b []byte) xml.ProcInst {
	name, b := Name(b)
	return xml.ProcInst{
		Target: string(name[1:]),
		Inst:   b[:len(b)-1],
	}
}

func Directive(b []byte) xml.Directive {
	return trim(b, "<!", ">")
}

func Token(e Event) (xml.Token, error) {
	switch e.Type() {
	case EventStart:
		return StartElement(e.Bytes), nil
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

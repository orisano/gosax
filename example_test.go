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

package gosax_test

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/orisano/gosax"
)

func ExampleReader_Event() {
	xmlData := `<root><element>Value</element></root>`
	reader := strings.NewReader(xmlData)

	r := gosax.NewReader(reader)
	for {
		e, err := r.Event()
		if err != nil {
			log.Fatal(err)
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		fmt.Println(string(e.Bytes))
	}
	// Output:
	// <root>
	// <element>
	// Value
	// </element>
	// </root>
}

func ExampleNewReaderBuf() {
	xmlData := `<root><element>Value</element></root>`
	reader := strings.NewReader(xmlData)

	var buf [4096]byte
	r := gosax.NewReaderBuf(reader, buf[:])
	for {
		e, err := r.Event()
		if err != nil {
			log.Fatal(err)
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		fmt.Println(string(e.Bytes))
	}
	// Output:
	// <root>
	// <element>
	// Value
	// </element>
	// </root>
}

func ExampleReader_Reset() {
	pool := sync.Pool{
		New: func() any {
			return gosax.NewReaderSize(nil, 16*1024)
		},
	}
	func(p *sync.Pool) {
		xmlData := `<root><element>Value</element></root>`
		reader := strings.NewReader(xmlData)

		r := p.Get().(*gosax.Reader)
		defer p.Put(r)
		r.Reset(reader)
		for {
			e, err := r.Event()
			if err != nil {
				log.Fatal(err)
			}
			if e.Type() == gosax.EventEOF {
				break
			}
			fmt.Println(string(e.Bytes))
		}
	}(&pool)
	// Output:
	// <root>
	// <element>
	// Value
	// </element>
	// </root>
}

func ExampleToken() {
	xmlData := `<root><element foo="&lt;bar&gt;" bar="qux">Value</element></root>`
	reader := strings.NewReader(xmlData)

	r := gosax.NewReader(reader)
	for {
		e, err := r.Event()
		if err != nil {
			log.Fatal(err)
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		t, err := gosax.Token(e)
		if err != nil {
			log.Fatal(err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			fmt.Println("StartElement", t.Name.Local)
			for _, attr := range t.Attr {
				fmt.Println("Attr", attr.Name.Local, attr.Value)
			}
		case xml.EndElement:
			fmt.Println("EndElement", t.Name.Local)
		case xml.CharData:
			fmt.Println("CharData", string(t))
		}
	}
	// Output:
	// StartElement root
	// StartElement element
	// Attr foo <bar>
	// Attr bar qux
	// CharData Value
	// EndElement element
	// EndElement root
}

func ExampleReader_EmitSelfClosingTag() {
	xmlData := `<root><element>Value</element><selfclosing/></root>`
	reader := strings.NewReader(xmlData)

	r := gosax.NewReader(reader)
	r.EmitSelfClosingTag = true
	for {
		e, err := r.Event()
		if err != nil {
			log.Fatal(err)
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		switch e.Type() {
		case gosax.EventStart:
			name, _ := gosax.Name(e.Bytes)
			fmt.Println("EventStart", string(name))
		case gosax.EventEnd:
			name, _ := gosax.Name(e.Bytes)
			fmt.Println("EventEnd", string(name))
		case gosax.EventText:
			fmt.Println("EventText", string(e.Bytes))
		default:
		}
	}
	// Output:
	// EventStart root
	// EventStart element
	// EventText Value
	// EventEnd element
	// EventStart selfclosing
	// EventEnd selfclosing
	// EventEnd root
}

func ExampleUnescape() {
	xmlData := "Line1\r\nLine2\rLine3\nLine4\r\nLine5\r\n"
	b, _ := gosax.Unescape([]byte(xmlData))
	fmt.Printf("%q", string(b))
	// Output:
	// "Line1\nLine2\nLine3\nLine4\nLine5\n"
}

func ExampleStartElement() {
	xmlData := `<root><element foo="bar"
	>
	</element></root>`
	reader := strings.NewReader(xmlData)

	r := gosax.NewReader(reader)
	for {
		e, err := r.Event()
		if err != nil {
			log.Fatal(err)
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		t, err := gosax.Token(e)
		if err != nil {
			log.Fatal(err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			fmt.Println("StartElement", t.Name.Local)
			for _, attr := range t.Attr {
				fmt.Println("Attr", attr.Name.Local, attr.Value)
			}
		case xml.EndElement:
			fmt.Println("EndElement", t.Name.Local)
		case xml.CharData:
			continue
		}
	}
	// Output:
	// StartElement root
	// StartElement element
	// Attr foo bar
	// EndElement element
	// EndElement root
}

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

func ExampleToken() {
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
		t, err := gosax.Token(e)
		if err != nil {
			log.Fatal(err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			fmt.Println("StartElement", t.Name.Local)
		case xml.EndElement:
			fmt.Println("EndElement", t.Name.Local)
		case xml.CharData:
			fmt.Println("CharData", string(t))
		}
	}
	// Output:
	// StartElement root
	// StartElement element
	// CharData Value
	// EndElement element
	// EndElement root
}

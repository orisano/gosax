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

package xmlb_test

import (
	"fmt"
	"strings"

	"github.com/orisano/gosax"
	"github.com/orisano/gosax/xmlb"
)

func Example() {
	r := strings.NewReader(`<root><element>Value</element></root>`)
	gr := gosax.NewReader(r)
	gr.EmitSelfClosingTag = true
	for {
		ev, err := gr.Event()
		if err != nil {
			break
		}
		if ev.Type() == gosax.EventEOF {
			break
		}
		switch t := xmlb.Token(ev); t.Type() {
		case xmlb.StartElement:
			t, _ := t.StartElement()
			fmt.Println("StartElement", t.Name.Local)
		case xmlb.CharData:
			t, _ := t.CharData()
			fmt.Println("CharData", string(t))
		case xmlb.EndElement:
			fmt.Println("EndElement", string(t.Name().Local()))
		}
	}
	// Output:
	// StartElement root
	// StartElement element
	// CharData Value
	// EndElement element
	// EndElement root
}

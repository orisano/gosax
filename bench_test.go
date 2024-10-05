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
	"bytes"
	"os"
	"testing"

	"github.com/orisano/gosax"
)

func BenchmarkReader_Event(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := countAfrica(b); err != nil {
			b.Fatal(err)
		}
	}
}

func countAfrica(b *testing.B) error {
	f, err := os.Open("testdata/out.xml")
	if err != nil {
		return err
	}
	defer f.Close()
	if stat, err := f.Stat(); err == nil {
		b.SetBytes(stat.Size())
	}

	r := gosax.NewReader(f)
	count := 0
	inLocation := false
	for {
		e, err := r.Event()
		if err != nil {
			return err
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		switch e.Type() {
		case gosax.EventStart:
			name, _ := gosax.Name(e.Bytes)
			if string(name) == "location" {
				inLocation = true
			} else {
				inLocation = false
			}
		case gosax.EventEnd:
			inLocation = false
		case gosax.EventText:
			if inLocation {
				if bytes.Contains(e.Bytes, []byte("Africa")) {
					count++
				}
			}
		default:
		}
	}
	return nil
}

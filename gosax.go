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

// Package gosax provides a Simple API for XML (SAX) parser for Go.
// It offers efficient, read-only XML parsing with streaming capabilities,
// inspired by quick-xml and other high-performance parsing techniques.
package gosax

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

const (
	eventUnknown = iota
	EventStart
	EventEnd
	EventText
	EventCData
	EventComment
	EventProcessingInstruction
	EventDocType
	EventEOF
)

type Event struct {
	Bytes []byte
	value uint32
}

func (e Event) Type() uint8 {
	return uint8(e.value)
}

type Reader struct {
	reader byteReader
	state  func(*Reader) (Event, error)
}

func NewReader(r io.Reader) *Reader {
	return NewReaderSize(r, 2*1024*1024)
}

func NewReaderSize(r io.Reader, bufSize int) *Reader {
	return NewReaderBuf(r, make([]byte, 0, bufSize))
}

func NewReaderBuf(r io.Reader, buf []byte) *Reader {
	var xr Reader
	xr.reader.data = buf
	xr.Reset(r)
	return &xr
}

// Event returns the next Event from the XML stream.
// It returns an Event and any error encountered.
//
// Note: The returned Event object is only valid until the next call to Event.
// The underlying byte slice may be overwritten by subsequent calls.
// If you need to retain the Event data, make a copy before the next Event call.
func (r *Reader) Event() (Event, error) {
	return r.state(r)
}

func (r *Reader) Reset(reader io.Reader) {
	data := r.reader.data
	if data != nil {
		data = data[:0]
	}
	r.reader = byteReader{
		data: data,
		r:    reader,
	}
	r.state = (*Reader).stateInit
}

func (r *Reader) stateInit() (Event, error) {
	// remove_utf8_bom
	return r.stateInsideText()
}

func (r *Reader) stateInsideText() (Event, error) {
	end, err := readText(&r.reader)
	if err == io.EOF {
		r.state = (*Reader).stateDone
		if end == 0 {
			return Event{
				value: EventEOF,
			}, nil
		} else {
			w := r.reader.window()
			r.reader.offset += len(w)
			return Event{
				Bytes: w,
				value: EventText,
			}, nil
		}
	}
	if err != nil {
		return Event{}, err
	}
	if end == 0 {
		return r.stateInsideMarkup()
	} else {
		r.state = (*Reader).stateInsideMarkup
		w := r.reader.window()[:end]
		r.reader.offset += len(w)
		return Event{
			Bytes: w,
			value: EventText,
		}, nil
	}
}

func (r *Reader) stateInsideMarkup() (Event, error) {
	r.state = (*Reader).stateInsideText
	rr := &r.reader
	if rr.offset+1 >= len(rr.data) {
		if rr.extend() == 0 {
			return Event{}, rr.err
		}
	}
	switch w := rr.window(); w[1] {
	case '!':
		if len(w) < 3 {
			if rr.extend() == 0 {
				return Event{}, rr.err
			}
			w = rr.window()
		}
		switch w[2] {
		case '[': // CData
			offset := 3
			for {
				if i := bytes.Index(w[offset:], []byte("]]>")); i >= 0 {
					r.reader.offset += offset + i + 3
					return Event{
						Bytes: w[:offset+i+3],
						value: EventCData,
					}, nil
				}
				offset = len(w) - 2
				if rr.extend() == 0 {
					return Event{}, rr.err
				}
				w = rr.window()
			}
		case '-': // Comment
			offset := 3
			for {
				if i := bytes.Index(w[offset:], []byte("-->")); i >= 0 {
					r.reader.offset += offset + i + 3
					return Event{
						Bytes: w[:offset+i+3],
						value: EventComment,
					}, nil
				}
				offset = len(w) - 2
				if rr.extend() == 0 {
					return Event{}, rr.err
				}
				w = rr.window()
			}
		case 'D', 'd': // DocType
			offset := 2
			for {
				lv := 1
				for i, c := range w[offset:] {
					if c == '>' {
						lv--
						if lv == 0 {
							r.reader.offset += offset + i + 1
							return Event{
								Bytes: w[:offset+i+1],
								value: EventDocType,
							}, nil
						}
					} else if c == '<' {
						lv++
					}
				}
				offset = len(w)
				if rr.extend() == 0 {
					return Event{}, rr.err
				}
				w = rr.window()
			}
		default:
			return Event{}, fmt.Errorf("unknown bang type: %c", w[1])
		}
	case '/': // close tag
		offset := 2
		for {
			if i := bytes.IndexByte(w[offset:], '>'); i >= 0 {
				r.reader.offset += offset + i + 1
				return Event{
					Bytes: w[:offset+i+1],
					value: EventEnd,
				}, nil
			}
			offset = len(w)
			if rr.extend() == 0 {
				return Event{}, rr.err
			}
			w = rr.window()
		}
	case '?': // processing instructions
		offset := 2
		for {
			if i := bytes.Index(w[offset:], []byte("?>")); i >= 0 {
				r.reader.offset += offset + i + 2
				return Event{
					Bytes: w[:offset+i+2],
					value: EventProcessingInstruction,
				}, nil
			}
			offset = len(w) - 1
			if rr.extend() == 0 {
				return Event{}, rr.err
			}
			w = rr.window()
		}
	default:
		const (
			splat uint64 = 0x0101010101010101
			v1           = '"' * splat
			v2           = '>' * splat
			v3           = '\'' * splat
		)
		state := byte('>')
		offset := 1
		for {
			for offset < len(w) {
				if state == '>' {
					for ; offset+8 < len(w); offset += 8 {
						v := binary.LittleEndian.Uint64(w[offset : offset+8])
						if hasZeroByte(v^v1) || hasZeroByte(v^v2) || hasZeroByte(v^v3) {
							break
						}
					}
					p := -1
					var ch byte
					for i, c := range w[offset:] {
						if c == '"' || c == '>' || c == '\'' {
							p = i
							ch = c
							break
						}
					}
					if p >= 0 {
						if ch == '>' {
							r.reader.offset += offset + p + 1
							return Event{
								Bytes: w[:offset+p+1],
								value: EventStart,
							}, nil
						} else {
							state = ch
							offset += p + 1
						}
					} else {
						break
					}
				} else {
					if i := bytes.IndexByte(w[offset:], state); i >= 0 {
						offset += i + 1
						state = '>'
					} else {
						break
					}
				}
			}
			offset = len(w)
			if rr.extend() == 0 {
				return Event{}, rr.err
			}
			w = rr.window()
		}
	}
}

func (r *Reader) stateDone() (Event, error) {
	return Event{
		value: EventEOF,
	}, nil
}

func hasZeroByte(x uint64) bool {
	const (
		lo uint64 = 0x0101010101010101
		hi uint64 = 0x8080808080808080
	)
	return (x-lo) & ^x & hi != 0
}

func readText(r *byteReader) (int, error) {
	offset := 0
	for {
		w := r.window()
		if i := bytes.IndexByte(w[offset:], '<'); i >= 0 {
			return offset + i, nil
		}
		offset += len(w)
		if r.extend() == 0 {
			return offset, r.err
		}
	}
}

// Name extracts the name from an XML tag.
// It returns the name and the remaining bytes.
func Name(b []byte) ([]byte, []byte) {
	if len(b) > 1 && b[0] == '<' {
		b = b[1:]
	}
	if len(b) > 1 && b[len(b)-1] == '>' {
		b = b[:len(b)-1]
	}
	for i, c := range b {
		if whitespace[c] {
			return b[:i], b[i+1:]
		}
	}
	return b, nil
}

type Attribute struct {
	Key   []byte
	Value []byte
}

// NextAttribute extracts the next attribute from an XML tag.
// It returns the Attribute and the remaining bytes.
func NextAttribute(b []byte) (Attribute, []byte, error) {
	i := 0
	for ; i < len(b) && whitespace[b[i]]; i++ {
	}
	if i == len(b) {
		return Attribute{}, nil, nil
	}
	keyStart := i
	for ; i < len(b) && !whitespace[b[i]] && b[i] != '='; i++ {
	}
	if i == len(b) {
		return Attribute{Key: b[keyStart:]}, nil, nil
	}
	key := b[keyStart:i]
	for ; i < len(b) && whitespace[b[i]]; i++ {
	}
	if i == len(b) {
		return Attribute{Key: key}, nil, nil
	}
	if b[i] != '=' {
		return Attribute{Key: key}, b[i:], nil
	}
	i++
	for ; i < len(b) && whitespace[b[i]]; i++ {
	}

	if b[i] == '"' {
		valueEnd := i + 1 + bytes.IndexByte(b[i+1:], '"') + 1
		value := b[i:valueEnd]
		return Attribute{Key: key, Value: value}, b[valueEnd:], nil
	}
	if b[i] == '\'' {
		valueEnd := i + 1 + bytes.IndexByte(b[i+1:], '\'') + 1
		value := b[i:valueEnd]
		return Attribute{Key: key, Value: value}, b[valueEnd:], nil
	}
	return Attribute{}, nil, fmt.Errorf("invalid attribute value: %c", b[i])
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
}

// Unescape decodes XML entity references in a byte slice.
// It returns the unescaped bytes and any error encountered.
func Unescape(b []byte) ([]byte, error) {
	p := bytes.IndexByte(b, '&')
	if p < 0 {
		return b, nil
	}
	begin := 0
	cur := p
	for {
		var escaped []byte
		for i := 2; i < 13 && p+i < len(b); i++ {
			if b[p+i] == ';' {
				escaped = b[p+1 : p+i]
				break
			}
		}
		if len(escaped) <= 1 {
			return nil, fmt.Errorf("invalid escape sequence")
		}
		if cur != p && begin != p {
			cur += copy(b[cur:], b[begin:p])
		}
		if escaped[0] == '#' {
			var x uint64
			var err error
			if escaped[1] == 'x' {
				x, err = strconv.ParseUint(string(escaped[2:]), 16, 32)
			} else {
				x, err = strconv.ParseUint(string(escaped[1:]), 10, 32)
			}
			if err != nil {
				return nil, fmt.Errorf("invalid char reference: %w", err)
			}
			cur += utf8.EncodeRune(b[cur:], rune(x))
		} else {
			switch string(escaped) {
			case "lt":
				b[cur] = '<'
			case "gt":
				b[cur] = '>'
			case "amp":
				b[cur] = '&'
			case "apos":
				b[cur] = '\''
			case "quot":
				b[cur] = '"'
			default:
				return nil, fmt.Errorf("invalid escape sequence: %q", string(escaped))
			}
			cur++
		}
		begin = p + len(escaped) + 2
		if i := bytes.IndexByte(b[begin:], '&'); i >= 0 {
			p = begin + i
		} else {
			break
		}
	}
	if len(b) != begin {
		cur += copy(b[cur:], b[begin:])
	}
	return b[:cur], nil
}

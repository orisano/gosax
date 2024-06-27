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

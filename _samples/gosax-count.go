package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/orisano/gosax"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := gosax.NewReader(f)
	count := 0
	inLocation := false
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
	fmt.Println("counter =", count)
}

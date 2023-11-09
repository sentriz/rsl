package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"golang.org/x/tools/txtar"
)

func TestFormats(t *testing.T) {
	paths, err := filepath.Glob("testdata/*.txtar")
	be.NilErr(t, err)

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			runPath(t, path)
		})
	}
}

func runPath(t *testing.T, path string) {
	parts := strings.Split(filepath.Base(path), ".")
	decoder, encoder := DefaultFormats[parts[0]], DefaultFormats[parts[1]]
	be.Nonzero(t, decoder)
	be.Nonzero(t, encoder)

	ar, err := txtar.ParseFile(path)
	be.NilErr(t, err)

	var in, expOut []byte
	for _, f := range ar.Files {
		switch f.Name {
		case "in":
			in = f.Data
		case "out":
			expOut = f.Data
		}
	}
	be.Nonzero(t, in)
	be.Nonzero(t, expOut)

	v, err := decoder.Decode(bytes.NewReader(in))
	be.NilErr(t, err)

	var out bytes.Buffer
	be.NilErr(t, encoder.Encode(&out, v))

	if !bytes.Equal(expOut, out.Bytes()) {
		t.Errorf("\nwant:\n%s\ngot:\n%s", string(expOut), out.String())
	}
}

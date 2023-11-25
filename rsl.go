package main

import (
	"bufio"
	"io"
	"log"
	"os"

	"go.senan.xyz/rsl/csv"
	"go.senan.xyz/rsl/ini"
	"go.senan.xyz/rsl/js"
	"go.senan.xyz/rsl/json"
	"go.senan.xyz/rsl/toml"
	"go.senan.xyz/rsl/xml"
	"go.senan.xyz/rsl/xmlstd"
	"go.senan.xyz/rsl/yaml"
)

type Format interface {
	Encode(w io.Writer, v any) error
	Decode(r io.Reader) (any, error)
}

var DefaultFormats = map[string]Format{
	"csv":     csv.New(false, ','),
	"csv-ph":  csv.New(true, ','),
	"tsv":     csv.New(false, '\t'),
	"tsv-ph":  csv.New(true, '\t'),
	"js":      js.New(),
	"json":    json.New(),
	"toml":    toml.New(),
	"xml":     xml.New(),
	"xml-std": xmlstd.New(),
	"yaml":    yaml.New(),
	"ini":     ini.New(),
}

func main() {
	const (
		cmd = iota
		srcn
		destn
		nargs
	)

	if len(os.Args) < nargs {
		log.Fatalf("usage: %s <src format> <dest format>", os.Args[cmd])
	}

	src, ok := DefaultFormats[os.Args[srcn]]
	if !ok {
		log.Fatalf("unknown src format %q", os.Args[srcn])
	}
	dest, ok := DefaultFormats[os.Args[destn]]
	if !ok {
		log.Fatalf("unknown dest format %q", os.Args[destn])
	}

	v, err := src.Decode(os.Stdin)
	if err != nil {
		log.Fatalf("error decoding to %s: %v", os.Args[srcn], err)
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if err := dest.Encode(out, v); err != nil {
		log.Fatalf("error encoding to %s: %v", os.Args[destn], err)
	}
}

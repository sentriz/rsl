package main

import (
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

var formats = map[string]Format{
	"csv":     csv.New(),
	"csv-ph":  csv.NewWithPseudoHeader(),
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

	src, ok := formats[os.Args[srcn]]
	if !ok {
		log.Fatalf("unknown src format %q", os.Args[srcn])
	}
	dest, ok := formats[os.Args[destn]]
	if !ok {
		log.Fatalf("unknown dest format %q", os.Args[destn])
	}

	v, err := src.Decode(os.Stdin)
	if err != nil {
		log.Fatalf("error decoding to %s: %v", os.Args[srcn], err)
	}
	if err := dest.Encode(os.Stdout, v); err != nil {
		log.Fatalf("error encoding to %s: %v", os.Args[destn], err)
	}
}

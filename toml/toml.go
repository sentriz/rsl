package toml

import (
	"io"
	"reflect"

	"github.com/BurntSushi/toml"
)

type TOML struct{}

func New() *TOML {
	return &TOML{}
}

func (*TOML) Encode(w io.Writer, v any) error {
	// we can't have a top level slice, but it in a map with one key
	if reflect.TypeOf(v).Kind() != reflect.Map {
		v = map[string]any{"result": v}
	}
	return toml.NewEncoder(w).Encode(v)
}

func (*TOML) Decode(r io.Reader) (any, error) {
	var v any
	_, err := toml.NewDecoder(r).Decode(&v)
	return v, err
}

package json

import (
	"encoding/json"
	"io"
)

type JSON struct{}

func New() *JSON {
	return &JSON{}
}

func (*JSON) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (*JSON) Decode(r io.Reader) (any, error) {
	var v any
	return v, json.NewDecoder(r).Decode(&v)
}

package json

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSON struct{}

func New() *JSON {
	return &JSON{}
}

func (*JSON) Encode(w io.Writer, v any) error {
	v = ensureStringKey(v)
	return json.NewEncoder(w).Encode(v)
}

func (*JSON) Decode(r io.Reader) (any, error) {
	var v any
	return v, json.NewDecoder(r).Decode(&v)
}

func ensureStringKey(v any) any {
	switch inner := v.(type) {
	case map[any]any:
		m := map[string]any{}
		for k, v := range inner {
			m[fmt.Sprint(k)] = ensureStringKey(v)
		}
		return m
	case []any:
		for i, v := range inner {
			inner[i] = ensureStringKey(v)
		}
	}
	return v
}

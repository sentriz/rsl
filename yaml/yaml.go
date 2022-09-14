package yaml

import (
	"io"

	"gopkg.in/yaml.v2"
)

func New() *YAML {
	return &YAML{}
}

type YAML struct{}

func (*YAML) Encode(w io.Writer, v any) error {
	return yaml.NewEncoder(w).Encode(v)
}

func (*YAML) Decode(r io.Reader) (any, error) {
	var v any
	return v, yaml.NewDecoder(r).Decode(&v)
}

package xmlstd

import (
	"encoding/xml"
	"io"
)

type XMLStdLib struct{}

func New() *XMLStdLib {
	return &XMLStdLib{}
}

func (*XMLStdLib) Encode(_ any, w io.Writer, v any) error {
	return xml.NewEncoder(w).Encode(v)
}

func (*XMLStdLib) Decode(_ any, r io.Reader) (any, error) {
	var v any
	return v, xml.NewDecoder(r).Decode(&v)
}

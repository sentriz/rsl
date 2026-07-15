package json

import (
	"encoding/json"
	"fmt"
	"io"

	"go.senan.xyz/rsl/omap"
)

type JSON struct{}

func New() *JSON {
	return &JSON{}
}

func (*JSON) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (*JSON) Decode(r io.Reader) (any, error) {
	return decodeValue(json.NewDecoder(r))
}

func decodeValue(d *json.Decoder) (any, error) {
	tok, err := d.Token()
	if err != nil {
		return nil, err
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}

	switch delim {
	case '{':
		m := omap.New()
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return nil, err
			}
			v, err := decodeValue(d)
			if err != nil {
				return nil, err
			}
			m.Set(fmt.Sprint(key), v)
		}
		_, err := d.Token()
		return m, err

	case '[':
		var s []any
		for d.More() {
			v, err := decodeValue(d)
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		_, err := d.Token()
		return s, err

	default:
		return nil, fmt.Errorf("unexpected delim %v", delim)
	}
}

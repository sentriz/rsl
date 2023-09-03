package ini

import (
	"fmt"
	"io"

	"gopkg.in/ini.v1"
)

func New() *INI {
	return &INI{}
}

type INI struct{}

func (*INI) Encode(w io.Writer, v any) error {
	return fmt.Errorf("encoding to ini is not supported")
}

func (*INI) Decode(r io.Reader) (any, error) {
	file, err := ini.Load(r)
	if err != nil {
		return nil, fmt.Errorf("ini load: %w", err)
	}

	v := map[string]map[string]string{}
	for _, sec := range file.Sections() {
		if len(sec.Keys()) == 0 {
			continue
		}
		v[sec.Name()] = sec.KeysHash()
	}
	return v, nil
}

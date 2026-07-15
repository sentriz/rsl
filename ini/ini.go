package ini

import (
	"fmt"
	"io"

	"go.senan.xyz/rsl/omap"
	"gopkg.in/ini.v1"
)

type INI struct{}

func New() *INI {
	return &INI{}
}

func (*INI) Encode(w io.Writer, v any) error {
	m, ok := v.(*omap.Map)
	if !ok {
		m = omap.New()
		m.Set("result", v)
	}

	file := ini.Empty()
	for k, v := range m.All() {
		switch sec := v.(type) {
		case *omap.Map:
			for kk, vv := range sec.All() {
				_, _ = file.Section(k).NewKey(kk, fmt.Sprint(vv))
			}
		default:
			_, _ = file.Section("").NewKey(k, fmt.Sprint(v))
		}
	}

	if _, err := file.WriteTo(w); err != nil {
		return fmt.Errorf("write to: %w", err)
	}
	return nil
}

func (*INI) Decode(r io.Reader) (any, error) {
	file, err := ini.Load(r)
	if err != nil {
		return nil, fmt.Errorf("ini load: %w", err)
	}

	root := omap.New()
	for _, sec := range file.Sections() {
		if len(sec.Keys()) == 0 {
			continue
		}
		m := omap.New()
		for _, key := range sec.Keys() {
			m.Set(key.Name(), key.Value())
		}
		root.Set(sec.Name(), m)
	}
	return root, nil
}

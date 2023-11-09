package ini

import (
	"fmt"
	"io"
	"reflect"

	"gopkg.in/ini.v1"
)

func New() *INI {
	return &INI{}
}

type INI struct{}

func (*INI) Encode(w io.Writer, v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Map {
		return fmt.Errorf("can't handle maps")
	}
	contents, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("can only support maps currently")
	}

	file := ini.Empty()
	for k, v := range contents {
		switch cont := v.(type) {
		case map[string]any:
			for kk, vv := range cont {
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

	val := map[string]map[string]string{}
	for _, sec := range file.Sections() {
		if len(sec.Keys()) == 0 {
			continue
		}
		val[sec.Name()] = sec.KeysHash()
	}
	return val, nil
}

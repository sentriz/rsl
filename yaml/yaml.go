package yaml

import (
	"fmt"
	"io"
	"maps"
	"slices"

	"go.senan.xyz/rsl/omap"
	"gopkg.in/yaml.v2"
)

type YAML struct{}

func New() *YAML {
	return &YAML{}
}

func (*YAML) Encode(w io.Writer, v any) error {
	return yaml.NewEncoder(w).Encode(v)
}

func (*YAML) Decode(r io.Reader) (any, error) {
	var n node
	if err := yaml.NewDecoder(r).Decode(&n); err != nil {
		return nil, err
	}
	return n.v, nil
}

type node struct {
	v any
}

func (n *node) UnmarshalYAML(unmarshal func(any) error) error {
	var probe any
	if err := unmarshal(&probe); err != nil {
		return err
	}

	switch probe.(type) {
	case map[any]any:
		var ms yaml.MapSlice
		if err := unmarshal(&ms); err != nil {
			return err
		}
		v := convert(ms)
		graftMergedKeys(v, probe)
		n.v = v

	case []any:
		var s []node
		if err := unmarshal(&s); err != nil {
			return err
		}
		vs := make([]any, 0, len(s))
		for _, e := range s {
			vs = append(vs, e.v)
		}
		n.v = vs

	default:
		n.v = probe
	}
	return nil
}

func convert(v any) any {
	switch t := v.(type) {
	case yaml.MapSlice:
		m := omap.New()
		for _, item := range t {
			m.Set(fmt.Sprint(item.Key), convert(item.Value))
		}
		return m
	case []any:
		for i, e := range t {
			t[i] = convert(e)
		}
	}
	return v
}

func graftMergedKeys(v, p any) {
	switch t := v.(type) {
	case *omap.Map:
		pm, ok := p.(map[any]any)
		if !ok {
			return
		}
		missing := map[string]any{}
		for k, pv := range pm {
			ks := fmt.Sprint(k)
			if cur, ok := t.Get(ks); ok {
				graftMergedKeys(cur, pv)
				continue
			}
			missing[ks] = pv
		}
		for _, k := range slices.Sorted(maps.Keys(missing)) {
			t.Set(k, convertPlain(missing[k]))
		}
	case []any:
		ps, ok := p.([]any)
		if !ok {
			return
		}
		for i := range t {
			if i < len(ps) {
				graftMergedKeys(t[i], ps[i])
			}
		}
	}
}

func convertPlain(v any) any {
	switch t := v.(type) {
	case map[any]any:
		byName := map[string]any{}
		for k, v := range t {
			byName[fmt.Sprint(k)] = v
		}
		m := omap.New()
		for _, k := range slices.Sorted(maps.Keys(byName)) {
			m.Set(k, convertPlain(byName[k]))
		}
		return m
	case []any:
		for i, e := range t {
			t[i] = convertPlain(e)
		}
	}
	return v
}

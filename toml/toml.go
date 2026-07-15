package toml

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/BurntSushi/toml"
	"go.senan.xyz/rsl/omap"
)

type TOML struct{}

func New() *TOML {
	return &TOML{}
}

func (*TOML) Encode(w io.Writer, v any) error {
	v = unconvert(v)
	// we can't have a top level slice, put it in a map with one key
	if _, ok := v.(map[string]any); !ok {
		v = map[string]any{"result": v}
	}
	return toml.NewEncoder(w).Encode(v)
}

func (*TOML) Decode(r io.Reader) (any, error) {
	var v any
	meta, err := toml.NewDecoder(r).Decode(&v)
	if err != nil {
		return nil, err
	}

	if m, ok := v.(map[string]any); ok {
		return convertMap(m, keyOrders(meta), ""), nil
	}
	return v, nil
}

func unconvert(v any) any {
	switch t := v.(type) {
	case *omap.Map:
		m := make(map[string]any, t.Len())
		for k, v := range t.All() {
			m[k] = unconvert(v)
		}
		return m
	case []any:
		for i, e := range t {
			t[i] = unconvert(e)
		}
	}
	return v
}

func keyOrders(meta toml.MetaData) map[string][]string {
	order := map[string][]string{}
	count := map[string]int{}
	for _, key := range meta.Keys() {
		path := ""
		for i, part := range key {
			full := childPath(path, part)
			if i < len(key)-1 {
				path = elemPath(full, count[full])
				continue
			}
			if slices.Contains(order[path], part) {
				count[full]++
			} else {
				order[path] = append(order[path], part)
			}
		}
	}
	return order
}

func convertMap(t map[string]any, order map[string][]string, path string) *omap.Map {
	pos := map[string]int{}
	for i, k := range order[path] {
		pos[k] = i
	}
	rank := func(k string) int {
		if p, ok := pos[k]; ok {
			return p
		}
		return len(pos)
	}

	keys := slices.Sorted(maps.Keys(t))
	slices.SortStableFunc(keys, func(a, b string) int {
		return cmp.Compare(rank(a), rank(b))
	})

	m := omap.New()
	for _, k := range keys {
		m.Set(k, convertValue(t[k], order, childPath(path, k)))
	}
	return m
}

func convertValue(v any, order map[string][]string, base string) any {
	switch t := v.(type) {
	case map[string]any:
		return convertMap(t, order, elemPath(base, 0))
	case []map[string]any:
		s := make([]any, 0, len(t))
		for i, e := range t {
			s = append(s, convertMap(e, order, elemPath(base, i)))
		}
		return s
	case []any:
		for i, e := range t {
			t[i] = convertValue(e, order, base)
		}
	}
	return v
}

func childPath(parent, key string) string {
	return parent + "\x00" + key
}

func elemPath(path string, i int) string {
	return fmt.Sprintf("%s#%d", path, i)
}

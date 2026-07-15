package omap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"gopkg.in/yaml.v2"
)

type Map struct {
	keys []string
	m    map[string]any
}

func New() *Map {
	return &Map{m: map[string]any{}}
}

func (m *Map) Set(k string, v any) {
	if _, ok := m.m[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.m[k] = v
}

func (m *Map) Add(k string, v any) {
	cur, ok := m.Get(k)
	if !ok {
		m.Set(k, v)
		return
	}
	if s, ok := cur.([]any); ok {
		m.Set(k, append(s, v))
		return
	}
	m.Set(k, []any{cur, v})
}

func (m *Map) Get(k string) (any, bool) {
	v, ok := m.m[k]
	return v, ok
}

func (m *Map) Keys() []string {
	return m.keys
}

func (m *Map) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for _, k := range m.keys {
			if !yield(k, m.m[k]) {
				return
			}
		}
	}
}

func (m *Map) Len() int {
	return len(m.keys)
}

func (m *Map) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(m.m[k])
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (m *Map) MarshalYAML() (any, error) {
	ms := make(yaml.MapSlice, 0, len(m.keys))
	for _, k := range m.keys {
		ms = append(ms, yaml.MapItem{Key: k, Value: m.m[k]})
	}
	return ms, nil
}

func (m *Map) String() string {
	var b strings.Builder
	b.WriteString("map[")
	for i, k := range m.keys {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%v:%v", k, m.m[k])
	}
	b.WriteByte(']')
	return b.String()
}

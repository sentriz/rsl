package json

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type JSON struct{}

func New() *JSON {
	return &JSON{}
}

func (*JSON) Encode(w io.Writer, v any) error {
	ensureStringKey(&v)
	return json.NewEncoder(w).Encode(v)
}

func (*JSON) Decode(r io.Reader) (any, error) {
	var v any
	return v, json.NewDecoder(r).Decode(&v)
}

func ensureStringKey(v any) {
	rv := reflect.ValueOf(v)
	switch inner := maybeElem(rv); inner.Kind() {
	case reflect.Map:
		if inner.Type().Key().Kind() == reflect.Interface {
			m := map[string]any{}
			iter := inner.MapRange()
			for iter.Next() {
				ensureStringKey(maybeAddr(iter.Value()).Interface())
				m[fmt.Sprint(iter.Key())] = iter.Value().Interface()
			}
			rv.Elem().Set(reflect.ValueOf(m))
		}
	case reflect.Slice:
		for i := 0; i < inner.Len(); i++ {
			ensureStringKey(maybeAddr(inner.Index(i)).Interface())
		}
	}
}

func maybeElem(rv reflect.Value) reflect.Value {
	for {
		switch rv.Kind() {
		case reflect.Interface, reflect.Pointer:
			rv = rv.Elem()
		default:
			return rv
		}
	}
}

func maybeAddr(rv reflect.Value) reflect.Value {
	if rv.CanAddr() {
		rv = rv.Addr()
	}
	return rv
}

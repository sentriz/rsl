package js

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/robertkrimen/otto"
	"go.senan.xyz/rsl/omap"
)

type JS struct{}

func New() *JS {
	return &JS{}
}

func (*JS) Encode(w io.Writer, v any) error {
	return errors.ErrUnsupported
}

func (*JS) Decode(r io.Reader) (any, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	vm := otto.New()

	obj, err := vm.Object(fmt.Sprintf(`(%s)`, input))
	if err != nil {
		return nil, fmt.Errorf("create object: %w", err)
	}

	return convert(obj.Value())
}

func convert(v otto.Value) (any, error) {
	if !v.IsObject() {
		return v.Export()
	}

	obj := v.Object()
	if obj.Class() == "Array" {
		length, _ := obj.Get("length")
		n, _ := length.ToInteger()
		s := make([]any, 0, n)
		for i := range n {
			el, err := obj.Get(strconv.FormatInt(i, 10))
			if err != nil {
				return nil, err
			}
			v, err := convert(el)
			if err != nil {
				return nil, err
			}
			s = append(s, v)
		}
		return s, nil
	}

	m := omap.New()
	for _, k := range obj.Keys() {
		el, err := obj.Get(k)
		if err != nil {
			return nil, err
		}
		v, err := convert(el)
		if err != nil {
			return nil, err
		}
		m.Set(k, v)
	}
	return m, nil
}

package js

import (
	"errors"
	"fmt"
	"io"

	"github.com/robertkrimen/otto"
)

type JavaScript struct{}

func New() *JavaScript {
	return &JavaScript{}
}

func (*JavaScript) Encode(_ any, w io.Writer, v any) error {
	return errors.ErrUnsupported
}

func (*JavaScript) Decode(_ any, r io.Reader) (any, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	vm := otto.New()

	obj, err := vm.Object(fmt.Sprintf(`(%s)`, input))
	if err != nil {
		return nil, fmt.Errorf("create object: %w", err)
	}

	return obj.Value().Export()
}

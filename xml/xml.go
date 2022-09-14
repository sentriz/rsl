package xml

import (
	"io"

	"github.com/sbabiv/xml2map"
	"go.senan.xyz/rsl/xmlstd"
)

type XML struct {
	*xmlstd.XMLStdLib
}

func New() *XML {
	return &XML{XMLStdLib: xmlstd.New()}
}

func (*XML) Decode(r io.Reader) (any, error) {
	v, err := xml2map.NewDecoder(r).Decode()
	return v, err
}

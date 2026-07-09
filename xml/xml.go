package xml

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

type XML struct{}

func New() *XML {
	return &XML{}
}

func (*XML) Encode(_ any, w io.Writer, v any) error {
	return xml.NewEncoder(w).Encode(v)
}

func (*XML) Decode(_ any, r io.Reader) (any, error) {
	d := xml.NewDecoder(r)
	d.Strict = false
	d.Entity = xml.HTMLEntity

	root := map[string]any{}
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			return root, nil
		}
		if err != nil {
			return nil, err
		}
		if t, ok := tok.(xml.StartElement); ok {
			v, err := decodeElement(d, t)
			if err != nil {
				return nil, err
			}
			addChild(root, t.Name.Local, v)
		}
	}
}

func decodeElement(d *xml.Decoder, start xml.StartElement) (any, error) {
	m := map[string]any{}
	for _, attr := range start.Attr {
		m["@"+attr.Name.Local] = attr.Value
	}

	var texts []string
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			v, err := decodeElement(d, t)
			if err != nil {
				return nil, err
			}
			addChild(m, t.Name.Local, v)
		case xml.CharData:
			if s := strings.TrimSpace(string(t)); s != "" {
				texts = append(texts, s)
			}
		case xml.EndElement:
			if len(m) == 0 {
				return strings.Join(texts, " "), nil
			}
			if len(texts) > 0 {
				m["#text"] = strings.Join(texts, " ")
			}
			return m, nil
		}
	}
}

func addChild(m map[string]any, key string, v any) {
	switch cur := m[key].(type) {
	case nil:
		m[key] = v
	case []any:
		m[key] = append(cur, v)
	default:
		m[key] = []any{cur, v}
	}
}

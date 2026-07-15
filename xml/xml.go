package xml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.senan.xyz/rsl/omap"
)

type XML struct{}

func New() *XML {
	return &XML{}
}

func (*XML) Encode(w io.Writer, v any) error {
	m, ok := v.(*omap.Map)
	if !ok {
		m = omap.New()
		m.Set("result", v)
	}

	e := xml.NewEncoder(w)
	for k, v := range m.All() {
		if err := encodeElement(e, k, v); err != nil {
			return err
		}
	}
	if err := e.Flush(); err != nil {
		return err
	}
	_, err := w.Write([]byte{'\n'})
	return err
}

func (*XML) Decode(r io.Reader) (any, error) {
	d := xml.NewDecoder(r)
	d.Strict = false
	d.Entity = xml.HTMLEntity

	root := omap.New()
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
			root.Add(t.Name.Local, v)
		}
	}
}

func encodeElement(e *xml.Encoder, name string, v any) error {
	if !validName(name) {
		return fmt.Errorf("invalid element name %q", name)
	}

	if s, ok := v.([]any); ok {
		for _, el := range s {
			if err := encodeElement(e, name, el); err != nil {
				return err
			}
		}
		return nil
	}

	start := xml.StartElement{Name: xml.Name{Local: name}}

	m, ok := v.(*omap.Map)
	if !ok {
		if err := e.EncodeToken(start); err != nil {
			return err
		}
		if v != nil {
			if err := e.EncodeToken(xml.CharData(fmt.Sprint(v))); err != nil {
				return err
			}
		}
		return e.EncodeToken(start.End())
	}

	for k, v := range m.All() {
		if strings.HasPrefix(k, "@") {
			if !validName(k[1:]) {
				return fmt.Errorf("invalid attribute name %q", k[1:])
			}
			start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: k[1:]}, Value: fmt.Sprint(v)})
		}
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for k, v := range m.All() {
		switch {
		case strings.HasPrefix(k, "@"):
		case k == "#text":
			if err := e.EncodeToken(xml.CharData(fmt.Sprint(v))); err != nil {
				return err
			}
		default:
			if err := encodeElement(e, k, v); err != nil {
				return err
			}
		}
	}
	return e.EncodeToken(start.End())
}

func validName(s string) bool {
	return s != "" && !strings.ContainsAny(s, " \t\r\n<>&'\"=/")
}

func decodeElement(d *xml.Decoder, start xml.StartElement) (any, error) {
	m := omap.New()
	for _, attr := range start.Attr {
		m.Set("@"+attr.Name.Local, attr.Value)
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
			m.Add(t.Name.Local, v)
		case xml.CharData:
			if s := strings.TrimSpace(string(t)); s != "" {
				texts = append(texts, s)
			}
		case xml.EndElement:
			if m.Len() == 0 {
				return strings.Join(texts, " "), nil
			}
			if len(texts) > 0 {
				m.Set("#text", strings.Join(texts, " "))
			}
			return m, nil
		}
	}
}

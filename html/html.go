package html

import (
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type HTML struct{}

func New() *HTML {
	return &HTML{}
}

func (*HTML) Encode(_ any, w io.Writer, v any) error {
	return errors.ErrUnsupported
}

func (*HTML) Decode(_ any, r io.Reader) (any, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	root := map[string]any{}
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			addChild(root, c.Data, convert(c))
		}
	}
	return root, nil
}

func convert(n *html.Node) any {
	m := map[string]any{}
	for _, attr := range n.Attr {
		m["@"+attr.Key] = attr.Val
	}

	var texts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			if t := strings.TrimSpace(c.Data); t != "" {
				texts = append(texts, t)
			}
		case html.ElementNode:
			addChild(m, c.Data, convert(c))
		}
	}

	if len(m) == 0 {
		return strings.Join(texts, " ")
	}
	if len(texts) > 0 {
		m["#text"] = strings.Join(texts, " ")
	}
	return m
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

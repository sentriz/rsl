package html

import (
	"errors"
	"io"
	"strings"

	"go.senan.xyz/rsl/omap"
	"golang.org/x/net/html"
)

type HTML struct{}

func New() *HTML {
	return &HTML{}
}

func (*HTML) Encode(w io.Writer, v any) error {
	return errors.ErrUnsupported
}

func (*HTML) Decode(r io.Reader) (any, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	root := omap.New()
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			root.Add(c.Data, convert(c))
		}
	}
	return root, nil
}

func convert(n *html.Node) any {
	m := omap.New()
	for _, attr := range n.Attr {
		m.Set("@"+attr.Key, attr.Val)
	}

	var texts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			if t := strings.TrimSpace(c.Data); t != "" {
				texts = append(texts, t)
			}
		case html.ElementNode:
			m.Add(c.Data, convert(c))
		}
	}

	if m.Len() == 0 {
		return strings.Join(texts, " ")
	}
	if len(texts) > 0 {
		m.Set("#text", strings.Join(texts, " "))
	}
	return m
}

package md

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"go.senan.xyz/rsl/omap"
)

type MD struct{}

func New() *MD {
	return &MD{}
}

func (*MD) Encode(w io.Writer, v any) error {
	return encode(w, v, 1)
}

func encode(w io.Writer, v any, level int) error {
	switch v := v.(type) {
	case *omap.Map:
		for k, val := range v.All() {
			fmt.Fprintf(w, "%s %s\n\n", strings.Repeat("#", min(level, 6)), k)
			if err := encode(w, val, level+1); err != nil {
				return err
			}
		}
	case []any:
		if len(v) > 0 && writeTable(w, v) {
			return nil
		}
		for _, item := range v {
			fmt.Fprintf(w, "- %s\n", escape(fmt.Sprint(item)))
		}
		fmt.Fprintln(w)
	default:
		fmt.Fprintf(w, "%v\n\n", v)
	}
	return nil
}

func writeTable(w io.Writer, rows []any) bool {
	var header []string
	var body [][]string

	switch rows[0].(type) {
	case *omap.Map:
		seen := map[string]bool{}
		for _, r := range rows {
			m, ok := r.(*omap.Map)
			if !ok {
				return false
			}
			for _, k := range m.Keys() {
				if !seen[k] {
					seen[k] = true
					header = append(header, k)
				}
			}
		}
		for _, r := range rows {
			m := r.(*omap.Map)
			row := make([]string, 0, len(header))
			for _, head := range header {
				if v, ok := m.Get(head); ok {
					row = append(row, fmt.Sprint(v))
				} else {
					row = append(row, "")
				}
			}
			body = append(body, row)
		}

	case []any:
		var width int
		for _, r := range rows {
			cells, ok := r.([]any)
			if !ok {
				return false
			}
			width = max(width, len(cells))
		}
		header = pseudoHeader(width)
		for _, r := range rows {
			cells := r.([]any)
			row := make([]string, width)
			for j := range cells {
				row[j] = fmt.Sprint(cells[j])
			}
			body = append(body, row)
		}

	default:
		return false
	}

	writeRow(w, header)
	writeSeparator(w, len(header))
	for _, row := range body {
		writeRow(w, row)
	}
	fmt.Fprintln(w)
	return true
}

func (*MD) Decode(r io.Reader) (any, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	parser := goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough)).Parser()
	doc := parser.Parse(text.NewReader(source))
	return convertBlocks(doc, source), nil
}

func convertBlocks(parent ast.Node, source []byte) []any {
	items := []any{}
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		if v, ok := convertBlock(c, source); ok {
			items = append(items, v)
		}
	}
	return items
}

func convertBlock(n ast.Node, source []byte) (any, bool) {
	block := func(k string, v any) (any, bool) {
		m := omap.New()
		m.Set(k, v)
		return m, true
	}

	switch n := n.(type) {
	case *ast.Heading:
		return block(fmt.Sprintf("h%d", n.Level), inlineContent(n, source))
	case *ast.Paragraph, *ast.TextBlock:
		return block("p", inlineContent(n, source))
	case *ast.FencedCodeBlock:
		if lang := n.Language(source); lang != nil {
			m := omap.New()
			m.Set("@lang", string(lang))
			m.Set("#text", blockLines(n, source))
			return block("code", m)
		}
		return block("code", blockLines(n, source))
	case *ast.CodeBlock:
		return block("code", blockLines(n, source))
	case *ast.Blockquote:
		return block("blockquote", convertBlocks(n, source))
	case *ast.List:
		tag := "ul"
		if n.IsOrdered() {
			tag = "ol"
		}
		items := []any{}
		for li := n.FirstChild(); li != nil; li = li.NextSibling() {
			items = append(items, convertListItem(li, source))
		}
		return block(tag, items)
	case *east.Table:
		return block("table", tableRows(n, source))
	case *ast.HTMLBlock:
		return block("html", blockLines(n, source))
	}
	return nil, false
}

func convertListItem(li ast.Node, source []byte) any {
	blocks := []any{}
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *ast.TextBlock, *ast.Paragraph:
			blocks = append(blocks, inlineContent(c, source))
		default:
			if v, ok := convertBlock(c, source); ok {
				blocks = append(blocks, v)
			}
		}
	}
	if len(blocks) == 1 {
		return blocks[0]
	}
	return blocks
}

func tableRows(table *east.Table, source []byte) []any {
	var header []string
	rows := []any{}
	for r := table.FirstChild(); r != nil; r = r.NextSibling() {
		var cells []any
		for c := r.FirstChild(); c != nil; c = c.NextSibling() {
			cells = append(cells, inlineContent(c, source))
		}
		if _, ok := r.(*east.TableHeader); ok {
			header = header[:0]
			for _, c := range cells {
				header = append(header, fmt.Sprint(c))
			}
			continue
		}
		row := omap.New()
		for i := range header {
			if i < len(cells) {
				row.Set(header[i], cells[i])
			} else {
				row.Set(header[i], "")
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func inlineContent(n ast.Node, source []byte) any {
	parts := []any{}
	var sb strings.Builder
	flush := func() {
		if t := strings.TrimSpace(sb.String()); t != "" {
			parts = append(parts, t)
		}
		sb.Reset()
	}

	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c := c.(type) {
		case *ast.Link:
			flush()
			parts = append(parts, link(string(c.Destination), inlineText(c, source)))
			return ast.WalkSkipChildren, nil
		case *ast.AutoLink:
			flush()
			url := string(c.URL(source))
			parts = append(parts, link(url, url))
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			sb.Write(c.Value(source))
			if c.SoftLineBreak() || c.HardLineBreak() {
				sb.WriteByte(' ')
			}
		}
		return ast.WalkContinue, nil
	})
	flush()

	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	return parts
}

func link(href, text string) *omap.Map {
	a := omap.New()
	a.Set("@href", href)
	a.Set("#text", text)
	m := omap.New()
	m.Set("a", a)
	return m
}

func inlineText(n ast.Node, source []byte) string {
	var sb strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := c.(*ast.Text); ok {
			sb.Write(t.Value(source))
			if t.SoftLineBreak() || t.HardLineBreak() {
				sb.WriteByte(' ')
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(sb.String())
}

func blockLines(n ast.Node, source []byte) string {
	var sb strings.Builder
	lines := n.Lines()
	for i := range lines.Len() {
		seg := lines.At(i)
		sb.Write(seg.Value(source))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func writeRow(w io.Writer, cells []string) {
	for _, c := range cells {
		fmt.Fprintf(w, "| %s ", escape(c))
	}
	fmt.Fprintln(w, "|")
}

func writeSeparator(w io.Writer, n int) {
	fmt.Fprintln(w, strings.Repeat("| --- ", n)+"|")
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func pseudoHeader(n int) []string {
	header := make([]string, 0, n)
	for i := range n {
		header = append(header, fmt.Sprintf("%c", 'a'+i))
	}
	return header
}

package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"go.senan.xyz/rsl/omap"
)

type CSV struct {
	addPseudoHeader bool
	delimiter       rune
}

func New(addPseudoHeader bool, delimiter rune) *CSV {
	return &CSV{addPseudoHeader: addPseudoHeader, delimiter: delimiter}
}

func (c *CSV) Encode(w io.Writer, v any) error {
	rows, ok := v.([]any)
	if !ok {
		rows = []any{v}
	}
	if len(rows) == 0 {
		return nil
	}

	writer := csv.NewWriter(w)
	writer.Comma = c.delimiter
	defer writer.Flush()

	switch first := rows[0].(type) {
	case *omap.Map:
		header := first.Keys()
		_ = writer.Write(header)
		for _, r := range rows {
			m, ok := r.(*omap.Map)
			if !ok {
				return fmt.Errorf("mixed row types %T", r)
			}
			row := make([]string, 0, len(header))
			for _, head := range header {
				if v, ok := m.Get(head); ok {
					row = append(row, fmt.Sprint(v))
				} else {
					row = append(row, "")
				}
			}
			_ = writer.Write(row)
		}

	case []any:
		_ = writer.Write(pseudoHeader(len(first)))
		for _, r := range rows {
			cells, ok := r.([]any)
			if !ok {
				return fmt.Errorf("mixed row types %T", r)
			}
			row := make([]string, len(first))
			for j := range row {
				if j < len(cells) {
					row[j] = fmt.Sprint(cells[j])
				}
			}
			_ = writer.Write(row)
		}

	default:
		_ = writer.Write([]string{"result"})
		for _, r := range rows {
			_ = writer.Write([]string{fmt.Sprint(r)})
		}
	}

	return writer.Error()
}

func (c *CSV) Decode(r io.Reader) (any, error) {
	var header []string
	var rows []any

	addRow := func(raw []string) {
		row := omap.New()
		for i := range header {
			row.Set(header[i], raw[i])
		}
		rows = append(rows, row)
	}

	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.Comma = c.delimiter

	firstRow, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read first row: %w", err)
	}

	if c.addPseudoHeader {
		header = pseudoHeader(len(firstRow))
		addRow(firstRow)
	} else {
		header = firstRow
	}

	for {
		raw, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		addRow(raw)
	}

	return rows, nil
}

func pseudoHeader(n int) []string {
	header := make([]string, 0, n)
	for i := range n {
		header = append(header, fmt.Sprintf("%c", 'a'+i))
	}
	return header
}

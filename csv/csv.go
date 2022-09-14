package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
)

type CSV struct{}

func New() *CSV {
	return &CSV{}
}

func (*CSV) Encode(w io.Writer, v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		v = []any{v}
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	switch rv := reflect.ValueOf(v); elemOr(rv.Index(0)).Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
		writer.Write([]string{"result"})
		for i := 0; i < rv.Len(); i++ {
			writer.Write([]string{fmt.Sprint(rv.Index(i))})
		}

	case reflect.Map:
		var header []string
		for _, v := range elemOr(rv.Index(0)).MapKeys() {
			header = append(header, v.String())
		}
		sort.Strings(header)
		writer.Write(header)
		for i := 0; i < rv.Len(); i++ {
			var row []string
			for _, head := range header {
				row = append(row, fmt.Sprint(elemOr(rv.Index(i)).MapIndex(reflect.ValueOf(head))))
			}
			writer.Write(row)
		}

	case reflect.Slice:
		var header []string
		for i := 0; i < elemOr(rv.Index(0)).Len(); i++ {
			header = append(header, fmt.Sprintf("%c", 'a'+i))
		}
		writer.Write(header)
		for i := 0; i < rv.Len(); i++ {
			var row []string
			for j := range header {
				row = append(row, fmt.Sprint(rv.Index(i).Elem().Index(j)))
			}
			writer.Write(row)
		}

	default:
		return fmt.Errorf("unsupported type %T", v)
	}

	return writer.Error()
}

func (*CSV) Decode(r io.Reader) (any, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reader header: %v", err)
	}

	var rows []map[string]string
	for {
		rawrow, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading row: %v", err)
		}

		row := map[string]string{}
		for i := range header {
			row[header[i]] = rawrow[i]
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func elemOr(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Interface {
		return rv.Elem()
	}
	return rv
}

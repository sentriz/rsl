package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"go.senan.xyz/rsl/csv"
	"go.senan.xyz/rsl/js"
	"go.senan.xyz/rsl/json"
	"go.senan.xyz/rsl/toml"
	"go.senan.xyz/rsl/yaml"
)

func TestEncodeMatrix(t *testing.T) {
	formats := []Format{
		csv.New(),
		json.New(),
		toml.New(),
		yaml.New(),
	}
	data := []any{
		1,
		1.0,
		"",
		true,
		[]any{1},
		[]float64{1.0},
		[]string{""},
		[]bool{true},
		map[string]string{"a": "one", "b": "two"},
		[]string{"one", "two"},
		map[string]any{"a": "one", "b": 2},
		map[string]map[string]int{"a": {"a": 1}, "b": {"b": 3}},
	}

	for _, d := range data {
		for _, f := range formats {
			t.Run(fmt.Sprintf("%T %T", f, d), func(t *testing.T) {
				be.NilErr(t, f.Encode(io.Discard, d))
			})
		}
	}
}

func TestJSONCSV(t *testing.T) {
	js := js.New()
	json := json.New()
	csv := csv.New()

	testLine(t,
		json, []string{`1`},
		csv, []string{"result", "1"},
	)
	testLine(t,
		json, []string{`"one"`},
		csv, []string{"result", "one"},
	)
	testLine(t,
		json, []string{`[1, "two", 3]`},
		csv, []string{"result", "1", "two", "3"},
	)
	testLine(t,
		json, []string{`[[1, 2, 3]]`},
		csv, []string{"a,b,c", "1,2,3"},
	)
	testLine(t,
		json, []string{`[[1, 2, 3],  [10, 20, 30]]`},
		csv, []string{"a,b,c", "1,2,3", "10,20,30"},
	)
	testLine(t,
		json, []string{`[{"age": 69, "name": "joe"}, {"age": 100, "name": "mick"}]`},
		csv, []string{"age,name", "69,joe", "100,mick"},
	)
	testLine(t,
		js, []string{`[{age: 69, name: "joe"}, {age: 100, name: "mick"}]`},
		csv, []string{"age,name", "69,joe", "100,mick"},
	)
	testLine(t,
		js, []string{`[{age: 69, name: {first: "joe", last: "davis"}}, {age: 100, name: {first: "mick", last: "jones"}}]`},
		csv, []string{"age,name", "69,map[first:joe last:davis]", "100,map[first:mick last:jones]"},
	)
}

func TestYAMLJSONNullKey(t *testing.T) {
	yaml := yaml.New()
	json := json.New()

	testLine(t,
		yaml, []string{`men: [John Smith, Bill Jones]`, `women:`, `  - Mary Smith`, `  - Susan Williams`},
		json, []string{`{"men":["John Smith","Bill Jones"],"women":["Mary Smith","Susan Williams"]}`},
	)
}

func testLine(t *testing.T, inFormat Format, in []string, outFormat Format, out []string) {
	v, err := inFormat.Decode(bytes.NewReader([]byte(strings.Join(in, "\n"))))
	be.NilErr(t, err)
	var buff bytes.Buffer
	be.NilErr(t, outFormat.Encode(&buff, v))

	for _, exp := range out {
		got, err := buff.ReadString('\n')
		be.NilErr(t, err)
		be.Equal(t, strings.TrimSpace(exp), strings.TrimSpace(got))
	}
}

package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/danrl/golibby/powerset"
	minify "github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

func main() {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	fatal(
		compile(m, "./tracker/src/yukio.js", "js/yukio.js", nil),
	)

	a := []string{"hash", "outbound-links", "exclusions"}
	for _, v := range pwset(a) {
		d := strings.Join(append([]string{"yukio"}, v...), ".")
		data := map[string]interface{}{}
		for _, n := range v {
			data[strings.ReplaceAll(n, "-", "_")] = true
		}
		fatal(
			compile(m, "./tracker/src/yukio.js", "js/"+d+".js", data),
		)
	}
}

func fatal(errs ...error) {
	for _, err := range errs {
		if err != nil {
			log.Fatal(err)
		}
	}
}

func compile(m *minify.M, src, dest string, data interface{}) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	t := template.Must(template.New("x").
		Delims("<<", ">>").
		Parse(string(b)))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if m == nil {
		_, err = io.Copy(f, &buf)
		return err
	}
	return m.Minify("js", f, &buf)
}

func pwset(src []string) (r [][]string) {
	s := make([]int, len(src))
	for i := 0; i < len(src); i++ {
		s[i] = i
	}
	for _, v := range powerset.Iterative(s) {
		o := make([]string, len(v))
		for i, j := range v {
			o[i] = src[j]
		}
		sort.Strings(o)
		r = append(r, o)
	}
	return
}

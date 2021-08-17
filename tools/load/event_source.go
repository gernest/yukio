package main

import (
	"bytes"
	"context"
	"net/http"
	"sync"
)

type EventPayload struct {
	Name     string            `json:"n"`
	Domain   string            `json:"d"`
	URL      string            `json:"url"`
	HashMode bool              `json:"h"`
	Referrer string            `json:"r"`
	Meta     map[string]string `json:"m"`
}

type Base struct{}

func (Base) Next() *EventPayload {
	return &EventPayload{
		Name:     "pageview",
		Domain:   "yukio.co.tz",
		URL:      "http://yukio.co.tz",
		HashMode: false,
	}
}

type Source interface {
	Next() *EventPayload
}

func NewHosts(base Source, host ...string) Source {
	var ls []interface{}
	for _, v := range host {
		ls = append(ls, v)
	}
	return &ListSource{
		Items: ls,
		Modify: func(e *EventPayload, v interface{}) {
			e.Domain = v.(string)
		},
	}
}

func NewPaths(base Source, paths ...string) Source {
	var ls []interface{}
	for _, v := range paths {
		ls = append(ls, v)
	}
	return &ListSource{
		Items: ls,
		Modify: func(e *EventPayload, v interface{}) {
			e.URL += v.(string)
		},
	}
}

type ListSource struct {
	Items  []interface{}
	Modify func(e *EventPayload, v interface{})
	idx    int
	Source Source
}

func (p *ListSource) Next() *EventPayload {
	if p.idx < len(p.Items) {
		p.idx++
	} else {
		p.idx = 0
	}
	v := p.Items[p.idx]
	e := p.Source.Next()
	p.Modify(e, v)
	return e
}

type Call interface {
	Do(ctx context.Context, e *EventPayload) (*http.Response, error)
}

type Endpoint string

var client = &http.Client{}

var buf = &sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func (e Endpoint) Do(ctx context.Context, body *EventPayload) (*http.Response, error) {
	b := buf.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		buf.Put(b)
	}()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, string(e), b)
	if err != nil {
		return nil, err
	}
	return client.Do(r)
}

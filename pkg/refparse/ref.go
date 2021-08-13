package refparse

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"sync"
)

//go:embed referrer.json
var refererJSON string

var ErrNotFound = errors.New("refparse: Referrer not found")

var index = &sync.Map{}

func init() {
	m := make(map[string]interface{})
	json.Unmarshal([]byte(refererJSON), &m)
	for mTyp, mData := range m {
		for refName, refConfig := range mData.(map[string]interface{}) {
			var params []string
			if p, ok := refConfig.(map[string]interface{})["parameters"]; ok {
				for _, v := range p.([]interface{}) {
					params = append(params, v.(string))
				}
			}
			for _, domain := range refConfig.(map[string]interface{})["domains"].([]interface{}) {
				index.Store(domain.(string), Medium{
					Type:       mTyp,
					Name:       refName,
					Parameters: params,
				})
			}
		}
	}
}

type Medium struct {
	Type       string
	Name       string
	Parameters []string
}

type Referrer struct {
	Medium   string
	Referrer string
	URI      *url.URL
	Search   map[string]string
}

func Parse(u string) (*Referrer, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitAfterN(uri.Host, ".", 2)
	rhost := ""
	if len(parts) > 1 {
		rhost = parts[1]
	}
	queries := []string{uri.Host + uri.Path, rhost + uri.Path, uri.Host, rhost}
	for _, query := range queries {
		m, ok := look(query)
		if ok {
			ref := &Referrer{
				Medium:   m.Type,
				Referrer: m.Name,
				URI:      uri,
			}
			if len(m.Parameters) > 0 {
				ref.Search = make(map[string]string)
				q := uri.Query()

				for _, param := range m.Parameters {
					if v := q.Get(param); v != "" {
						ref.Search[param] = v
					}
				}
			}
			return ref, nil
		}
	}
	return nil, ErrNotFound
}

func exact(m string) (Medium, bool) {
	if v, ok := index.Load(m); ok {
		return v.(Medium), true
	}
	return Medium{}, false
}

func prefix(m string) (o Medium, ok bool) {
	index.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(m, key.(string)) {
			o = value.(Medium)
			ok = true
			return false
		}
		return true
	})
	return
}

func suffix(m string) (o Medium, ok bool) {
	index.Range(func(key, value interface{}) bool {
		if strings.HasSuffix(m, key.(string)) {
			o = value.(Medium)
			ok = true
			return false
		}
		return true
	})
	return
}

func try(r string, fns ...func(string) (Medium, bool)) (m Medium, ok bool) {
	for _, fn := range fns {
		m, ok = fn(r)
		if ok {
			return
		}
	}
	return
}

func look(r string) (Medium, bool) {
	return try(r, exact, suffix, prefix)
}

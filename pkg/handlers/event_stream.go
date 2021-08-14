package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gernest/yukio/pkg/config"
	"github.com/gernest/yukio/pkg/db"
	"github.com/gernest/yukio/pkg/events"
	"github.com/gernest/yukio/pkg/models"
	"github.com/gernest/yukio/pkg/refparse"
	"github.com/gorilla/mux"
	ua "github.com/mileusna/useragent"
)

func Events(w http.ResponseWriter, r *http.Request) {
	ux := ua.Parse(r.UserAgent())
	if ux.Bot {
		return
	}
	var p models.EventPayload
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		return
	}
	uri, err := url.Parse(p.URL)
	if err != nil {
		return
	}
	query := uri.Query()
	ref := parseRefer(uri, p.Referrer)
	path := uri.Path
	if p.HashMode && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	refSource := query.Get("utm_source")
	if refSource == "" {
		if ref.URI != nil {
			cp := *ref.URI
			cp.Host = cleanHost(cp.Host)
			refSource = cp.String()
		}
	}
	referrer := ""
	if ref.URI != nil {
		cp := *ref.URI
		cp.Host = cleanHost(cp.Host)
		cp.RawPath = strings.TrimSuffix(cp.RawPath, "/")
		referrer = cp.String()
	}
	e := &models.Event{
		TS:             time.Now(),
		Name:           p.Name,
		Hostname:       cleanHost(uri.Host),
		Pathname:       path,
		ReferrerSource: refSource,
		Referrer:       referrer,
		UTMMedium:      query.Get("utm_medium"),
		UTMSource:      query.Get("utm_source"),
		UTMCampaign:    query.Get("utm_campaign"),
		Meta:           p.Meta,
	}
	ctx := r.Context()
	settings := config.Get(ctx)
	s, err := db.SaveSession(ctx, e, settings.SessionWindow)
	if err != nil {
		return
	}
	events.Record(ctx, e)
	events.RecordSession(s)
}

func parseRefer(uri *url.URL, r string) refparse.Referrer {
	if r == "" {
		return refparse.Referrer{}
	}
	refUri, err := url.Parse(r)
	if err != nil {
		return refparse.Referrer{}
	}
	a := strings.HasPrefix(refUri.Host, "www.")
	b := strings.HasPrefix(uri.Host, "www.")
	if a != b && refUri.Host != "localhost" {
		rf, err := refparse.Parse(r)
		if err != nil {
			return refparse.Referrer{}
		}
		return *rf
	}
	return refparse.Referrer{}
}

func cleanHost(h string) string {
	return strings.TrimPrefix(h, "www.")
}

func AddRoutes(m *mux.Router) {
	m.HandleFunc("/api/events", Events)
}

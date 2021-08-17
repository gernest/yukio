package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gernest/yukio/pkg/config"
	"github.com/gernest/yukio/pkg/db"
	"github.com/gernest/yukio/pkg/events"
	"github.com/gernest/yukio/pkg/models"
	"github.com/gernest/yukio/pkg/refparse"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	ua "github.com/mileusna/useragent"
	"go.uber.org/zap"
)

func Events(log *zap.Logger) http.HandlerFunc {
	eventSLog := log.Named("events")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ux := ua.Parse(r.UserAgent())
		if ux.Bot {
			return
		}
		var p models.EventPayload
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			return
		}
		// we only accept events for registered domains.
		usr, err := db.GetUserByDomain(ctx, p.Domain)
		if err != nil {
			if db.IsNotFound(err) {
				eventSLog.Error("Failed getting user of the domain",
					zap.String("domain", p.Domain),
					zap.Error(err),
				)
			}
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
			Timestamp:      ptypes.TimestampNow(),
			Name:           p.Name,
			UserId:         usr.Id,
			Hostname:       cleanHost(uri.Host),
			Pathname:       path,
			ReferrerSource: refSource,
			Referrer:       referrer,
			UtmMedium:      query.Get("utm_medium"),
			UtmSource:      query.Get("utm_source"),
			UtmCampaign:    query.Get("utm_campaign"),
			Meta:           p.Meta,
		}
		settings := config.Get(ctx)
		s, err := db.SaveSession(ctx, e, settings.SessionWindow)
		if err != nil {
			eventSLog.Error("Failed to save session",
				zap.String("domain", e.Domain),
				zap.String("event_name", e.Name),
			)
			return
		}
		events.Record(ctx, e)
		events.RecordSession(s)
	})
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

func AddRoutes(m *mux.Router, log *zap.Logger) {
	m.HandleFunc("/api/events", Events(log))
}

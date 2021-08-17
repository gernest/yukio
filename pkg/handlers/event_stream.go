package handlers

import (
	"encoding/json"
	"net"
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

		uri, err := url.Parse(p.URL)
		if err != nil {
			return
		}
		query := uri.Query()
		ref := ParseRefer(uri, p.Referrer)
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
		for _, domain := range GetDomains(r.URL, p.Domain) {
			userID, err := db.GenerateUserID(ctx,
				GetRemoteIP(r), r.UserAgent(), domain, e.Hostname,
			)
			if err != nil {
				eventSLog.Error("Failed generate id for the domain",
					zap.String("domain", p.Domain),
					zap.Error(err),
				)
				return
			}
			e.UserId = userID
			e.Domain = domain
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
		}
	})
}

func ParseRefer(uri *url.URL, r string) refparse.Referrer {
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

func GetRemoteIP(r *http.Request) string {
	var raw string
	switch {
	case r.Header.Get("X-Real-IP") != "":
		raw = r.Header.Get("X-Real-IP")
	case r.Header.Get("X-Forwarded-For") != "":
		raw = r.Header.Get("X-Forwarded-For")
	case r.Header.Get("X-Client-IP") != "":
		raw = r.Header.Get("X-Client-IP")
	case r.RemoteAddr != "":
		raw = r.RemoteAddr
	}
	var host string
	host, _, err := net.SplitHostPort(raw)
	if err != nil {
		host = raw
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "-"
	}
	return ip.String()
}

func GetDomains(r *url.URL, domains string) []string {
	if domains == "" {
		return []string{
			cleanHost(r.Host),
		}
	}
	parts := strings.Split(domains, ",")
	for i := 0; i < len(parts); i++ {
		parts[i] = cleanHost(parts[i])
	}
	return parts
}

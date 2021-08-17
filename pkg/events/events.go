package events

import (
	"context"
	"strconv"

	"github.com/gernest/yukio/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Domain          = "domain"
	Referer         = "referer"
	Path            = "path"
	CustomEventName = "custom_event_name"
	EntryPage       = "entry_page"
	ExitPage        = "exit_page"
	UserID          = "user_id"
)

var PageView = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "page_view",
		Help: "Counts total page views",
	},
	[]string{
		Domain,
		UserID,
		Referer,
		Path,
	},
)

var Custom = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "custom_event",
		Help: "Counts custom event",
	},
	[]string{
		CustomEventName,
		Domain,
		UserID,
		Referer,
		Path,
	},
)

func init() {
	register(PageView, Custom)
}

func Record(ctx context.Context, e *models.Event) {
	if e.Name == "pageview" {
		PageView.WithLabelValues(
			e.Domain,
			strconv.FormatUint(e.UserId, 64),
			e.Referrer,
			e.Pathname,
		).Inc()
	} else {
		Custom.WithLabelValues(
			e.Name,
			e.Domain,
			strconv.FormatUint(e.UserId, 64),
			e.Referrer,
			e.Pathname,
		).Inc()
	}
}

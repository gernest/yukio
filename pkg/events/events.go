package events

import (
	"context"

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
)

var PageView = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "page_view",
		Help: "Counts total page views",
	},
	[]string{
		Domain,
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
		Domain,
		Referer,
		Path,
		CustomEventName,
	},
)

func init() {
	register(PageView, Custom)
}

func Record(ctx context.Context, e *models.Event) {
	if e.Name == "pageview" {
		PageView.WithLabelValues(e.Domain, e.Referrer, e.Pathname).Inc()
	} else {
		Custom.WithLabelValues(e.Name, e.Domain, e.Referrer, e.Pathname).Inc()
	}
}

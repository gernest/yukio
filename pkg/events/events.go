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
	SessionID       = "session_id"
)

type metricname string

const (
	pageView    = metricname("page_view")
	customEvent = metricname("custom_event")
	visitors    = metricname("visitors")
)

var PageView = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: string(pageView),
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
		Name: string(customEvent),
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
			formatNumber(e.UserId),
			e.Referrer,
			e.Pathname,
		).Inc()
	} else {
		Custom.WithLabelValues(
			e.Name,
			e.Domain,
			formatNumber(e.UserId),
			e.Referrer,
			e.Pathname,
		).Inc()
	}
}

func formatNumber(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func eventsSeries(m ...string) (r []metricname) {
	for _, v := range m {
		switch metricname(v) {
		case pageView, visits, bounceRate:
			r = append(r, metricname(v))
		}
	}
	return
}

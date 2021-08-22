package events

import (
	"context"
	"regexp"
	"strconv"

	"github.com/gernest/yukio/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Domain    = "domain"
	Referer   = "referer"
	Path      = "path"
	EventName = "name"
	EntryPage = "entry_page"
	ExitPage  = "exit_page"
	UserID    = "user_id"
	SessionID = "session_id"
)

type metricName string

const (
	customEvent = metricName("events")
	visitors    = metricName("visitors")
)

var Visitors = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: string(visitors),
		Help: "Tracks site visitors",
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
		Help: "Counts events",
	},
	[]string{
		EventName,
		Domain,
		UserID,
		Referer,
		Path,
	},
)

func init() {
	register(Visitors, Custom)
}

func Record(ctx context.Context, e *models.Event) {
	Custom.WithLabelValues(
		e.Name,
		e.Domain,
		formatNumber(e.UserId),
		e.Referrer,
		e.Pathname,
	).Inc()
	Visitors.WithLabelValues(
		e.Name,
		e.Domain,
		formatNumber(e.UserId),
		e.Referrer,
		e.Pathname,
	).Set(1)
}

func formatNumber(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// event filters

func (q *Query) Domain(domain string) *Query {
	return q.Equal(Domain, domain)
}

func (q *Query) IsPage(page string) *Query {
	return q.Equal(Path, page)
}

func (q *Query) IsNotPage(page string) *Query {
	return q.NotEqual(Path, page)
}

func (q *Query) MatchPage(page string) *Query {
	return q.Re(Path, regexp.QuoteMeta(page))
}

func (q *Query) NotMatchPage(page string) *Query {
	return q.NotRe(Path, regexp.QuoteMeta(page))
}

func (q *Query) PageInList(pages ...string) *Query {
	for _, page := range pages {
		q.MatchPage(page)
	}
	return q
}

func (q *Query) PageNotInList(pages ...string) *Query {
	for _, page := range pages {
		q.NotMatchPage(page)
	}
	return q
}

func (q *Query) MatchName(name string) *Query {
	return q.Re(EventName, regexp.QuoteMeta(name))
}

func (q *Query) NameInList(names ...string) *Query {
	for _, name := range names {
		q.MatchName(name)
	}
	return q
}

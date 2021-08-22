package events

import (
	"time"

	"github.com/gernest/yukio/pkg/models"
)

func PageViewQuery(site *models.Site, period Period, nowFunc func() time.Time) QueryRangeRequest {
	return QueryFrom(period, nowFunc).
		Name(string(customEvent)).
		Domain(site.Domain).Request()
}

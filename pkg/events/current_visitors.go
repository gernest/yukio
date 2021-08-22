package events

import (
	"time"

	"github.com/gernest/yukio/pkg/models"
)

func CurrentVisitorsQuery(site *models.Site, nowFunc func() time.Time) QueryRangeRequest {
	return QueryFrom(Current, nowFunc).
		Name(string(visitors)).
		Domain(site.Domain).Request(func(s string) string {
		return GroupBy(s, UserID)
	})
}

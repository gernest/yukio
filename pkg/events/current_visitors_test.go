package events

import (
	"testing"

	"github.com/gernest/yukio/pkg/models"
)

func TestCurrentVisitor(t *testing.T) {
	nowFunc := testNowFunc(t)
	req := CurrentVisitorsQuery(&models.Site{Domain: Yukio}, nowFunc)
	expect := `group by (user_id)(visitors{domain="yukio.co.tz"}[5m0s])`
	got := req.Query
	if got != expect {
		t.Errorf("expected %q got %q", expect, got)
	}
}

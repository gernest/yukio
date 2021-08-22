package stats

import (
	"encoding/json"
	"net/http"

	"github.com/gernest/yukio/pkg/events"
	"github.com/gernest/yukio/pkg/loga"
	"github.com/gernest/yukio/pkg/models"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func CurrentVisitors(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	site := &models.Site{
		Domain: params["domain"],
	}
	req := events.CurrentVisitorsQuery(site, events.Now)
	ctx := r.Context()
	lg := loga.Get(ctx)
	res, err := events.RangeQuery(r.Context(), &req)
	if err != nil {
		lg.Error("Failed to execute query",
			zap.String("query", req.Query),
			zap.Error(err),
		)
		E505(w, err)
		return
	}
	Ok(w, res)
}

func Respond(w http.ResponseWriter, body []byte, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body)
}

func Ok(w http.ResponseWriter, body interface{}) {
	JSON(w, body, http.StatusOK)
}

func JSON(w http.ResponseWriter, body interface{}, code int) {
	b, _ := json.Marshal(body)
	Respond(w, b, code)
}

type Error struct {
	Message string
}

func E505(w http.ResponseWriter, err error) {
	JSON(w, Error{Message: err.Error()}, http.StatusInternalServerError)
}

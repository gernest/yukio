package stats

import "github.com/gorilla/mux"

func AddRoutes(m *mux.Router) {
	m.HandleFunc("/api/stats/{domain}/current", CurrentVisitors)
}

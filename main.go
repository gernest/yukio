package main

import (
	"net/http"

	"github.com/gernest/yukio/pkg/web"
	"github.com/gorilla/mux"
)

//go:generate go run tools/tracker/main.go
//go:generate protoc  --go_out=. --go_opt=paths=source_relative   pkg/models/models.proto
func main() {
	m := mux.NewRouter()
	web.AddRoute(m)
	http.ListenAndServe(":8090", m)
}

package main

//go:generate go run tools/tracker/main.go
//go:generate protoc  --go_out=. --go_opt=paths=source_relative   pkg/models/models.proto
func main() {
}

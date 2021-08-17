package models

type EventPayload struct {
	Name     string            `json:"n"`
	Domain   string            `json:"d"`
	URL      string            `json:"url"`
	HashMode bool              `json:"h"`
	Referrer string            `json:"r"`
	Meta     map[string]string `json:"m"`
}

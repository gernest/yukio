package models

import (
	"time"

	"github.com/google/uuid"
)

type EventPayload struct {
	Name     string            `json:"n"`
	Domain   string            `json:"d"`
	URL      string            `json:"url"`
	HashMode bool              `json:"h"`
	Referrer string            `json:"r"`
	Meta     map[string]string `json:"m"`
}

type Event struct {
	Name                   string
	Domain                 string
	URL                    string
	Hostname               string
	Pathname               string
	Meta                   map[string]string
	UserID                 uuid.UUID
	Referrer               string
	ReferrerSource         string
	UTMMedium              string
	UTMSource              string
	UTMCampaign            string
	CountryCode            string
	ScreenSize             string
	OperatingSystem        string
	OperatingSystemVersion string
	Browser                string
	BrowserVersion         string
	TS                     time.Time
}

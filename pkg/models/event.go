package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID                     int64
	Name                   string
	Domain                 string
	Hostname               string
	Pathname               string
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

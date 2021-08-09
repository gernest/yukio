package models

import "time"

type Event struct {
	ID                     int64
	Name                   string
	Domain                 string
	Hostname               string
	Pathname               string
	UserID                 int64
	SessionID              int64
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

package models

import "time"

type Session struct {
	ID                     int64
	Sign                   int32
	Domain                 string
	Hostname               string
	IsBounce               bool
	EntryPage              string
	ExitPage               string
	PageViews              int32
	Events                 int32
	Duration               time.Duration
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
	Start                  time.Time
	TS                     time.Time
}

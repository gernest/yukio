package models

import "time"

type Session struct {
	ID              int64
	Sign            int32
	Domain          string
	UserID          int64
	Hostname        string
	IsBounce        bool
	EntryPage       string
	ExitPage        string
	PageViews       int32
	Events          int32
	Duration        time.Duration
	Referrer        string
	ReferrerSource  string
	CountryCode     string
	ScreenSize      string
	OperatingSystem string
	Browser         string
	Start           time.Time
	TS              time.Time
}

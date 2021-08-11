package db

import (
	"context"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/models"
)

func resetSession(s *models.Session, e *models.Event) *models.Session {
	var pageview int32
	if e.Name == "pageview" {
		pageview = 1
	}
	s.Sign = 1
	s.Hostname = e.Hostname
	s.Domain = e.Domain
	s.EntryPage = e.Pathname
	s.ExitPage = e.Pathname
	s.IsBounce = true
	s.PageViews = pageview
	s.Events = 1
	s.Referrer = e.Referrer
	s.ReferrerSource = e.ReferrerSource
	s.UTMMedium = e.UTMMedium
	s.UTMSource = e.UTMSource
	s.UTMCampaign = e.UTMCampaign
	s.CountryCode = e.CountryCode
	s.ScreenSize = e.ScreenSize
	s.OperatingSystem = e.OperatingSystem
	s.OperatingSystemVersion = e.OperatingSystemVersion
	s.Browser = e.Browser
	s.BrowserVersion = e.BrowserVersion
	s.TS = e.TS
	s.Start = e.TS
	return s
}

func marshalSession(s *models.Session) ([]byte, error) {
	return nil, nil
}

func unmarshalSession(b []byte, v interface{}) error {
	return nil
}

func CreateSessionFromEvent(ctx context.Context, s *models.Session, event *models.Event) error {
	resetSession(s, event)
	k := gk().SessionID(event.UserID, event.Domain)
	defer pk(k)
	v, err := marshalSession(s)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(k.Bytes(), v)
	})
}

func HandleSession(ctx context.Context, event *models.Event, sessionWindow time.Duration) (err error) {
	var os models.Session
	k := gk().SessionID(event.UserID, event.Domain)
	defer pk(k)
	err = db.View(func(txn *badger.Txn) error {
		x, err := txn.Get(k.Bytes())
		if err != nil {
			return err
		}
		return x.Value(func(val []byte) error {
			return unmarshalSession(val, &os)
		})
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// This is  new event with no session
			return CreateSessionFromEvent(ctx, &os, event)
		}
		return
	}
	active := event.TS.Sub(os.TS) < sessionWindow
	if active {
		var pageView int32
		if event.Name == "pageview" {
			pageView++
		}
		os.TS = event.TS
		os.ExitPage = event.Pathname
		os.Duration = event.TS.Sub(os.Start)
		os.Events++
		os.PageViews++
		return db.Update(func(txn *badger.Txn) error {
			v, err := marshalSession(&os)
			if err != nil {
				return err
			}
			return txn.Set(k.Bytes(), v)
		})
	}
	return CreateSessionFromEvent(ctx, &os, event)
}

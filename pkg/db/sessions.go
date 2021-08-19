package db

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/yukio/pkg/models"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/proto"
)

var sequences = &sync.Map{}

const LeaseSize = uint64(1000)

var sessionIndex *badger.Sequence

func GetSessionID(ctx context.Context) (uint64, error) {
	if sessionIndex != nil {
		return sessionIndex.Next()
	}
	var err error
	sessionIndex, err = GetStore(ctx).GetSequence(SessionLease, LeaseSize)
	if err != nil {
		return 0, err
	}
	return sessionIndex.Next()
}

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
	s.UtmMedium = e.UtmMedium
	s.UtmSource = e.UtmSource
	s.UtmCampaign = e.UtmCampaign
	s.CountryCode = e.CountryCode
	s.ScreenSize = e.ScreenSize
	s.OperatingSystem = e.OperatingSystem
	s.OperatingSystemVersion = e.OperatingSystemVersion
	s.Browser = e.Browser
	s.BrowserVersion = e.BrowserVersion
	s.Timestamp = e.Timestamp
	s.Start = e.Timestamp
	return s
}

func CreateSessionFromEvent(ctx context.Context, s *models.Session, event *models.Event) error {
	id, err := GetSessionID(ctx)
	if err != nil {
		return err
	}
	s.Id = id
	event.SessionId = id
	resetSession(s, event)
	k := gk().SessionID(event.UserId, event.Domain)
	defer pk(k)
	v, err := proto.Marshal(s)
	if err != nil {
		return err
	}
	return GetStore(ctx).Update(func(txn *badger.Txn) error {
		return txn.Set(k.Bytes(), v)
	})
}

func SaveSession(ctx context.Context, event *models.Event, sessionWindow time.Duration) (ss *models.Session, err error) {
	var os models.Session
	defer func() {
		if err == nil {
			ss = &os
		}
	}()
	k := gk().SessionID(event.UserId, event.Domain)
	defer pk(k)
	err = GetStore(ctx).View(func(txn *badger.Txn) error {
		x, err := txn.Get(k.Bytes())
		if err != nil {
			return err
		}
		return x.Value(func(val []byte) error {
			return proto.Unmarshal(val, &os)
		})
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// This is  new event with no session
			err = CreateSessionFromEvent(ctx, &os, event)
		}
		return
	}
	osts, _ := ptypes.Timestamp(os.Timestamp)
	evts, _ := ptypes.Timestamp(event.Timestamp)
	active := evts.Sub(osts) < sessionWindow
	if active {
		var pageView int32
		if event.Name == "pageview" {
			pageView++
		}
		os.Timestamp = event.Timestamp
		os.ExitPage = event.Pathname
		startTS, _ := ptypes.Timestamp(os.Start)
		os.Duration = ptypes.DurationProto(evts.Sub(startTS))
		os.Events++
		os.PageViews++
		err = GetStore(ctx).Update(func(txn *badger.Txn) error {
			v, err := proto.Marshal(&os)
			if err != nil {
				return err
			}
			return txn.Set(k.Bytes(), v)
		})
		return
	}
	err = CreateSessionFromEvent(ctx, &os, event)
	return
}

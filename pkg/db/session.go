package db

import (
	"context"
	"errors"
	"time"

	"github.com/gernest/yukio/pkg/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func newSessionFromEvent(e *models.Event) *models.Session {
	var pageview int32
	if e.Name == "pageview" {
		pageview = 1
	}
	return &models.Session{
		Sign:                   1,
		Hostname:               e.Hostname,
		Domain:                 e.Domain,
		UserID:                 e.UserID,
		EntryPage:              e.Pathname,
		ExitPage:               e.Pathname,
		IsBounce:               true,
		PageViews:              pageview,
		Events:                 1,
		Referrer:               e.Referrer,
		ReferrerSource:         e.ReferrerSource,
		UTMMedium:              e.UTMMedium,
		UTMSource:              e.UTMSource,
		UTMCampaign:            e.UTMCampaign,
		CountryCode:            e.CountryCode,
		ScreenSize:             e.ScreenSize,
		OperatingSystem:        e.OperatingSystem,
		OperatingSystemVersion: e.OperatingSystemVersion,
		Browser:                e.Browser,
		BrowserVersion:         e.BrowserVersion,
		TS:                     e.TS,
		Start:                  e.TS,
	}
}

func CreateSessionFromEvent(ctx context.Context, event *models.Event) (id int64, err error) {
	s := newSessionFromEvent(event)
	err = Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		row := conn.QueryRow(ctx, `insert into sessions (
			sign ,
			domain,
			user_id ,
			hostname,
			is_bounce,
			entry_page,
			exit_page,
			pageviews ,
			events ,
			duration ,
			referrer,
			referrer_source,
			utm_medium,
			utm_source,
			utm_campaign,
			country_code,
			screen_size,
			operating_system,
			operating_system_version,
			browser,
			start ,
			browser_version,
			ts) values(
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23) RETURNING id;`,
			&s.Sign,
			&s.Domain,
			&s.UserID,
			&s.Hostname,
			&s.IsBounce,
			&s.EntryPage,
			&s.ExitPage,
			&s.PageViews,
			&s.Events,
			&s.Duration,
			&s.Referrer,
			&s.ReferrerSource,
			&s.UTMMedium,
			&s.UTMSource,
			&s.UTMCampaign,
			&s.CountryCode,
			&s.ScreenSize,
			&s.OperatingSystem,
			&s.OperatingSystemVersion,
			&s.Browser,
			&s.Start,
			&s.BrowserVersion,
			&s.TS,
		)
		return row.Scan(&id)
	})
	return
}

func HandleSession(ctx context.Context, event *models.Event, sessionWindow time.Duration) (id int64, err error) {
	var ts, start time.Time
	err = Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		return conn.QueryRow(ctx, `select id,ts,start from sessions where user_id=$1 and domain=$2;`,
			event.UserID, event.Domain).Scan(&id, &ts, &start)
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This is  new event with no session
			return CreateSessionFromEvent(ctx, event)
		}
		return
	}
	active := event.TS.Sub(ts) < sessionWindow
	if active {
		var pageView int32
		if event.Name == "pageview" {
			pageView++
		}
		err = updateSession(ctx,
			id, event.TS, event.Pathname, event.TS.Sub(start), pageView)
		return
	}
	return CreateSessionFromEvent(ctx, event)
}

func updateSession(ctx context.Context,
	id int64,
	ts time.Time,
	exitPage string,
	duration time.Duration,
	pageView int32,
) error {
	return Do(ctx, func(ctx context.Context, conn *pgxpool.Conn) error {
		_, err := conn.Exec(ctx, `update sessions set 
		(ts,exit_page,is_bounce,duration,pageview,events) =
		($1,$2,$3,$4,pageview+$5,events+1)
		where id=$6;`,
			ts, exitPage, false, duration, pageView, id,
		)
		return err
	})
}

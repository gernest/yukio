CREATE TABLE IF NOT EXISTS users(
    id bigserial PRIMARY KEY,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS events(
    id bigserial,
    name TEXT,
    domain TEXT,
    hostname TEXT,
    pathname TEXT,
    user_id bigint,
    session_id bigint,
    referrer TEXT,
    referrer_source TEXT,
    utm_medium TEXT,
    utm_source TEXT,
    utm_campaign TEXT,
    country_code TEXT,
    screen_size TEXT,
    operating_system TEXT,
    operating_system_version TEXT,
    browser TEXT,
    browser_version TEXT,
    ts TIMESTAMP
);
CREATE TABLE IF NOT EXISTS sessions(
    id bigserial,
    sign integer,
    domain TEXT,
    user_id bigint,
    hostname TEXT,
    is_bounce boolean,
    entry_page TEXT,
    exit_page TEXT,
    pageviews integer,
    events integer,
    duration integer,
    referrer TEXT,
    referrer_source TEXT,
    utm_medium TEXT,
    utm_source TEXT,
    utm_campaign TEXT,
    country_code TEXT,
    screen_size TEXT,
    operating_system TEXT,
    operating_system_version TEXT,
    browser TEXT,
    start TIMESTAMP,
    browser_version TEXT,
    ts TIMESTAMP
);
SELECT create_hypertable('events', 'ts', if_not_exists => TRUE);
SELECT create_hypertable('sessions', 'start', if_not_exists => TRUE);
module github.com/rankguessr/api

go 1.25.5

require (
	github.com/go-co-op/gocron/v2 v2.19.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.9.1
	github.com/labstack/echo/v5 v5.1.0
	github.com/segmentio/ksuid v1.0.4
	github.com/urfave/cli/v3 v3.8.0
	github.com/wieku/rplpa v1.0.2
)

require (
	github.com/bnch/uleb128 v0.0.0-20160221084957-fac1fe18ad59 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/getsentry/sentry-go v0.45.1 // indirect
	github.com/getsentry/sentry-go/echo v0.45.1 // indirect
	github.com/getsentry/sentry-go/slog v0.45.1 // indirect
	github.com/itchio/lzma v0.0.0-20190703113020-d3e24e3e3d49 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/redis/go-redis/v9 v9.18.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)

replace github.com/wieku/rplpa => ./rplpa

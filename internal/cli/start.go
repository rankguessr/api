package cli

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/internal/handlers"
	rmiddleware "github.com/rankguessr/api/internal/middleware"
	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/migrate"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v3"
)

func StartCmd(ctx context.Context, c *cli.Command) error {
	isDev := c.Bool("dev")

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	dbCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to parse database url: ", err)
	}

	dbCfg.MaxConns = 10
	dbCfg.MaxConnLifetime = 30 * time.Minute
	dbCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatal("failed to ping db: ", err)
	}

	err = migrate.RunMigrations(ctx, pool)
	if err != nil {
		log.Fatal("failed to run migrations: ", err)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("failed to parse redis conn string")
	}

	rdb := redis.NewClient(opt)

	err = rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatal("failed to ping redis: ", err)
	}

	client := osuapi.NewClient(cfg.OsuClientID, cfg.OsuClientSecret)

	userRepo := repo.NewUsers(pool)
	userService := service.NewUser(userRepo)

	playerRepo := repo.NewPlayers(pool)
	playerService := service.NewPlayer(playerRepo)

	roomsRepo := repo.NewRooms(pool)
	roomsService := service.NewRooms(roomsRepo, client, rdb)

	guessRepo := repo.NewGuesses(pool)
	guessService := service.NewGuess(guessRepo)

	sessionsRepo := repo.NewSessions(pool)
	sessionsService := service.NewSessions(cfg, sessionsRepo)

	submissionsRepo := repo.NewSubmissions(pool)
	submissionsService := service.NewSubmissions(submissionsRepo)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(rmiddleware.RequestLogger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{cfg.WebURL},
		AllowHeaders:     []string{echo.HeaderAccept, echo.HeaderOrigin, echo.HeaderContentType},
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(60.0)))
	e.Use(middleware.ContextTimeout(time.Second * 30))

	if !isDev {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:           cfg.SentryDSN,
			EnableLogs:    true,
			EnableTracing: true,
		}); err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}

		handler := sentryslog.Option{
			EventLevel: []slog.Level{slog.LevelError},
			LogLevel:   []slog.Level{slog.LevelWarn, slog.LevelInfo, slog.LevelDebug},
		}.NewSentryHandler(ctx)

		logger := slog.New(handler)

		defer sentry.Flush(2 * time.Second)

		slog.SetDefault(logger)
		e.Use(sentryecho.New(sentryecho.Options{}))
	}

	sessions := rmiddleware.Session(client, sessionsService)

	{
		e.GET("/health", handlers.HealthCheck)
		e.GET("/stats", handlers.PublicStatsGet(guessService, userService, rdb))
	}

	auth := e.Group("/auth")
	{
		auth.GET("/login", handlers.AuthLogin(cfg))
		auth.GET("/callback", handlers.AuthCallback(cfg, client, userService, sessionsService))
		auth.GET("/logout", handlers.AuthLogout(cfg))
	}

	user := e.Group("/user")
	{
		user.Use(sessions)
		user.GET("/me", handlers.AuthMe(userService))
		user.GET("/rooms", handlers.UserGetRoomsData(roomsService, client, guessService))
	}

	room := e.Group("/room")
	{
		room.Use(sessions)
		room.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20.0)))
		room.GET("/:id/score", handlers.RoomGetScore(roomsService, guessService, client))
		room.GET("/replay/:filename", handlers.RoomDownloadReplay(roomsService, client))

		room.POST("/:id", handlers.RoomSubmitGuess(roomsService, guessService, client))
		room.POST("/:id/next", handlers.RoomGetNext(roomsService, playerService, client))
		room.POST("/start", handlers.RoomStart(playerService, roomsService, client))
	}

	submissions := e.Group("/submission")
	{
		submissions.Use(sessions)
		submissions.POST("/", handlers.SubmissionCreate(submissionsService, client))
		submissions.POST("/:id/accept", handlers.SubmissionSetAccepted(submissionsService))

		submissions.GET("/unaccepted", handlers.SubmissionFindUnaccepted(submissionsService))
		submissions.DELETE("/:id", handlers.SubmissionDelete(submissionsService))
	}

	return e.Start(fmt.Sprintf(":%s", cfg.PORT))
}

package cli

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/internal/handlers"
	rmiddleware "github.com/rankguessr/api/internal/middleware"
	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/migrate"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v3"
)

var lifecycleConfig = &lifecycle.Configuration{
	Rules: []lifecycle.Rule{
		{
			ID: "OldReplays",
			Expiration: lifecycle.Expiration{
				Days: 1,
			},
			RuleFilter: lifecycle.Filter{
				Prefix: "replays/",
			},
			Status: "Enabled",
		},
	},
}

func StartCmd(ctx context.Context, c *cli.Command) error {
	// isDev := c.Bool("dev")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	minioClient, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("s3 connection failed: ", err)
	}

	exists, err := minioClient.BucketExists(ctx, cfg.S3BucketName)
	if err != nil {
		log.Fatal("failed to check if bucket exists: ", err)
	}

	if !exists {
		log.Fatal("bucket does not exist: ", cfg.S3BucketName)
	}

	err = minioClient.SetBucketLifecycle(ctx, cfg.S3BucketName, lifecycleConfig)
	if err != nil {
		slog.Error("failed to set bucket lifecycle: ", slog.String("err", err.Error()))
	}

	rdb := redis.NewClient(opt)

	err = rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatal("redis ping failed: ", err)
	}

	client := osuapi.NewClient(cfg.OsuClientID, cfg.OsuClientSecret)

	uow := uow.New(pool)
	usersRepo := repo.NewUsers(uow)
	usersSvc := service.NewUser(usersRepo)

	playersRepo := repo.NewPlayers(uow)
	playersSvc := service.NewPlayer(playersRepo)

	guessesRepo := repo.NewGuesses(uow)
	guessesSvc := service.NewGuess(guessesRepo, usersRepo, uow)

	roomsRepo := repo.NewRooms(uow)
	roomsSvc := service.NewRooms(roomsRepo, usersRepo, playersRepo, guessesSvc, client, rdb, uow)

	sessionsRepo := repo.NewSessions(uow)
	sessionsSvc := service.NewSessions(cfg, sessionsRepo)

	submissionsRepo := repo.NewSubmissions(uow)
	submissionsSvc := service.NewSubmissions(submissionsRepo, client)

	replaysSvc := service.NewReplays(cfg, minioClient)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{cfg.WebURL},
		AllowHeaders:     []string{echo.HeaderAccept, echo.HeaderOrigin, echo.HeaderContentType},
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(60.0)))
	e.Use(middleware.ContextTimeout(time.Second * 30))

	e.Use(rmiddleware.RequestLogger())

	sessions := rmiddleware.Session(client, sessionsSvc)

	e.Use(echoprometheus.NewMiddleware("rankguessr"))

	{
		e.GET("/metrics", echoprometheus.NewHandler())
		e.GET("/health", handlers.HealthCheck)
	}

	stats := e.Group("/stats")
	{
		stats.GET("", handlers.PublicGetStats(guessesSvc, usersSvc, rdb))
		stats.GET("/top-users", handlers.PublicGetTopUsers(usersSvc, rdb))
	}

	auth := e.Group("/auth")
	{
		auth.GET("/login", handlers.AuthLogin(cfg))
		auth.GET("/callback", handlers.AuthCallback(cfg, client, usersSvc, sessionsSvc))
		auth.GET("/logout", handlers.AuthLogout(cfg))
	}

	user := e.Group("/user")
	{
		user.Use(sessions)
		user.GET("/me", handlers.AuthMe(usersSvc, roomsSvc))
		user.GET("/guesses", handlers.UserGetGuesses(guessesSvc))
		user.GET("/room", handlers.UserGetCurrentRoom(roomsSvc))
	}

	room := e.Group("/room")
	{
		room.Use(sessions)
		room.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20.0)))
		room.GET("/:id/score", handlers.RoomGetScore(roomsSvc, guessesSvc, submissionsSvc))
		room.GET("/replay/:filename", handlers.RoomDownloadReplay(roomsSvc, replaysSvc, submissionsSvc, client))

		room.POST("/:id/prepare", handlers.RoomPrepareReplay(roomsSvc, client, replaysSvc, submissionsSvc))
		room.POST("/:id", handlers.RoomSubmitGuess(roomsSvc, guessesSvc, client, cfg))
		room.POST("/:id/next", handlers.RoomGetNext(roomsSvc, playersSvc, submissionsSvc))
		room.POST("/start", handlers.RoomStart(playersSvc, roomsSvc, submissionsSvc))
	}

	submissions := e.Group("/submissions")
	{
		submissions.Use(sessions)
		submissions.GET("", handlers.SubmissionsFind(submissionsSvc))
		submissions.POST("", handlers.SubmissionCreate(submissionsSvc, client, replaysSvc))
		submissions.POST("/:id/accept", handlers.SubmissionSetAccepted(submissionsSvc))

		submissions.DELETE("/:id", handlers.SubmissionDelete(submissionsSvc))
	}

	return e.Start(fmt.Sprintf(":%s", cfg.PORT))
}

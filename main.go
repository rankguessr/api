package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/internal/handlers"
	"github.com/rankguessr/api/internal/jobs"
	rmiddleware "github.com/rankguessr/api/internal/middleware"
	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/migrate"
	"github.com/rankguessr/api/pkg/osuapi"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func init() {
	slog.SetDefault(logger)
}

func main() {
	ctx := context.Background()

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	client := osuapi.NewClient(cfg.OsuClientID, cfg.OsuClientSecret, cfg.AppURL)

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
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

	userRepo := repo.NewUsers(pool)
	userService := service.NewUser(userRepo)

	playerRepo := repo.NewPlayers(pool)
	playerService := service.NewPlayer(playerRepo)

	roomsRepo := repo.NewRooms(pool)
	roomsService := service.NewRooms(roomsRepo)

	guessRepo := repo.NewGuesses(pool)
	guessService := service.NewGuess(guessRepo)

	sessionsRepo := repo.NewSessions(pool)
	sessionsService := service.NewSessions(sessionsRepo)

	sch, err := jobs.NewScheduler(client, playerService)
	if err != nil {
		log.Fatal("failed to create scheduler: ", err)
	}

	err = sch.RegisterJobs()
	if err != nil {
		log.Fatal("failed to register jobs: ", err)
	}

	go sch.Start()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{cfg.WebURL},
		AllowHeaders:     []string{echo.HeaderAccept, echo.HeaderOrigin, echo.HeaderContentType},
	}))
	e.Use(rmiddleware.RequestLogger(logger))
	e.Use(middleware.ContextTimeout(time.Second * 30))
	sessions := rmiddleware.Session(client, sessionsService)

	{
		e.GET("/health", handlers.HealthCheck)
		e.GET("/stats", handlers.PublicStatsGet(guessService))
	}

	auth := e.Group("/auth")
	{
		auth.GET("/login", handlers.AuthLogin(cfg))
		auth.GET("/callback", handlers.AuthCallback(cfg, client, userService, sessionsService))
	}

	user := e.Group("/user")
	{
		user.Use(sessions)
		user.GET("/me", handlers.AuthMe(userService))
		user.GET("/latest", handlers.UserGetLatest(userService, guessService))
	}

	room := e.Group("/room")
	{
		room.Use(sessions)
		room.GET("/:id/score", handlers.RoomGetScore(roomsService, client))
		room.GET("/:id/replay", handlers.RoomDownloadReplay(roomsService, client))
		room.POST("/:id", handlers.RoomSubmitGuess(roomsService, guessService, client))
		room.POST("/start", handlers.RoomStart(playerService, roomsService, client))
	}

	log.Fatal(e.Start(":8080"))
}

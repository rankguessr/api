package main

import (
	"context"
	"fmt"
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
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/migrate"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/urfave/cli/v3"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func init() {
	slog.SetDefault(logger)
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start rankguessr api",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "port",
						Value: 8080,
						Usage: "api port",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					port := c.Int("port")

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
					sessionsService := service.NewSessions(cfg, sessionsRepo)

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
						e.GET("/stats", handlers.PublicStatsGet(guessService, userService))
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
						user.GET("/current-room", handlers.UserGetCurrentRoom(roomsService, client))
						user.GET("/latest", handlers.UserGetLatest(userService, guessService))
					}

					room := e.Group("/room")
					{
						room.Use(sessions)
						room.GET("/:id/score", handlers.RoomGetScore(roomsService, guessService, client))
						room.GET("/replay/:filename", handlers.RoomDownloadReplay(roomsService, client))

						room.POST("/:id", handlers.RoomSubmitGuess(roomsService, guessService, client))
						room.POST("/:id/next", handlers.RoomGetNext(roomsService, playerService, client))
						room.POST("/start", handlers.RoomStart(playerService, roomsService, client))
					}

					return e.Start(fmt.Sprintf(":%d", port))
				},
			},
			{
				Name:  "collect",
				Usage: "collect osu! players",
				Commands: []*cli.Command{
					{
						Name: "top10k",
						Action: func(ctx context.Context, c *cli.Command) error {
							databaseURL, ok := os.LookupEnv("DATABASE_URL")
							if !ok {
								log.Fatal("DATABASE_URL environment variable not set")
							}

							osuClientId, ok := os.LookupEnv("OSU_CLIENT_ID")
							if !ok {
								log.Fatal("OSU_CLIENT_ID environment variable not set")
							}

							osuClientSecret, ok := os.LookupEnv("OSU_CLIENT_SECRET")
							if !ok {
								log.Fatal("OSU_CLIENT_SECRET environment variable not set")
							}

							pool, err := pgxpool.New(ctx, databaseURL)
							if err != nil {
								log.Fatal("failed to open db connection: ", err)
							}
							defer pool.Close()

							err = pool.Ping(ctx)
							if err != nil {
								log.Fatal("failed to ping db: ", err)
							}

							playerRepo := repo.NewPlayers(pool)
							playerSvc := service.NewPlayer(playerRepo)

							client := osuapi.NewClient(osuClientId, osuClientSecret, "")

							t, err := client.GetClientAccessToken(ctx)
							if err != nil {
								log.Fatal("failed to get client access token: ", err)
							}

							log.Println(t)

							cursor := &osuapi.Cursor{Page: 60}
							for {
								time.Sleep(2 * time.Second)
								rankings, err := client.GetRankings(ctx, t.AccessToken, cursor)
								if err != nil {
									log.Printf("failed to get rankings for page %d: %v", cursor.Page, err)
									break
								}

								if len(rankings.Ranking) == 0 {
									log.Printf("no more players found at page %d, stopping", cursor.Page)
									break
								}

								var players []domain.PlayerCreate
								for _, user := range rankings.Ranking {
									players = append(players, domain.PlayerCreate{
										OsuId:  user.User.ID,
										Source: "top10k",
									})
								}

								inserted, err := playerSvc.CreateMany(ctx, players)
								if err != nil {
									log.Printf("failed to create players for page %d: %v", cursor.Page, err)
									break
								}

								log.Printf("collected page %d with %d players, inserted %d", cursor.Page, len(rankings.Ranking), inserted)

								cursor = rankings.Cursor
							}

							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

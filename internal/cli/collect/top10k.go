package collect

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/urfave/cli/v3"
)

func Top10kCmd(ctx context.Context, c *cli.Command) error {
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

	client := osuapi.NewClient(osuClientId, osuClientSecret)

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
}

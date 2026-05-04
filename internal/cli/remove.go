package cli

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/urfave/cli/v3"
)

func RemoveCmd(ctx context.Context, c *cli.Command) error {
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

	uow := uow.New(pool)
	playerRepo := repo.NewPlayers(uow)
	playerSvc := service.NewPlayer(playerRepo)

	client := osuapi.NewClient(osuClientId, osuClientSecret)

	t, err := client.GetClientAccessToken(ctx)
	if err != nil {
		log.Fatal("failed to get client access token: ", err)
	}

	page := 0
	for {
		players, err := playerSvc.Find(ctx, 49, page)
		if err != nil {
			log.Printf("failed to get players for page %d: %v", page, err)
			break
		}

		if len(players) == 0 {
			log.Printf("no more players found at page %d, stopping", page)
			break
		}

		playerIds := make([]int, len(players))
		for i, p := range players {
			playerIds[i] = p.OsuId
		}

		log.Printf("checking page %d with %d players\n", page, len(players))

		batchResponse, err := client.GetUsers(ctx, t.AccessToken, playerIds)
		if err != nil {
			log.Printf("failed to get osu players for page %d: %v", page, err)
			break
		}

		log.Printf("got osu players for page %d, %d\n", page, len(batchResponse.Users))

		batch := &pgx.Batch{}
		for _, p := range batchResponse.Users {
			if p.StatisticsRulesets.Standard.GlobalRank > 1_000_000 || p.StatisticsRulesets.Standard.GlobalRank == 0 {
				log.Printf("queuing delete for player %d with rank %d", p.ID, p.StatisticsRulesets.Standard.GlobalRank)
				batch.Queue(`DELETE FROM players WHERE osu_id = $1`, p.ID)
			} else {
				log.Printf("keeping player %d with rank %d", p.ID, p.StatisticsRulesets.Standard.GlobalRank)
			}
		}

		res := pool.SendBatch(ctx, batch)
		err = res.Close()
		if err != nil {
			log.Printf("failed to execute batch for page %d: %v", page, err)
			break
		}

		log.Printf("finished checking page %d\n", page)
		page++
		time.Sleep(time.Second * 15)
	}

	return nil
}

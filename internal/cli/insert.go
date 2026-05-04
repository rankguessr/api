package cli

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rankguessr/api/internal/repo"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/internal/uow"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/urfave/cli/v3"
)

func InsertPlayers(ctx context.Context, c *cli.Command) error {
	databaseURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable not set")
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

	dir := c.String("dir")

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		log.Println("Processing file: ", e.Name())
		file, err := os.Open(path.Join(dir, e.Name()))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		r := csv.NewReader(file)
		_, err = r.Read()
		if err != nil {
			log.Fatal("Failed to read csv header: ", err)
		}

		for {
			players, err := readPlayersWithLimit(r, 50)
			if errors.Is(err, io.EOF) {
				log.Println("Finished processing file: ", e.Name())
				break
			} else if err != nil {
				log.Fatal("Failed to read players from csv: ", err)
			}

			created, err := playerSvc.CreateMany(ctx, players)
			if err != nil {
				log.Fatal("Failed to create players: ", err)
			}

			log.Printf("Created %d players", created)
		}
	}

	return nil
}

func readPlayersWithLimit(r *csv.Reader, limit int) ([]domain.PlayerCreate, error) {
	players := make([]domain.PlayerCreate, limit)
	for i := range 50 {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		} else if err != nil {
			log.Fatal("Failed to read csv record: ", err)
		}

		player, err := recordToPlayer(record)
		if err != nil {
			log.Printf("Skipping invalid record: %v, error: %v", record, err)
			continue
		}

		players[i] = player
	}

	return players, nil
}

func recordToPlayer(record []string) (domain.PlayerCreate, error) {
	if len(record) != 2 {
		return domain.PlayerCreate{}, errors.New("invalid record format")
	}

	osuId, err := strconv.Atoi(record[0])
	if err != nil {
		return domain.PlayerCreate{}, errors.New("invalid osuId format")
	}

	return domain.PlayerCreate{
		OsuId:  osuId,
		Source: record[1],
	}, nil
}

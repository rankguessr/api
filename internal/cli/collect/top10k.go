package collect

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/urfave/cli/v3"
)

func Top10kCmd(ctx context.Context, c *cli.Command) error {
	cfg, err := config.ReadOsuAuth()
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}

	client := osuapi.NewClient(cfg.OsuClientID, cfg.OsuClientSecret)

	t, err := client.GetClientAccessToken(ctx)
	if err != nil {
		log.Fatal("failed to get client access token: ", err)
	}

	page := c.Int("page")
	delay := c.Int("delay")
	outFile := c.String("out")
	maxRank := c.Int("max-rank")
	country := c.String("country")
	if outFile == "" {
		outFile = fmt.Sprintf("collected/top10k_%s.csv", strings.ToLower(country))
	}

	dir := filepath.Dir(outFile)
	err = os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		log.Fatal("failed to create output directory: ", err)
	}

	out, err := os.Create(outFile)
	if err != nil {
		log.Fatal("failed to create output file: ", err)
	}
	defer out.Close()

	w := csv.NewWriter(out)

	cursor := &osuapi.Cursor{Page: page}
	opts := osuapi.RankingsOpts{
		Country: &country,
	}

	if country == "global" {
		opts.Country = nil
	}

	if err := w.Write([]string{"osu_id", "source"}); err != nil {
		log.Fatal("failed to write header: ", err)
	}

	defer w.Flush()

	for {
		rankings, err := client.GetRankings(ctx, t.AccessToken, cursor, opts)
		if err != nil {
			log.Printf("failed to get rankings for page %d: %v", cursor.Page, err)
			break
		}

		if len(rankings.Ranking) == 0 {
			log.Printf("no more players found at page %d, stopping", cursor.Page)
			break
		}

		if len(rankings.Ranking) != 0 && rankings.Ranking[0].GlobalRank > maxRank {
			log.Printf("reached max rank %d at page %d, stopping", maxRank, cursor.Page)
			break
		}

		var players [][]string
		for _, user := range rankings.Ranking {
			if user.GlobalRank < maxRank && user.GlobalRank != 0 {
				players = append(players, []string{strconv.Itoa(user.User.ID), "top10k"})
			}
		}

		if err := w.WriteAll(players); err != nil {
			log.Fatal("failed to write players: ", err)
		}

		if cursor == nil || cursor.Page == 0 {
			log.Printf("no more pages to fetch, stopping")
			break
		}

		log.Printf("collected page %d with %d players", cursor.Page, len(players))

		cursor = rankings.Cursor

		time.Sleep(time.Duration(delay) * time.Second)
	}

	return nil
}

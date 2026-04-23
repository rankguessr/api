package cli

import (
	"context"
	"os"

	"github.com/rankguessr/api/internal/cli/collect"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context) error {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name: "start",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dev",
						Value: false,
						Usage: "either to start api in dev mode",
					},
				},
				Usage:  "start rankguessr server",
				Action: StartCmd,
			},
			{
				Name:   "remove",
				Usage:  "remove osu players with rank >1mil",
				Action: RemoveCmd,
			},
			{
				Name:  "insert",
				Usage: "insert players from csv files in a directory",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "directory containing csv files to insert players from",
						Value:   "./collected",
					},
				},
				Action: InsertPlayers,
			},
			{
				Name:  "collect",
				Usage: "collect osu! players",
				Commands: []*cli.Command{
					{
						Name:  "top10k",
						Usage: "collect top 10k players",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "out",
								Aliases: []string{"o"},
								Usage:   "output file for collected players (default is collected/top10k_*country*.csv)",
							},
							&cli.StringFlag{
								Name:  "country",
								Value: "global",
								Usage: "country code to collect top 10k players for",
							},
							&cli.IntFlag{
								Name:  "page",
								Value: 1,
								Usage: "page to start collecting from",
							},
							&cli.IntFlag{
								Name:  "delay",
								Value: 3,
								Usage: "delay in seconds between requests",
							},
							&cli.IntFlag{
								Name:  "max-rank",
								Usage: "max rank to collect",
								Value: 500000,
							},
						},
						Action: collect.Top10kCmd,
					},
				},
			},
		},
	}

	return cmd.Run(ctx, os.Args)
}

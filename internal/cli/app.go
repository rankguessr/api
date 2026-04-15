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
				Name:   "start",
				Usage:  "start rankguessr server",
				Action: StartCmd,
			},
			{
				Name:   "remove",
				Usage:  "remove osu players with rank >1mil",
				Action: RemoveCmd,
			},
			{
				Name:  "collect",
				Usage: "collect osu! players",
				Commands: []*cli.Command{
					{
						Name:   "top10k",
						Action: collect.Top10kCmd,
					},
				},
			},
		},
	}

	return cmd.Run(ctx, os.Args)
}

package jobs

import (
	"context"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
)

type Scheduler struct {
	client *osuapi.Client
	player service.Players
	sch    gocron.Scheduler
}

func NewScheduler(client *osuapi.Client, player service.Players) (*Scheduler, error) {
	sch, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		sch:    sch,
		client: client,
		player: player,
	}, nil
}

func (s *Scheduler) collectFromMultis() (int64, error) {
	ctx := context.Background()
	token, err := s.client.GetClientAccessToken(ctx)
	if err != nil {
		return 0, err
	}

	rooms, err := s.client.GetMultiRooms(ctx, token.AccessToken)
	if err != nil {
		return 0, err
	}

	players := make([]domain.PlayerCreate, 0)
	for _, room := range rooms {
		for _, user := range room.RecentParticipants {
			players = append(players, domain.PlayerCreate{
				OsuId:  user.ID,
				Source: "multi",
			})
		}
	}

	count, err := s.player.CreateMany(context.Background(), players)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Scheduler) RegisterJobs() error {
	j, err := s.sch.NewJob(
		gocron.DurationJob(time.Minute),
		gocron.NewTask(
			func() {
				count, err := s.collectFromMultis()
				if err != nil {
					log.Println("Error collecting from multis:", err)
					return
				}

				log.Printf("Collected %d players from multis\n", count)
			},
		),
	)
	if err != nil {
		return err
	}

	log.Println("Created multis job with id: ", j.ID())

	return nil
}

func (s *Scheduler) Start() {
	s.sch.Start()
}

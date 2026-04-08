package handlers

import (
	"errors"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
	"github.com/wieku/rplpa"
)

func findWithReplay(scores []osuapi.Score) (osuapi.Score, error) {
	if len(scores) == 0 {
		return osuapi.Score{}, errors.New("no replays available in top plays")
	}

	score := scores[0]
	if score.Replay() {
		return score, nil
	}

	return findWithReplay(scores[1:])
}

func RoomStart(player service.Players, rooms service.Rooms, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		p, err := player.FindRandom(ctx)
		if err != nil {
			return err
		}

		scores, err := client.GetUserScores(ctx, session.AccessToken, p.OsuId)
		if err != nil {
			return err
		}

		for i := range scores {
			j := rand.Intn(i + 1)
			scores[i], scores[j] = scores[j], scores[i]
		}

		score, err := findWithReplay(scores)
		if err != nil {
			return err
		}

		s, err := rooms.Create(ctx, p.OsuId, session.User.OsuID, score.ID)
		if err != nil {
			return err
		}

		return c.JSON(200, utils.Map{
			"session_id": s.ID,
		})
	}
}

func RoomDownloadReplay(rooms service.Rooms, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")
		room, err := rooms.FindByID(ctx, roomId)
		if err != nil {
			return err
		}

		replay, err := client.DownloadReplay(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return err
		}

		r, err := rplpa.ParseReplay(replay)
		if err != nil {
			return err
		}

		r.Username = "rankguessr"
		r.ScoreID = 0
		if r.ScoreInfo != nil {
			r.ScoreInfo.ScoreId = 0
		}

		anonymized, err := rplpa.WriteReplay(r)
		if err != nil {
			return err
		}

		return c.Blob(200, "application/x-osu-replay", anonymized)
	}
}

func RoomGetScore(rooms service.Rooms, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")
		room, err := rooms.FindByID(ctx, roomId)
		if err != nil {
			return err
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		score, err := client.GetScore(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return err
		}

		return c.JSON(200, utils.Map{
			"score": utils.Map{
				"pp":         score.PP,
				"mods":       score.Mods,
				"accuracy":   score.Accuracy,
				"beatmapset": score.BeatmapSet,
				"beatmap":    score.BeatmapExtended,
			},
			"is_closed": room.IsClosed,
		})
	}
}

type submitRequest struct {
	Guess int `json:"guess"`
}

func RoomSubmitGuess(rooms service.Rooms, guess service.Guess, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		var req submitRequest
		if err := c.Bind(&req); err != nil {
			return echo.ErrBadRequest.Wrap(err)
		}

		roomId := c.Param("id")
		room, err := rooms.FindByID(ctx, roomId)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		if room.IsClosed {
			return echo.NewHTTPError(http.StatusBadRequest, "room is already closed")
		}

		player, err := client.GetUser(ctx, session.AccessToken, room.PlayerID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		g, err := guess.Create(ctx, session.User.OsuID, player.ID, req.Guess, player.Statistics.GlobalRank)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		err = rooms.Close(ctx, roomId)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(200, g)
	}
}

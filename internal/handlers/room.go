package handlers

import (
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/ranking"
	"github.com/rankguessr/api/pkg/utils"
	"github.com/wieku/rplpa"
)

const RoomStartMaxRetries = 5
const ScoresLimit = 20

// TODO: this MUST be refactored, we need to use transactions
func RoomStart(player service.Players, rooms service.Rooms, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		_, err = rooms.FindByUserUnguessed(ctx, session.User.OsuID)
		if err == nil {
			return c.JSON(http.StatusBadRequest, utils.Map{
				"message": "user is already in room",
			})
		}

		err = rooms.DeleteByUser(ctx, session.User.OsuID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		for range RoomStartMaxRetries {
			tryFind := func() (string, error) {
				p, err := player.FindRandom(ctx)
				if err != nil {
					return "", err
				}

				scoreIdx := rand.Intn(ScoresLimit)

				scores, err := client.GetUserScores(ctx, session.AccessToken, p.OsuId, 1, scoreIdx)
				if err != nil || len(scores) == 0 {
					return "", err
				}

				scoreId := scores[0].ID
				// check if score exists and warm up cache
				score, err := client.GetScore(ctx, session.AccessToken, scoreId)
				if err != nil {
					return "", err
				}

				if score.PP == 0 {
					return "", errors.New("score has 0 pp")
				}

				room, err := rooms.Create(ctx, score.User.ID, session.User.OsuID, score.ID)
				if err != nil {
					return "", err
				}

				return room.ID, nil
			}

			roomId, err := tryFind()
			if err == nil {
				refill, err := rooms.RefillForUser(ctx, session.User.OsuID, 1)
				if errors.Is(err, utils.ErrNotEnoughGuesses) {
					return echo.NewHTTPError(http.StatusBadRequest, "no guesses left")
				}

				if err != nil {
					rooms.DeleteById(ctx, roomId)
					return echo.ErrInternalServerError.Wrap(err)
				}

				return c.JSON(200, utils.Map{
					"room_id": roomId,
					"refill":  refill,
				})
			}
		}

		return c.JSON(http.StatusBadRequest, utils.Map{
			"message": "failed to find a score",
		})
	}
}

func RoomGetNext(rooms service.Rooms, players service.Players, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")

		// TODO: remove copy pasted bullshit
		for range RoomStartMaxRetries {
			tryFind := func() (osuapi.Score, error) {
				p, err := players.FindRandom(ctx)
				if err != nil {
					return osuapi.Score{}, err
				}

				scoreIdx := rand.Intn(ScoresLimit)

				scores, err := client.GetUserScores(ctx, session.AccessToken, p.OsuId, 1, scoreIdx)
				if err != nil || len(scores) == 0 {
					return osuapi.Score{}, err
				}

				scoreId := scores[0].ID
				// check if score exists and warm up cache
				score, err := rooms.GetScore(ctx, session.AccessToken, scoreId)
				if err != nil {
					return osuapi.Score{}, err
				}

				if score.PP == 0 {
					return osuapi.Score{}, errors.New("score has 0 pp")
				}

				_, err = rooms.UpdateScore(ctx, roomId, score.User.ID, score.ID)
				if err != nil {
					return osuapi.Score{}, err
				}

				return score, nil
			}

			score, err := tryFind()
			if err == nil {
				refill, err := rooms.RefillForUser(ctx, session.User.OsuID, 1)
				if errors.Is(err, utils.ErrNotEnoughGuesses) {
					return echo.NewHTTPError(http.StatusBadRequest, "no guesses left")
				}

				if err != nil {
					rooms.DeleteById(ctx, roomId)
					return echo.ErrInternalServerError.Wrap(err)
				}

				return c.JSON(200, utils.Map{
					"score": utils.Map{
						"pp":         score.PP,
						"mods":       score.Mods,
						"accuracy":   score.Accuracy,
						"beatmapset": score.BeatmapSet,
						"beatmap":    score.Beatmap,
						"statistics": score.Statistics,
					},
					"refill": refill,
				})
			}
		}

		return c.JSON(http.StatusBadRequest, utils.Map{
			"message": "failed to find a score",
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

		filename := c.Param("filename")

		roomId := strings.TrimSuffix(filename, ".osr")
		room, err := rooms.FindByID(ctx, roomId)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		replay, err := client.DownloadReplay(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		r, err := rplpa.ParseReplay(replay)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		r.Username = "rankguessr"
		r.ScoreID = 0
		if r.ScoreInfo != nil {
			r.ScoreInfo.ScoreId = 0
		}

		anonymized, err := rplpa.WriteReplay(r)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.Blob(200, "application/x-osu-replay", anonymized)
	}
}

func RoomGetScore(rooms service.Rooms, guesses service.Guess, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return err
		}

		roomId := c.Param("id")
		room, err := rooms.FindByID(ctx, roomId)
		if err != nil {
			return echo.ErrNotFound.Wrap(err)
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		var guess *domain.Guess
		if room.GuessID != nil {
			g, err := guesses.FindById(ctx, string(*room.GuessID))
			if err != nil {
				return err
			}
			guess = &g
		}

		score, err := rooms.GetScore(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			err := rooms.DeleteById(ctx, room.ID)
			if err != nil {
				return err
			}

			return echo.ErrNotFound.Wrap(err)
		}

		if room.ClosesAt.Before(time.Now()) {
			player, err := client.GetUser(ctx, session.AccessToken, score.User.ID)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			_, _, err = guesses.Create(
				ctx, session.User.OsuID, domain.GuessCreate{
					PlayerID:     player.ID,
					Guess:        0,
					ScoreID:      room.ScoreID,
					BeatmapID:    score.BeatmapID,
					BeatmapSetID: score.Beatmap.BeatmapSetId,
					ActualRank:   player.Statistics.GlobalRank,
				},
			)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			err = rooms.DeleteById(ctx, room.ID)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			return echo.NewHTTPError(http.StatusBadRequest, "room is already closed")
		}

		var user *osuapi.User
		if guess != nil {
			user = &score.User
		}

		return c.JSON(200, utils.Map{
			"score": utils.Map{
				"pp":         score.PP,
				"mods":       score.Mods,
				"accuracy":   score.Accuracy,
				"beatmapset": score.BeatmapSet,
				"beatmap":    score.Beatmap,
				"statistics": score.Statistics,
				"user":       user,
			},
			"closes_at": room.ClosesAt,
			"guess":     guess,
		})
	}
}

type submitRequest struct {
	Guess int `json:"guess"`
}

func RoomSubmitGuess(rooms service.Rooms, guesses service.Guess, client *osuapi.Client) echo.HandlerFunc {
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

		if req.Guess > 3000000 || req.Guess < 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "guess must be between 1 and 3 million")
		}

		if room.UserID != session.User.OsuID {
			return echo.ErrUnauthorized
		}

		if room.GuessID != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "room is already closed")
		}

		score, err := rooms.GetScore(ctx, session.AccessToken, room.ScoreID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		player, err := client.GetUser(ctx, session.AccessToken, score.User.ID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		elo, guess, err := guesses.Create(
			ctx, session.User.OsuID, domain.GuessCreate{
				PlayerID:     player.ID,
				Guess:        req.Guess,
				ScoreID:      room.ScoreID,
				BeatmapID:    score.BeatmapID,
				BeatmapSetID: score.Beatmap.BeatmapSetId,
				ActualRank:   player.Statistics.GlobalRank,
			},
		)

		// this should never happen with new pool
		// TODO: check player rank against ranges before creating a room, it's cached anyway
		if errors.Is(err, ranking.ErrRangeNotFound) {
			err := rooms.DeleteById(ctx, roomId)
			if err != nil {
				return echo.ErrInternalServerError.Wrap(err)
			}

			return echo.NewHTTPError(http.StatusInternalServerError, "actual rank is out of range")
		}

		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		err = rooms.UpdateGuessID(ctx, room.ID, guess.ID)
		if err != nil {
			return echo.ErrInternalServerError.Wrap(err)
		}

		return c.JSON(200, utils.Map{
			"guess":   guess,
			"player":  player,
			"new_elo": elo,
		})
	}
}

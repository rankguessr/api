package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/domain"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
	"github.com/wieku/rplpa"
)

func SubmissionCreate(submissions service.Submissions, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		comment := c.FormValue("comment")
		if comment == "null" {
			comment = ""
		}

		if len(comment) > 500 {
			return echo.NewHTTPError(http.StatusBadRequest, "comment is too long").Wrap(utils.ErrLimitExceeded)
		}

		anonymous := c.FormValue("is_anonymous") == "true"

		previous, err := submissions.FindByUser(ctx, session.User.OsuID, false)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to check previous submissions").Wrap(err)
		}

		if len(previous) >= 5 {
			return echo.NewHTTPError(http.StatusBadRequest, "submission limit reached").Wrap(utils.ErrLimitExceeded)
		}

		var scoreId int
		form, err := c.MultipartForm()
		if err == nil && form.File["score_file"] != nil {
			scoreFile, _ := c.FormFile("score_file")
			src, err := scoreFile.Open()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to open score file").Wrap(err)
			}
			defer src.Close()

			if scoreFile.Size > 10*1024*1024 {
				return echo.NewHTTPError(http.StatusBadRequest, "score file is too large").Wrap(utils.ErrLimitExceeded)
			}

			data, err := io.ReadAll(src)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to read score file").Wrap(err)
			}

			replay, err := rplpa.ParseReplay(data)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to parse score file").Wrap(err)
			}

			if replay.ScoreInfo != nil {
				scoreId = int(replay.ScoreInfo.ScoreId)
			} else {
				scoreId = int(replay.ScoreID)
			}
		} else if scoreUrl := c.FormValue("score_url"); scoreUrl != "" {
			scoreId, err = utils.ParseScoreURL(scoreUrl)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid score url").Wrap(err)
			}
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, "score url or file is required")
		}

		score, err := client.GetScore(ctx, session.AccessToken, scoreId)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "failed to get score from osu api").Wrap(err)
		}

		submission, err := submissions.Create(ctx, domain.SubmissionCreate{
			UserID:       session.User.OsuID,
			PlayerID:     score.User.ID,
			ScoreID:      score.ID,
			Comment:      comment,
			IsAnonymous:  anonymous,
			BeatmapID:    score.Beatmap.ID,
			BeatmapsetID: score.Beatmap.BeatmapSetId,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create submission").Wrap(err)
		}

		return c.JSON(http.StatusCreated, submission)
	}
}

func SubmissionDelete(submissions service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		submissionId := c.Param("id")

		if !session.User.IsAdmin {
			return echo.ErrForbidden
		}

		err = submissions.Delete(ctx, submissionId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete submission").Wrap(err)
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func SubmissionSetAccepted(submissions service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		submissionId := c.Param("id")

		if !session.User.IsAdmin {
			return echo.ErrForbidden
		}

		err = submissions.SetAccepted(ctx, submissionId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to accept submission").Wrap(err)
		}

		return c.JSON(http.StatusOK, utils.Map{
			"ok": true,
		})
	}
}

// func SubmissionFindByUser(submissions service.Submissions) echo.HandlerFunc {
// 	return func(c *echo.Context) error {
// 		ctx := c.Request().Context()
// 		session, err := utils.GetSession(c)
// 		if err != nil {
// 			return echo.ErrUnauthorized.Wrap(err)
// 		}

// 		submissions, err := submissions.FindByUser(ctx, session.User.OsuID)
// 		if err != nil {
// 			return echo.NewHTTPError(http.StatusInternalServerError, "failed to find submissions").Wrap(err)
// 		}

// 		return c.JSON(http.StatusOK, submissions)
// 	}
// }

func SubmissionsFind(submissions service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		if !session.User.IsAdmin {
			return echo.NewHTTPError(http.StatusForbidden, "not an admin")
		}

		accepted := c.QueryParam("accepted") == "true"
		limit, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		if limit > 50 {
			return echo.ErrBadRequest.Wrap(utils.ErrLimitExceeded)
		}

		submissions, err := submissions.Find(ctx, accepted, limit, page)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get submissions").Wrap(err)
		}

		return c.JSON(http.StatusOK, submissions)
	}
}

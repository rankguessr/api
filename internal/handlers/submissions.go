package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rankguessr/api/internal/service"
	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/rankguessr/api/pkg/utils"
)

func SubmissionCreate(submissions service.Submissions, client *osuapi.Client) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		var req struct {
			ScoreURL string `json:"score_url"`
		}

		if err := c.Bind(&req); err != nil {
			return err
		}

		scoreId, err := utils.ParseScoreURL(req.ScoreURL)
		if err != nil {
			return c.JSON(http.StatusBadRequest, utils.Map{
				"message": "invalid score url",
			})
		}

		score, err := client.GetScore(ctx, session.AccessToken, scoreId)
		if err != nil {
			return c.JSON(http.StatusNotFound, utils.Map{
				"message": "score not found",
			})
		}

		submission, err := submissions.Create(ctx, session.User.OsuID, score.User.ID, score.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.Map{
				"message": "failed to create submission",
			})
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
			return c.JSON(http.StatusInternalServerError, utils.Map{
				"message": "failed to delete submission",
			})
		}

		return c.JSON(http.StatusNoContent, utils.Map{
			"ok": true,
		})
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
			return c.JSON(http.StatusInternalServerError, utils.Map{
				"message": "failed to accept submission",
			})
		}

		return c.JSON(http.StatusOK, utils.Map{
			"ok": true,
		})
	}
}

func SubmissionFindByUser(submissions service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		submissions, err := submissions.FindByUser(ctx, session.User.OsuID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.Map{
				"message": "failed to get submissions",
			})
		}

		return c.JSON(http.StatusOK, submissions)
	}
}

func SubmissionFindUnaccepted(submissions service.Submissions) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()
		session, err := utils.GetSession(c)
		if err != nil {
			return echo.ErrUnauthorized.Wrap(err)
		}

		if !session.User.IsAdmin {
			return echo.ErrForbidden
		}

		submissions, err := submissions.FindUnaccepted(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.Map{
				"message": "failed to get submissions",
			})
		}

		return c.JSON(http.StatusOK, submissions)
	}
}

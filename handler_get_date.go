package main

import (
	"flow-sprints/jwt"
	"flow-sprints/sprint"
	"net/http"
	"time"

	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type GetDateQueryParam struct {
	Range     *uint   `query:"range" validate:"omitempty,gte=2"`
	ProjectId *uint64 `query:"project_id" validate:"omitempty,gte=1"`
}

func getDate(c echo.Context) error {
	// Check token
	u := c.Get("user").(*jwtGo.Token)
	userId, err := jwt.CheckToken(*jwtIssuer, u)
	if err != nil {
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusNotFound, map[string]string{"message": err.Error()}, "	")
	}

	// Bind param
	dateStr := c.Param("date")

	// Validate param
	_, err = time.Parse("20060102", dateStr)
	if err != nil {
		_, err = time.Parse("2006-1-2", dateStr)
		if err != nil {
			// 404: Not found
			return echo.ErrNotFound
		}
	}

	// Bind query
	q := new(GetDateQueryParam)
	if err = c.Bind(q); err != nil {
		// 400: Bad request
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": err.Error()}, "	")
	}

	// Validate query
	if err = c.Validate(q); err != nil {
		// 422: Unprocessable entity
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusUnprocessableEntity, map[string]string{"message": err.Error()}, "	")
	}

	// Get sprints
	sprints, _, _, err := sprint.GetListDate(userId, dateStr, q.Range, q.ProjectId)
	if err != nil {
		// 500: Internal server error
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}

	if sprints == nil {
		return c.JSONPretty(http.StatusOK, []interface{}{}, "	")
	}
	return c.JSONPretty(http.StatusOK, sprints, "	")
}

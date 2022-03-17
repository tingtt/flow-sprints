package main

import (
	"flow-sprints/jwt"
	"flow-sprints/sprint"
	"fmt"
	"net/http"

	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

func post(c echo.Context) error {
	// Check `Content-Type`
	if c.Request().Header.Get("Content-Type") != "application/json" &&
		c.Request().Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		// 415: Invalid `Content-Type`
		return c.JSONPretty(http.StatusUnsupportedMediaType, map[string]string{"message": "unsupported media type"}, "	")
	}

	// Check token
	u := c.Get("user").(*jwtGo.Token)
	userId, err := jwt.CheckToken(*jwtIssuer, u)
	if err != nil {
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusUnauthorized, map[string]string{"message": err.Error()}, "	")
	}

	// Bind request body
	post := new(sprint.PostBody)
	if err = c.Bind(post); err != nil {
		// 400: Bad request
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": err.Error()}, "	")
	}

	// Validate request body
	if err = c.Validate(post); err != nil {
		// 422: Unprocessable entity
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusUnprocessableEntity, map[string]string{"message": err.Error()}, "	")
	}

	// TODO: Check parent id

	// TODO: Check project id

	p, startAfterEnd, invalidParentId, invalidChildDate, err := sprint.Post(userId, *post)
	if err != nil {
		// 500: Internal server error
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}
	if startAfterEnd {
		// 400: Bad request
		c.Logger().Debug("`start` must before `end`")
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": "`start` must before `end`"}, "	")
	}
	if invalidParentId && post.ParentId != nil {
		// 409: Conflict
		c.Logger().Debug(fmt.Sprintf("sprint id: %d does not exists", *post.ParentId))
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": fmt.Sprintf("sprint id: %d does not exists", *post.ParentId)}, "	")
	}
	if invalidChildDate {
		// 409: Conflict
		c.Logger().Debug("child must between parent")
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": "child must between parent"}, "	")
	}

	// 200: Success
	return c.JSONPretty(http.StatusOK, p, "	")
}

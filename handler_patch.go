package main

import (
	"flow-terms/jwt"
	"flow-terms/term"
	"fmt"
	"net/http"
	"strconv"

	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

func patch(c echo.Context) error {
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
		return c.JSONPretty(http.StatusNotFound, map[string]string{"message": err.Error()}, "	")
	}

	// id
	idStr := c.Param("id")

	// string -> uint64
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		// 404: Not found
		return echo.ErrNotFound
	}

	// Bind request body
	patch := new(term.Patch)
	if err = c.Bind(patch); err != nil {
		// 400: Bad request
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": err.Error()}, "	")
	}

	// Validate request body
	if err = c.Validate(patch); err != nil {
		// 422: Unprocessable entity
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusUnprocessableEntity, map[string]string{"message": err.Error()}, "	")
	}

	// TODO: Check term id

	// TODO: Check project id

	p, notFound, startAfterEnd, parentNotFound, loopParent, invalidChildDate, err := term.Update(userId, id, *patch)
	if err != nil {
		// 500: Internal server error
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}
	if notFound {
		// 404: Not found
		c.Logger().Debug("project not found")
		return echo.ErrNotFound
	}
	if startAfterEnd {
		// 400: Bad request
		c.Logger().Debug("`start` must before `end`")
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": "`start` must before `end`"}, "	")
	}
	if parentNotFound && patch.ParentId != nil {
		// 409: Conflict
		c.Logger().Debug(fmt.Sprintf("term id: %d does not exists", *patch.ParentId))
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": fmt.Sprintf("term id: %d does not exists", *patch.ParentId)}, "	")
	}
	if loopParent && patch.ParentId != nil {
		// 409: Conflict
		c.Logger().Debug(fmt.Sprintf("term id: %d does not exists", *patch.ParentId))
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": "cannot set own child"}, "	")
	}
	if invalidChildDate {
		// 409: Conflict
		c.Logger().Debug("child must between parent")
		return c.JSONPretty(http.StatusConflict, map[string]string{"message": "child must between parent"}, "	")
	}

	// 200: Success
	return c.JSONPretty(http.StatusOK, p, "	")
}

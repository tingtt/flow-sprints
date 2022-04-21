package main

import (
	"flow-sprints/jwt"
	"flow-sprints/sprint"
	"net/http"

	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

func deleteAll(c echo.Context) error {
	// Check token
	u := c.Get("user").(*jwtGo.Token)
	userId, err := jwt.CheckToken(*jwtIssuer, u)
	if err != nil {
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusUnauthorized, map[string]string{"message": err.Error()}, "	")
	}

	err = sprint.DeleteAll(userId)
	if err != nil {
		// 500: Internal server error
		c.Logger().Debug(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}

	// 204: No content
	return c.JSONPretty(http.StatusNoContent, map[string]string{"message": "Deleted"}, "	")
}

package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JwtCustumClaims struct {
	Id    uint64 `json:"id"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func CheckToken(issuer string, token *jwt.Token) (id uint64, err error) {
	claims := token.Claims.(*JwtCustumClaims)

	if !claims.VerifyIssuer(issuer, true) {
		// Invalid token
		return 0, errors.New("invalid token")
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		// Token expired
		return 0, errors.New("token expired")
	}

	return claims.Id, nil
}

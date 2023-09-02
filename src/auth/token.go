package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaim struct {
	IdUser    string `json:"us"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Nombres   string `json:"nombres"`
	Apellidos string `json:"apellidos"`
	jwt.StandardClaims
}

func ValidateToken(cookieToken string) (body JWTClaim, err error) {
	signingKey := GetKey_PrivateJwt()
	token, err := jwt.ParseWithClaims(
		cookieToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(signingKey), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("error en el token")
		return
	}
	body = *claims
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expirado")
		return
	}
	return
}

func GetValToken(token string) interface{} {
	token_verify, _ := ValidateToken(token)
	var myMap map[string]interface{}
	data, _ := json.Marshal(token_verify)

	json.Unmarshal(data, &myMap)
	return myMap
}

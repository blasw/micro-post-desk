package tokens

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type Tokenizer interface {
	NewAccessToken(UserClaims) (string, error)
	NewRefreshToken(jwt.StandardClaims) (string, error)
	ParseAccessToken(string) (*UserClaims, error)
	ParseRefreshToken(string) (*jwt.StandardClaims, error)
}

type UserClaims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

type JwtTokenizer struct {
	Logger *zap.Logger
}

func (j *JwtTokenizer) NewAccessToken(userClaims UserClaims) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	return accessToken.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
}

func (j *JwtTokenizer) NewRefreshToken(claims jwt.StandardClaims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return refreshToken.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
}

func (j *JwtTokenizer) ParseAccessToken(accessToken string) (*UserClaims, error) {
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedAccessToken.Valid {
		return nil, errors.New("invalid access token provided")
	}

	return parsedAccessToken.Claims.(*UserClaims), nil
}

func (j *JwtTokenizer) ParseRefreshToken(refreshToken string) (*jwt.StandardClaims, error) {
	parsedRefreshToken, err := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedRefreshToken.Valid {
		return nil, errors.New("invalid refresh token provided")
	}

	return parsedRefreshToken.Claims.(*jwt.StandardClaims), nil
}

package tokens

import (
	"errors"
	"fmt"
	"go-users/storage"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

type ValidationResults struct {
	Access_Token  string
	Refresh_Token string
	User_id       int
	Username      string
}

func ValidateUser(storage *storage.Storage, tokenizer Tokenizer, access_token string, refresh_token string) (*ValidationResults, error) {
	accessClaims, err := tokenizer.ParseAccessToken(access_token)
	if err == nil && accessClaims != nil {
		// Access token is valid
		user_id, _ := strconv.Atoi(accessClaims.Id)
		res := &ValidationResults{
			Access_Token:  "",
			Refresh_Token: "",
			User_id:       user_id,
			Username:      accessClaims.Username,
		}

		return res, nil
	}

	// Access token is invalid
	_, err = tokenizer.ParseRefreshToken(refresh_token)
	if err != nil {
		//Refresh token is invalid
		return nil, errors.New("Invalid tokens")
	}

	user, err := storage.GetUserByRefreshToken(refresh_token)
	if err != nil {
		return nil, errors.New("Invalid tokens")
	}

	newRefreshToken, err := tokenizer.NewRefreshToken(jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix()})
	if err != nil {
		return nil, errors.New("Internal server error")
	}

	storage.UpdateUserRefreshToken(user.Username, newRefreshToken)

	newAccessToken, err := tokenizer.NewAccessToken(UserClaims{
		Id:       fmt.Sprint(user.ID),
		Username: user.Username,
		Email:    user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	})

	res := &ValidationResults{
		User_id:       int(user.ID),
		Username:      user.Username,
		Access_Token:  newAccessToken,
		Refresh_Token: newRefreshToken,
	}

	return res, nil
}

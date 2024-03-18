package controllers

import (
	"fmt"
	"go-users/storage"
	"go-users/tokens"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/shirou/gopsutil/cpu"
	"go.uber.org/zap"
)

type AuthDto struct {
	Access_token  string `json:"access_token" binding:"required"`
	Refresh_token string `json:"refresh_token" binding:"required"`
}

type AuthSuccessResp struct {
	User_id       int    `json:"user_id"`
	Username      string `json:"username"`
	Access_token  string `json:"access_token"`
	Refresh_token string `json:"refresh_token"`
}

func Authenticate(storage *storage.Storage, tokenizer tokens.Tokenizer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authDto AuthDto
		// Bind the request body to the AuthDto struct
		if err := c.ShouldBindJSON(&authDto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if the access token is valid
		accessClaims, err := tokenizer.ParseAccessToken(authDto.Access_token)
		if err == nil && accessClaims != nil {
			// Access token is valid
			user_id, _ := strconv.Atoi(accessClaims.Id)
			resp := &AuthSuccessResp{
				User_id:  user_id,
				Username: accessClaims.Username,
			}

			c.JSON(http.StatusOK, resp)
			return
		}

		// Access token is invalid or missing, check for refresh token
		refreshClaims, err := tokenizer.ParseRefreshToken(authDto.Refresh_token)
		if err != nil && refreshClaims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tokens"})
			return
		}
		// Refresh token has valid structure
		// Generate a new access token and refresh token pair
		user, err := storage.GetUserByRefreshToken(authDto.Refresh_token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tokens"})
			return
		}

		newRefreshToken, err := tokenizer.NewRefreshToken(jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix()})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		storage.UpdateUserRefreshToken(user.Username, newRefreshToken)

		newAccessToken, err := tokenizer.NewAccessToken(tokens.UserClaims{
			Id:       fmt.Sprint(user.ID),
			Username: user.Username,
			Email:    user.Email,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Set the new access token and refresh token in the response
		resp := &AuthSuccessResp{
			User_id:       int(user.ID),
			Username:      user.Username,
			Access_token:  newAccessToken,
			Refresh_token: newRefreshToken,
		}

		c.JSON(http.StatusOK, resp)
	}
}

func GetLoadstate() gin.HandlerFunc {
	return func(c *gin.Context) {
		percent, err := cpu.Percent(time.Second, false)
		if err != nil {
			fmt.Println("Error occured while getting the cpu usage", err.Error())
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		c.JSON(200, percent)
	}
}

type GetUserByIdDto struct {
	id int `form:"id" binding:"required"`
}

func GetUserById(storage *storage.Storage, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dto GetUserByIdDto
		if err := c.ShouldBindQuery(&dto); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		}

		user, err := storage.GetUserByID(dto.id)
		if err != nil {
			c.JSON(http.StatusNoContent, err.Error())
		}

		c.JSON(http.StatusOK, user.Username)
	}
}

type GetUserByUsernameDto struct {
	username string `form:"username" binding:"required"`
}

func GetUserByUsername(storage *storage.Storage, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var dto GetUserByUsernameDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		}

		user, err := storage.GetUserByUsername(dto.username)
		if err != nil {
			c.JSON(http.StatusNoContent, err.Error())
		}

		c.JSON(http.StatusOK, user.Username)
	}
}

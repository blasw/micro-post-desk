package controllers

import (
	"fmt"
	"go-users/storage"
	"go-users/tokens"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

		res, err := tokens.ValidateUser(storage, tokenizer, authDto.Access_token, authDto.Refresh_token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err.Error())
		}

		resp := &AuthSuccessResp{
			User_id:       res.User_id,
			Username:      res.Username,
			Access_token:  res.Access_Token,
			Refresh_token: res.Refresh_Token,
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

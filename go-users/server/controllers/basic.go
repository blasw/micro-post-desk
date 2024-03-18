package controllers

import (
	"fmt"
	"go-users/storage"
	"go-users/storage/models"
	"go-users/tokens"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type signupDto struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=8,max=32"`
	Email    string `json:"email" binding:"required,email"`
}

func SignUp(storage *storage.Storage, tokenizer tokens.Tokenizer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		//validating request
		var dto signupDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// hashing password
		hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Error occured while hashing the password", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		// creating refresh token
		refreshToken, err := tokenizer.NewRefreshToken(jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix()})
		if err != nil {
			logger.Error("Error occured while creating the refresh token", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		// creating new user
		new_user := &models.User{
			Username:     dto.Username,
			Password:     string(hash),
			Email:        dto.Email,
			RefreshToken: refreshToken,
		}
		// saving new user
		new_user_id, err := storage.CreateUser(new_user)
		if err != nil {
			logger.Error("Error occured while creating the user", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		// creating access token
		accessToken, err := tokenizer.NewAccessToken(tokens.UserClaims{Id: fmt.Sprint(new_user_id), Username: new_user.Username, Email: new_user.Email, StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24).Unix()}})
		if err != nil {
			logger.Error("Error occured while creating the access token", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		// setting cookies
		c.SetCookie("access_token", accessToken, 3600*24, "/", "localhost", false, true)
		c.SetCookie("refresh_token", refreshToken, 3600*24*7, "/", "localhost", false, true)

		resp := gin.H{"username": new_user.Username, "email": new_user.Email}

		c.JSON(201, resp)
	}
}

type SignInDto struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

func SignIn(storage *storage.Storage, tokenizer tokens.Tokenizer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		//validating request
		var dto SignInDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// getting the user
		user, err := storage.GetUserByUsername(dto.Username)
		if err != nil {
			logger.Error("Error occured while getting the user", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		// validating password
		if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)) != nil {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
			return
		}

		// creating access token
		accessToken, err := tokenizer.NewAccessToken(tokens.UserClaims{Id: fmt.Sprint(user.ID), Username: user.Username, Email: user.Email, StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24).Unix()}})
		if err != nil {
			logger.Error("Error occured while creating the access token", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		refreshToken, err := tokenizer.NewRefreshToken(jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix()})
		if err != nil {
			logger.Error("Error occured while creating the refresh token", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		err = storage.UpdateUserRefreshToken(user.Username, refreshToken)
		if err != nil {
			logger.Debug("Error occured when creating refresh token", zap.String("Error: ", err.Error()))
			c.JSON(500, gin.H{"error": "Internal server error"})
		}

		c.SetCookie("access_token", accessToken, 3600*24, "/", "localhost", false, true)
		c.SetCookie("refresh_token", refreshToken, 3600*24*7, "/", "localhost", false, true)

		resp := gin.H{"username": user.Username, "email": user.Email}

		c.JSON(200, resp)
	}
}

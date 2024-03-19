package server

import (
	"go-users/server/controllers"
	"go-users/storage"
	"go-users/tokens"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	Engine    *gin.Engine
	Storage   *storage.Storage
	Tokenizer tokens.Tokenizer
	Logger    *zap.Logger
}

func CreateServer(s *storage.Storage, tokenizer tokens.Tokenizer, logger *zap.Logger) *Server {
	return &Server{Engine: gin.Default(), Storage: s, Tokenizer: tokenizer, Logger: logger}
}

func (s *Server) SetupRoutes() {
	s.Engine.POST("/users/signup", controllers.SignUp(s.Storage, s.Tokenizer, s.Logger))
	s.Engine.POST("/users/signin", controllers.SignIn(s.Storage, s.Tokenizer, s.Logger))
	s.Engine.GET("/users/stats", controllers.GetStats(s.Storage, s.Tokenizer, s.Logger))

	//rabbitmq side -->
	s.Engine.POST("/users/auth", controllers.Authenticate(s.Storage, s.Tokenizer, s.Logger))
	s.Engine.GET("/users/getbyid", controllers.GetUserById(s.Storage, s.Logger))
	s.Engine.GET("/users/getbyusername", controllers.GetUserByUsername(s.Storage, s.Logger))
	s.Engine.GET("/users/load", controllers.GetLoadstate())
}

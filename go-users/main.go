package main

import (
	"go-users/server"
	"go-users/storage"
	"go-users/tokens"
	"os"

	"go.uber.org/zap"
)

func setupService() server.Server {
	logger, _ := zap.NewDevelopment()
	logger.Info("Starting go-users...")
	storage := &storage.Storage{Logger: logger}
	storage.Init(os.Getenv("DB_ADDR"))

	tokenizer := &tokens.JwtTokenizer{Logger: logger}

	server := server.CreateServer(storage, tokenizer, logger)
	server.SetupRoutes()
	logger.Info("Go-users is ready to be launched")

	return *server
}

func main() {
	usersService := setupService()
	usersService.Engine.Run(":5000")
}

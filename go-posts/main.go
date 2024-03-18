package main

import (
	"go-posts/cache"
	"go-posts/server"
	"go-posts/storage"
	"go-posts/utils"

	"github.com/charmbracelet/log"
)

func init() {
	utils.CheckENVS()
}

func setupService() *server.Server {
	store := &storage.PostgreStore{}
	store.CreateStorage()
	store.Migrate()

	cache := &cache.RedisCache{}
	cache.ConnectCache()

	server := server.CreateService(store, cache)
	server.SetupRoutes()

	return server
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("Starting the Go-Posts...")

	server := setupService()

	server.Run(5000)
}

package server

import (
	"fmt"
	"go-posts/cache"
	"go-posts/server/controllers"
	"go-posts/storage"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Store  storage.Storage
	Cache  *cache.RedisCache
	Engine *gin.Engine
}

func CreateService(store storage.Storage, cache *cache.RedisCache) *Server {
	return &Server{
		Store:  store,
		Cache:  cache,
		Engine: gin.Default(),
	}
}

func (s *Server) SetupRoutes() {
	log.Debug("Setting up routes...")
	// Loadbalancer ----
	s.Engine.GET("/posts/load", controllers.GetLoadState())

	// Free ----
	s.Engine.GET("/posts/latest", controllers.GetLatestPosts(s.Store, s.Cache))
	s.Engine.GET("/posts/mostliked", controllers.GetMostLikedPosts(s.Store, s.Cache))

	// Protected
	s.Engine.GET("/posts/user", controllers.GetUsersPosts(s.Store, s.Cache))
	s.Engine.POST("/posts/new", controllers.CreatePost(s.Store, s.Cache))
	s.Engine.DELETE("/posts/delete", controllers.DeletePost(s.Store, s.Cache))

	//TODO Should be moved to likes
	// s.Engine.PATCH("/posts/like", controllers.LikePost(s.Store, s.Cache))
	// s.Engine.PATCH("/posts/unlike", controllers.UnlikePost(s.Store, s.Cache))
}

func (s *Server) Run(basePort int) {
	for {
		log.Debug("Trying to start the server", "port", basePort)
		err := s.Engine.Run(fmt.Sprintf(":%d", basePort))
		if err != nil {
			log.Debug("Port is busy, trying another one", "port", basePort)
			basePort++
		} else {
			log.Info("Server is running", "port", basePort)
			break
		}
	}
}

func (s *Server) GeneratePingings(amount int) bool {
	for i := 0; i < amount; i++ {
		fmt.Println("Pinging...")
	}
	return true
}

package controllers

import (
	"encoding/json"
	"fmt"
	"go-posts/cache"
	"go-posts/server/middleware"
	"go-posts/storage"
	"go-posts/storage/models"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

// TODO Should be removed
func Test(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}

type CreatePostDto struct {
	Title string `json:"title" binding:"required,min=2,max=50"`
	Body  string `json:"body" binding:"required,min=2,max=350"`
}

// TODO Should somehow extract author_id from jwt roken instead of having it inside the dto
func CreatePost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePostDto
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("Unable to bind json: handlers.CreatePost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, v := middleware.ValidateUser(c)

		if !v {
			c.JSON(http.StatusUnauthorized, "Not authorized/invalid tokens")
			return
		}

		post := &models.Post{
			Title:      req.Title,
			Body:       req.Body,
			AuthorID:   user.User_Id,
			LikesCount: 0,
		}

		err := store.CreatePost(post)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, post)
	}
}

type GetLatestPostDto struct {
	PageID   int `form:"pageid" binding:"required,min=1"`
	PageSize int `form:"pagesize" binding:"required,min=1"`
}

func GetLatestPosts(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GetLatestPostDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind query: handlers.GetLatestPosts()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key := fmt.Sprintf("%d,%d,latest", req.PageSize, req.PageID)

		resJson, err := cache.Client.HGet(c, key, "posts").Result()
		if err == nil {
			log.Debug("Result was found in cache, returning it...")
			var res []models.Post
			json.Unmarshal([]byte(resJson), &res)
			c.JSON(http.StatusOK, res)
			return
		}

		log.Debug("Result was not found in cache, getting from the database...")

		posts := store.GetLatestPosts(req.PageSize, (req.PageID-1)*req.PageSize)

		go func() {
			postsJSON, _ := json.Marshal(posts)

			_, err = cache.Client.HSet(c, key, "posts", postsJSON).Result()
			if err != nil {
				log.Error("Unable to set data to cache: ", err)
			} else {
				_, err := cache.Client.Expire(c, key, 60*time.Second).Result()
				if err != nil {
					log.Error("Unable to set expiration time, running undo changes")
					_, err := cache.Client.HDel(c, key).Result()
					if err != nil {
						log.Error("Unable to delete invalid data, delete manually, key: ", key)
					}
				}
			}
		}()

		c.JSON(http.StatusOK, posts)
	}
}

type GetMostLikedPostsDto struct {
	PageID   uint `form:"pageid" binding:"required,min=1"`
	PageSize uint `form:"pagesize" binding:"required,min=1"`
}

func GetMostLikedPosts(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GetMostLikedPostsDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind query: handlers.GetMostLikedPosts()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key := fmt.Sprintf("%d,%d,mostliked", req.PageSize, req.PageID)

		resJson, err := cache.Client.HGet(c, key, "posts").Result()
		if err == nil {
			log.Debug("Result was found in cache, returning it...")
			var res []models.Post
			json.Unmarshal([]byte(resJson), &res)
			c.JSON(http.StatusOK, res)
			return
		}

		log.Debug("Result was not found in cache, getting from the database...")

		posts := store.GetMostLikedPosts(int(req.PageSize), int((req.PageID-1)*req.PageSize))

		go func() {
			postsJSON, _ := json.Marshal(posts)

			_, err = cache.Client.HSet(c, key, "posts", postsJSON).Result()
			if err != nil {
				log.Error("Unable to set data to cache: ", err)
			} else {
				_, err := cache.Client.Expire(c, key, 60*time.Second).Result()
				if err != nil {
					log.Error("Unable to set expiration time, running undo changes")
					_, err := cache.Client.HDel(c, key).Result()
					if err != nil {
						log.Error("Unable to delete invalid data, delete manually, key: ", key)
					}
				}
			}
		}()

		c.JSON(http.StatusOK, posts)
	}
}

type GetUsersPostsDto struct {
	PageID   uint `form:"pageid" binding:"required,min=1"`
	PageSize uint `form:"pagesize" binding:"required,min=1"`
	AuthorID uint `form:"author_id" binding:"required"`
}

// TODO Should only be available for Author
func GetUsersPosts(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GetUsersPostsDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind query: handlers.GetUsersPosts()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//if !middleware.ValidateUser(c) {
		//c.JSON(http.StatusUnauthorized, "Not authorized/invalid tokens")
		//}

		posts := store.GetUsersPosts(int(req.PageSize), int((req.PageID-1)*req.PageSize), req.AuthorID)
		c.JSON(http.StatusOK, posts)
	}
}

// TODO Probably should not exist??
type GetPostDto struct {
	PostID uint `form:"post_id" binding:"required"`
}

func GetPost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GetPostDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind query: handlers.GetPost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		post := store.GetPost(req.PostID)
		c.JSON(http.StatusOK, post)
	}
}

// TODO Should be moved to go-likes
type LikePostDto struct {
	PostID uint `form:"post_id" binding:"required"`
}

func LikePost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LikePostDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind json: handlers.LikePost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := store.LikePost(req.PostID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}

// TODO Should be moved to go-users
type UnlikePostDto struct {
	PostID uint `form:"post_id" binding:"required"`
}

func UnlikePost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UnlikePostDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind json: handlers.UnlikePost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := store.UnlikePost(req.PostID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}

type DeletePostDto struct {
	PostID uint `form:"post_id" binding:"required"`
}

// TODO Should only be available for the Author. Currently is available for everyone
func DeletePost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeletePostDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind json: handlers.DeletePost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//if !middleware.ValidateUser(c) {
		//c.JSON(http.StatusUnauthorized, "Not authorized/invalid tokens")
		//}

		err := store.DeletePost(req.PostID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	}
}

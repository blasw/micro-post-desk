package controllers

import (
	"encoding/json"
	"fmt"
	"go-posts/cache"
	"go-posts/server/middleware"
	"go-posts/storage"
	"go-posts/storage/models"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

type CreatePostDto struct {
	Title string `json:"title" binding:"required,min=2,max=50"`
	Body  string `json:"body" binding:"required,min=2,max=350"`
}

func CreatePost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePostDto
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("Unable to bind json: handlers.CreatePost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
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
}

func GetUsersPosts(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		//validating request
		var req GetUsersPostsDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind query: handlers.GetUsersPosts()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		user, isValid := middleware.ValidateUser(c)
		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Not authorized / invalid tokens"})
			return
		}

		author_id := user.User_Id

		posts := store.GetUsersPosts(int(req.PageSize), int((req.PageID-1)*req.PageSize), author_id)
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
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		post := store.GetPost(req.PostID)
		c.JSON(http.StatusOK, post)
	}
}

type DeletePostDto struct {
	PostID uint `form:"post_id" binding:"required"`
}

// TODO Should only be available for the Author. Currently is available for everyone
func DeletePost(store storage.Storage, cache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		//validating request
		var req DeletePostDto
		if err := c.ShouldBindQuery(&req); err != nil {
			log.Error("Unable to bind json: handlers.DeletePost()", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		//validating user
		user, isValid := middleware.ValidateUser(c)
		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Not authorized / invalid tokens"})
			return
		}

		//checking if the user is actually an author of the post
		post := store.GetPost(req.PostID)
		if post.AuthorID != user.User_Id {
			c.JSON(http.StatusForbidden, gin.H{"Error": "Unable to delete other user's posts"})
			return
		}

		err := store.DeletePost(req.PostID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Message": "Success"})
	}
}

type CountPostsDto struct {
	User_id string `form:"id" binding:"required,min=1"`
}

func CountPosts(storage storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CountPostsDto
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		id, err := strconv.Atoi(req.User_id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		amount := storage.CountPosts(uint(id))
		c.JSON(http.StatusOK, gin.H{"amount": amount})
	}
}

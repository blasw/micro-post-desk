package storage

import (
	"errors"
	"go-posts/storage/models"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage interface {
	CreateStorage()
	Migrate()
	CreatePost(post *models.Post) error
	GetLatestPosts(limit int, offset int) []models.Post
	GetUsersPosts(limit int, offset int, authorID uint) []models.Post
	GetMostLikedPosts(limit int, offset int) []models.Post
	GetPost(postID uint) models.Post
	DeletePost(postID uint) error
	LikePost(postID uint) error
	UnlikePost(postID uint) error
}

type PostgreStore struct {
	Conn *gorm.DB
}

// CreateStorage creates *gorm.DB instance with connection to the database
func (store *PostgreStore) CreateStorage() {
	log.Debug("Creating storage...")

	// dsn := "postgres://postgres:secret@localhost:5000?sslmode=disable"
	dsn := os.Getenv("DB_ADDR")
	log.Debug("--------DB_ADDR:------", dsn)

	for i := 0; i < 3; i++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			store.Conn = db
			return
		}
		time.Sleep(1 * time.Second)
	}

	log.Fatal("Unable to reach database")
}

// Migrate automatically migrates models to the database (Post model in this case)
func (store *PostgreStore) Migrate() {
	log.Debug("AutoMigrating models...")
	store.Conn.AutoMigrate(&models.Post{})
}

func (store *PostgreStore) CreatePost(post *models.Post) error {
	pk := store.Conn.Create(post)
	if pk == nil {
		return errors.New("Failed to create post")
	}

	return nil
}

func (store *PostgreStore) GetLatestPosts(limit int, offset int) []models.Post {
	var posts []models.Post
	store.Conn.Limit(limit).Offset(offset).Find(&posts).Order("id desc")
	return posts
}

func (store *PostgreStore) GetMostLikedPosts(limit int, offset int) []models.Post {
	var posts []models.Post
	store.Conn.Limit(limit).Offset(offset).Find(&posts).Order("likes_count desc")
	return posts
}

func (store *PostgreStore) GetUsersPosts(limit int, offset int, authorID uint) []models.Post {
	var posts []models.Post
	store.Conn.Limit(limit).Offset(offset).Find(&posts).Where("author_id = ?", authorID)
	return posts
}

func (store *PostgreStore) GetPost(postID uint) models.Post {
	var post models.Post
	store.Conn.First(&post, postID)
	return post
}

func (store *PostgreStore) DeletePost(postID uint) error {
	store.Conn.Delete(&models.Post{}, postID)
	return nil
}

func (store *PostgreStore) LikePost(postID uint) error {
	post := store.GetPost(postID)
	post.LikesCount++
	store.Conn.Save(&post)
	return nil
}

func (store *PostgreStore) UnlikePost(postID uint) error {
	post := store.GetPost(postID)
	post.LikesCount--
	store.Conn.Save(&post)
	return nil
}

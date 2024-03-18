package storage

import (
	"errors"
	"go-users/storage/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db     *gorm.DB
	Logger *zap.Logger
}

// Init initializes the PostgreStorage and connects to the given database
func (st *Storage) Init(dsn string) {
	st.Logger.Debug("Conncting to the database...", zap.String("dsn: ", dsn))

	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if db == nil {
		st.Logger.Error("Error occured while connecting to the database", zap.String("Erorr: ", err.Error()))
		panic("Failed to connect to the database")
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		st.Logger.Error("Error occured while migrating models", zap.String("Erorr: ", err.Error()))
		panic(err)
	}

	st.Logger.Info("Successfully connected to the database")

	st.db = db
}

func (st *Storage) CreateUser(user *models.User) (uint, error) {
	res := st.db.Create(user)
	if res.Error != nil {
		return 0, res.Error
	}
	return user.ID, nil
}

func (st *Storage) GetUserByUsername(username string) (*models.User, error) {
	var user *models.User
	res := st.db.First(&user, "username", username)
	if res.Error != nil {
		return nil, res.Error
	}
	return user, nil
}

func (st *Storage) GetUserByID(id int) (*models.User, error) {
	var user *models.User
	res := st.db.First(&user, "id", id)
	if res.Error != nil {
		return nil, res.Error
	}

	return user, nil
}

func (st *Storage) UpdateUserRefreshToken(username string, new_token string) error {
	var user *models.User
	res := st.db.First(&user, "username", username)
	if res.Error != nil {
		return res.Error
	}

	res = st.db.Model(&user).Update("RefreshToken", new_token)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (st *Storage) GetUserByRefreshToken(token string) (*models.User, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}

	var user *models.User
	res := st.db.First(&user, "refresh_token", token)
	if res.Error != nil {
		return nil, res.Error
	}

	return user, nil
}

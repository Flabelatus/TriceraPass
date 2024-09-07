package repositories

import (
	"TriceraPass/internal/models"
	"sync"

	"gorm.io/gorm"
)

type GORMRepo struct {
	DB    *gorm.DB
	Mutex sync.Mutex
}

func (repo *GORMRepo) Migrate() error {
	err := repo.DB.AutoMigrate(
		&models.User{},
		&models.UserConfirmation{},
		&models.PasswordRestToken{},
		&models.Mode{},
		&models.ProfileImage{},
	)
	if err != nil {
		return err
	}
	return nil
}

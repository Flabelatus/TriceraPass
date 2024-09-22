package repositories

import (
	"TriceraPass/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (r *GORMRepo) UpdateUser(id string, user *models.User) (*models.User, error) {
	var existingUser *models.User
	err := r.DB.Where("id = ?", id).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	tx := r.DB.Begin()
	tx.SavePoint("beforeUserUpdate")
	err = tx.Model(&existingUser).Updates(user).Error
	if err != nil {
		tx.RollbackTo("beforeUserUpdate")
		return nil, err
	}
	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *GORMRepo) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := r.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *GORMRepo) GetUserByID(id string) (*models.User, error) {
	var currentUser *models.User
	err := r.DB.Preload("Mode").First(&currentUser, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return currentUser, nil
}

func (r *GORMRepo) GetUserByEmail(email string) (*models.User, error) {
	var user *models.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *GORMRepo) CreateAdminUser(user *models.User) error {
	tx := r.DB.Begin()
	tx.SavePoint("beforeCreateAdminUser")
	if err := tx.Create(&user).Error; err != nil {
		tx.RollbackTo("beforeCreateAdminUser")
		return err
	}
	tx.Commit()
	return nil
}

func (r *GORMRepo) CreateUser(user *models.User) (string, error) {
	var currUser models.User

	if err := r.DB.Where("email = ?", user.Email).First(&currUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx := r.DB.Begin()
			tx.SavePoint("beforeCreateUser")
			if err := tx.Create(&user).Error; err != nil {
				tx.RollbackTo("beforeCreateUser")
				return "", err
			}
			tx.Commit()
			return user.ID, nil
		}
		return "", err
	}

	return "", fmt.Errorf("user already exists")
}

func (r *GORMRepo) DeleteUserByID(id string) error {
	user := &models.User{}

	if err := r.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return err
	}
	tx := r.DB.Begin()
	tx.SavePoint("beforeDeleteUser")
	if err := tx.Delete(user, "id = ?", id).Error; err != nil {
		tx.RollbackTo("beforeDeleteUser")
		return err
	}
	tx.Commit()
	return nil
}

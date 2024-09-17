package repositories

import (
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (r *GORMRepo) GetAllModes() ([]models.Mode, error) {
	var modes []models.Mode
	err := r.DB.Find(&modes).Error
	if err != nil {
		return nil, err
	}
	return modes, nil
}

func (r *GORMRepo) GetModeByID(id int) (*models.Mode, error) {
	var mode *models.Mode
	err := r.DB.First(&mode, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("mode not found")
		}
		return nil, err
	}
	return mode, nil
}

func (r *GORMRepo) CreateMode(mode *models.Mode) error {

	tx := r.DB.Begin()
	tx.SavePoint("beforeCreateMode")
	if err := tx.Create(&mode).Error; err != nil {
		tx.RollbackTo("beforeCreateMode")
		return err
	}
	tx.Commit()
	return nil
}

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

func (r *GORMRepo) DeleteModeByID(id int) error {
	mode := &models.Mode{}

	if err := r.DB.Where("id = ?", id).First(&mode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mode not found")
		}
		return err
	}

	tx := r.DB.Begin()
	tx.SavePoint("beforeDeleteMode")
	if err := tx.Delete(&mode, "id = ?", id).Error; err != nil {
		tx.RollbackTo("beforeDeleteMode")
		return err
	}
	return nil
}

func (r *GORMRepo) UpdateMode(id int, mode *models.Mode) error {
	var existingMode *models.Mode
	err := r.DB.First(&existingMode, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mode not found")
		}
		return err
	}

	tx := r.DB.Begin()
	tx.SavePoint("beforeUpdateMode")
	err = tx.Model(&existingMode).Updates(mode).Error
	if err != nil {
		tx.RollbackTo("beforeUpdateMode")
		return err
	}
	return nil
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

func (r *GORMRepo) InsertConfirmation(confirmationModel *models.UserConfirmation) (string, error) {
	tx := r.DB.Begin()
	tx.SavePoint("beforeConfirmationInsert")
	err := r.DB.Create(&confirmationModel).Error
	if err != nil {
		tx.RollbackTo("beforeConfirmationInsert")
		return "", err
	}
	tx.Commit()
	return confirmationModel.ID, nil
}

func (r *GORMRepo) GetLastConfirmation(userID string) (*models.UserConfirmation, error) {
	var userConfirmation *models.UserConfirmation
	err := r.DB.Where("user_id = ?", userID).Order("expired_at DESC").First(&userConfirmation).Error
	if err != nil {
		return nil, err
	}
	return userConfirmation, nil
}

func (r *GORMRepo) GetConfirmationsByUserID(userID string) ([]*models.UserConfirmation, error) {
	var userConfirmations []*models.UserConfirmation
	err := r.DB.Where("user_id = ?", userID).Find(&userConfirmations).Error
	if err != nil {
		return nil, err
	}
	return userConfirmations, nil
}

func (r *GORMRepo) ConfirmUser(confirmationID string, confirmation *models.UserConfirmation) error {
	var existingConfirmation models.UserConfirmation
	if err := r.DB.Where("ID = ?", confirmationID).First(&existingConfirmation).Error; err != nil {
		return err // Handle record not found or other errors
	}

	// Start a transaction
	tx := r.DB.Begin()

	if err := tx.Model(&existingConfirmation).Updates(confirmation).Error; err != nil {
		tx.Rollback() // Rollback the transaction in case of errors
		return err
	}

	// Commit the transaction if everything went well
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *GORMRepo) GetConfirmationByID(confirmationID string) (*models.UserConfirmation, error) {
	var userConfirmation *models.UserConfirmation
	err := r.DB.First(&userConfirmation, confirmationID).Error
	if err != nil {
		return nil, err
	}
	return userConfirmation, nil
}

func (r *GORMRepo) InsertPasswordToken(passwordToken *models.PasswordRestToken) (string, error) {
	tx := r.DB.Begin()
	tx.SavePoint("beforeTokenInsert")
	err := r.DB.Create(&passwordToken).Error
	if err != nil {
		tx.RollbackTo("beforeTokenInsert")
		return "", err
	}
	tx.Commit()
	return passwordToken.ID, nil
}

func (r *GORMRepo) GetLastPasswordTokenByUserID(userID string) (*models.PasswordRestToken, error) {
	var passwordToken *models.PasswordRestToken
	err := r.DB.Where("user_id = ?", userID).Order("created_at DESC").First(&passwordToken).Error
	if err != nil {
		return nil, err
	}
	return passwordToken, nil
}

func (r *GORMRepo) GetPasswordTokenByID(tokenID string) (*models.PasswordRestToken, error) {
	var passwordToken *models.PasswordRestToken
	if err := r.DB.First(&passwordToken, tokenID).Error; err != nil {
		return nil, err
	}
	return passwordToken, nil
}

func (r *GORMRepo) SetTokenToUsed(tokenID string, passwordToken *models.PasswordRestToken) error {
	var existingToken models.PasswordRestToken
	if err := r.DB.Where("ID = ?", tokenID).First(&existingToken).Error; err != nil {
		return err
	}

	tx := r.DB.Begin()
	if err := tx.Model(&existingToken).Updates(passwordToken).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *GORMRepo) ChangePasswordByUserID(userID, newPassword string) error {
	var user *models.User
	if err := r.DB.Where("ID = ?", userID).First(&user).Error; err != nil {
		return err
	}

	// Update the password field
	hashedPassword, err := controllers.HashAPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// Save the updated user back to the database
	if err := r.DB.Save(&user).Error; err != nil {
		return err
	}
	r.DB.Logger.LogMode(logger.LogLevel(1))
	return nil
}

func (r *GORMRepo) GetUserPasswordByID(userID string) (string, error) {
	var user *models.User
	if err := r.DB.Where("ID = ?", userID).First(&user).Error; err != nil {
		return "", err
	}
	return user.Password, nil
}

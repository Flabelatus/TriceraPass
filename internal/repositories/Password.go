package repositories

import (
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/internal/models"

	"gorm.io/gorm/logger"
)

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

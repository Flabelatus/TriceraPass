package repositories

import "TriceraPass/internal/models"

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

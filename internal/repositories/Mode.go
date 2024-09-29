package repositories

import (
	"TriceraPass/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
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

func (r *GORMRepo) DeleteModeByID(id int) error {
	mode := &models.Mode{}

	if err := r.DB.Where("id = ?", id).First(&mode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mode not found")
		}
		return err
	}

	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tx.SavePoint("beforeDeleteMode")

	if err := tx.Delete(&mode, "id = ?", id).Error; err != nil {
		tx.RollbackTo("beforeDeleteMode")
		return err
	}

	if err := tx.Commit().Error; err != nil {
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

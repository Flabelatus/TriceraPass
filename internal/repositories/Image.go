package repositories

import "TriceraPass/internal/models"


func (r *GORMRepo) InsertProfileImage(image *models.ProfileImage) (*models.ProfileImage, error) {
	result := r.DB.Create(&image)
	if result.Error != nil {
		return nil, result.Error
	}

	return image, nil
}

func (r *GORMRepo) GetProfileImageByUserID(userID string) (*models.ProfileImage, error) {
	var image models.ProfileImage

	// Order by ID in descending order to get the latest image
	err := r.DB.Where("user_id = ?", userID).Order("id DESC").First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *GORMRepo) GetProfileImageByFilename(filename string) (*models.ProfileImage, error) {
	var image *models.ProfileImage

	err := r.DB.Where("filename = ?", filename).First(&image).Error
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (r *GORMRepo) DeleteProfileImageByFilename(filename string) error {
	var image *models.ProfileImage
	err := r.DB.Where("filename = ?", filename).First(&image).Error
	if err != nil {
		return err
	}

	tx := r.DB.Begin()
	tx.SavePoint("beforeDeleteImage")

	err = tx.Delete(image, filename).Error

	if err != nil {
		tx.RollbackTo("beforeDeleteImage")
		return err
	}

	return nil
}

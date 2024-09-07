package models

import "gorm.io/gorm"

type ProfileImage struct {
	gorm.Model
	Filename string `json:"filename"`
	FilePath string `json:"filepath"`
	UserID   string `json:"user_id"`
}

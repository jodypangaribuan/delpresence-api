package repository

import (
	"errors"

	"delpresence-api/internal/models"

	"gorm.io/gorm"
)

// LecturerRepository adalah interface untuk operasi repository dosen
type LecturerRepository interface {
	FindByID(id uint) (*models.Lecturer, error)
	FindByCampusUserID(campusUserID uint) (*models.Lecturer, error)
	FindByUserID(userID uint) (*models.Lecturer, error)
	Create(lecturer *models.Lecturer) error
	Update(lecturer *models.Lecturer) error
	Delete(id uint) error
}

// lecturerRepository implementasi dari LecturerRepository
type lecturerRepository struct {
	db *gorm.DB
}

// NewLecturerRepository membuat instance baru dari LecturerRepository
func NewLecturerRepository(db *gorm.DB) LecturerRepository {
	return &lecturerRepository{
		db: db,
	}
}

// FindByID mencari dosen berdasarkan ID
func (r *lecturerRepository) FindByID(id uint) (*models.Lecturer, error) {
	var lecturer models.Lecturer
	if err := r.db.Where("id = ?", id).First(&lecturer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lecturer, nil
}

// FindByCampusUserID mencari dosen berdasarkan campus_user_id
func (r *lecturerRepository) FindByCampusUserID(campusUserID uint) (*models.Lecturer, error) {
	var lecturer models.Lecturer
	if err := r.db.Where("user_id = ?", campusUserID).First(&lecturer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lecturer, nil
}

// FindByUserID mencari dosen berdasarkan user_id
func (r *lecturerRepository) FindByUserID(userID uint) (*models.Lecturer, error) {
	var lecturer models.Lecturer
	if err := r.db.Where("lecturer_user_id = ?", userID).First(&lecturer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lecturer, nil
}

// Create membuat record dosen baru
func (r *lecturerRepository) Create(lecturer *models.Lecturer) error {
	return r.db.Create(lecturer).Error
}

// Update memperbarui data dosen
func (r *lecturerRepository) Update(lecturer *models.Lecturer) error {
	return r.db.Save(lecturer).Error
}

// Delete menghapus data dosen berdasarkan ID
func (r *lecturerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Lecturer{}, id).Error
}

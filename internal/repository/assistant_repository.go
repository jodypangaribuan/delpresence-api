package repository

import (
	"delpresence-api/internal/models"

	"gorm.io/gorm"
)

// AssistantRepository adalah interface untuk operasi repository asisten dosen
type AssistantRepository interface {
	FindByID(id uint) (*models.Assistant, error)
	FindByCampusUserID(campusUserID uint) (*models.Assistant, error)
	FindByUserID(userID uint) (*models.Assistant, error)
	Create(assistant *models.Assistant) error
	Update(assistant *models.Assistant) error
	Delete(id uint) error
}

// assistantRepository implementasi dari AssistantRepository
type assistantRepository struct {
	db *gorm.DB
}

// NewAssistantRepository membuat instance baru dari AssistantRepository
func NewAssistantRepository(db *gorm.DB) AssistantRepository {
	return &assistantRepository{
		db: db,
	}
}

// FindByID mencari asisten dosen berdasarkan ID
func (r *assistantRepository) FindByID(id uint) (*models.Assistant, error) {
	var assistant models.Assistant
	if err := r.db.Where("id = ?", id).First(&assistant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &assistant, nil
}

// FindByCampusUserID mencari asisten dosen berdasarkan campus user ID
func (r *assistantRepository) FindByCampusUserID(campusUserID uint) (*models.Assistant, error) {
	var assistant models.Assistant
	if err := r.db.Where("campus_user_id = ?", campusUserID).First(&assistant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &assistant, nil
}

// FindByUserID mencari asisten dosen berdasarkan assistant user ID
func (r *assistantRepository) FindByUserID(userID uint) (*models.Assistant, error) {
	var assistant models.Assistant
	if err := r.db.Where("assistant_user_id = ?", userID).First(&assistant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &assistant, nil
}

// Create menyimpan asisten dosen baru ke database
func (r *assistantRepository) Create(assistant *models.Assistant) error {
	return r.db.Create(assistant).Error
}

// Update memperbarui data asisten dosen di database
func (r *assistantRepository) Update(assistant *models.Assistant) error {
	return r.db.Save(assistant).Error
}

// Delete menghapus asisten dosen dari database
func (r *assistantRepository) Delete(id uint) error {
	return r.db.Delete(&models.Assistant{}, id).Error
}

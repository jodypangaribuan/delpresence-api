package repository

import (
	"errors"
	"log"

	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"gorm.io/gorm"
)

var (
	ErrLectureNotFound      = errors.New("lecture not found")
	ErrLectureAlreadyExists = errors.New("lecture already exists")
)

// LectureRepository handles database operations for lectures
type LectureRepository struct {
	DB *gorm.DB
}

// NewLectureRepository creates a new instance of LectureRepository
func NewLectureRepository() *LectureRepository {
	return &LectureRepository{
		DB: database.DB,
	}
}

// CreateLecture creates a new lecture
func (r *LectureRepository) CreateLecture(lecture *models.Lecture) error {
	// Check if lecture with NIP already exists
	var count int64
	if err := r.DB.Model(&models.Lecture{}).Where("nip = ?", lecture.NIP).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrLectureAlreadyExists
	}

	return r.DB.Create(lecture).Error
}

// GetLectureByID retrieves a lecture by ID
func (r *LectureRepository) GetLectureByID(id uint) (*models.Lecture, error) {
	var lecture models.Lecture
	if err := r.DB.Preload("User").First(&lecture, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLectureNotFound
		}
		return nil, err
	}
	return &lecture, nil
}

// GetLectureByUserID retrieves a lecture by user ID
func (r *LectureRepository) GetLectureByUserID(userID uint) (*models.Lecture, error) {
	var lecture models.Lecture
	if err := r.DB.Preload("User").Where("user_id = ?", userID).First(&lecture).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLectureNotFound
		}
		return nil, err
	}
	return &lecture, nil
}

// GetLectureByNIP retrieves a lecture by NIP
func (r *LectureRepository) GetLectureByNIP(nip string) (*models.Lecture, error) {
	var lecture models.Lecture

	// Debug: Print the SQL query
	tx := r.DB.Preload("User").Where("nip = ?", nip)
	log.Printf("GetLectureByNIP SQL: %v", tx.Statement.SQL.String())
	log.Printf("GetLectureByNIP params: %v", nip)

	if err := tx.First(&lecture).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("GetLectureByNIP: No lecture found with NIP: %s", nip)
			return nil, ErrLectureNotFound
		}
		log.Printf("GetLectureByNIP error: %v", err)
		return nil, err
	}

	log.Printf("GetLectureByNIP success: Found lecture with ID: %d, NIP: %s", lecture.ID, lecture.NIP)
	return &lecture, nil
}

// UpdateLecture updates a lecture's information
func (r *LectureRepository) UpdateLecture(lecture *models.Lecture) error {
	return r.DB.Save(lecture).Error
}

// DeleteLecture deletes a lecture
func (r *LectureRepository) DeleteLecture(lectureID uint) error {
	return r.DB.Delete(&models.Lecture{}, lectureID).Error
}

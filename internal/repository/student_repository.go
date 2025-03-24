package repository

import (
	"errors"

	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"gorm.io/gorm"
)

var (
	ErrStudentNotFound      = errors.New("student not found")
	ErrStudentAlreadyExists = errors.New("student already exists")
)

// StudentRepository handles database operations for students
type StudentRepository struct {
	DB *gorm.DB
}

// NewStudentRepository creates a new instance of StudentRepository
func NewStudentRepository() *StudentRepository {
	return &StudentRepository{
		DB: database.DB,
	}
}

// CreateStudent creates a new student
func (r *StudentRepository) CreateStudent(student *models.Student) error {
	// Check if student with NIM already exists
	var count int64
	if err := r.DB.Model(&models.Student{}).Where("nim = ?", student.NIM).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrStudentAlreadyExists
	}

	return r.DB.Create(student).Error
}

// GetStudentByID retrieves a student by ID
func (r *StudentRepository) GetStudentByID(id uint) (*models.Student, error) {
	var student models.Student
	if err := r.DB.Preload("User").First(&student, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}
	return &student, nil
}

// GetStudentByUserID retrieves a student by user ID
func (r *StudentRepository) GetStudentByUserID(userID uint) (*models.Student, error) {
	var student models.Student
	if err := r.DB.Preload("User").Where("user_id = ?", userID).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}
	return &student, nil
}

// GetStudentByNIM retrieves a student by NIM
func (r *StudentRepository) GetStudentByNIM(nim string) (*models.Student, error) {
	var student models.Student
	if err := r.DB.Preload("User").Where("nim = ?", nim).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}
	return &student, nil
}

// UpdateStudent updates a student's information
func (r *StudentRepository) UpdateStudent(student *models.Student) error {
	return r.DB.Save(student).Error
}

// DeleteStudent deletes a student
func (r *StudentRepository) DeleteStudent(studentID uint) error {
	return r.DB.Delete(&models.Student{}, studentID).Error
}

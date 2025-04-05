package repository

import (
	"errors"
	"os"
	"time"

	"delpresence-api/internal/models"
	"delpresence-api/pkg/database"

	"github.com/golang-jwt/jwt/v5"
)

// AdminRepository menangani operasi terkait admin
type AdminRepository struct{}

// NewAdminRepository membuat instance AdminRepository baru
func NewAdminRepository() *AdminRepository {
	return &AdminRepository{}
}

// GetAdminByUserID mendapatkan admin berdasarkan ID user
func (r *AdminRepository) GetAdminByUserID(userID uint) (*models.AdminWithUser, error) {
	var admin models.Admin
	var user models.User

	// Cari admin berdasarkan UserID
	if err := database.DB.Where("user_id = ?", userID).First(&admin).Error; err != nil {
		return nil, err
	}

	// Cari user terkait
	if err := database.DB.Where("id = ?", admin.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	return &models.AdminWithUser{
		Admin: &admin,
		User:  &user,
	}, nil
}

// GetAdminByUsername mendapatkan admin berdasarkan username user
func (r *AdminRepository) GetAdminByUsername(username string) (*models.AdminWithUser, error) {
	var user models.User
	var admin models.Admin

	// Cari user berdasarkan username
	if err := database.DB.Where("username = ? AND user_type = ? AND active = ?",
		username, models.AdminType, true).First(&user).Error; err != nil {
		return nil, errors.New("admin tidak ditemukan atau tidak aktif")
	}

	// Cari admin profile berdasarkan user ID
	if err := database.DB.Where("user_id = ? AND is_active = ?",
		user.ID, true).First(&admin).Error; err != nil {
		return nil, errors.New("profil admin tidak ditemukan atau tidak aktif")
	}

	return &models.AdminWithUser{
		Admin: &admin,
		User:  &user,
	}, nil
}

// LoginAdmin menangani proses login admin
func (r *AdminRepository) LoginAdmin(username, password string, clientIP string) (*models.AdminLoginResponse, error) {
	// Dapatkan admin by username
	adminWithUser, err := r.GetAdminByUsername(username)
	if err != nil {
		return nil, err
	}

	user := adminWithUser.User
	admin := adminWithUser.Admin

	// Verifikasi password
	if !user.ComparePassword(password) {
		return nil, errors.New("password salah")
	}

	// Begin transaction
	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update last login time
	now := time.Now()
	user.LastLogin = &now

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update admin activity
	admin.LastActivity = &now
	admin.LoginCount += 1
	admin.IPAddress = clientIP

	if err := tx.Save(&admin).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Generate token JWT
	token, refreshToken, err := generateAdminTokens(*user, *admin)
	if err != nil {
		return nil, err
	}

	// Buat response
	adminUser := models.AdminAPIUser{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        "Admin",
		Status:      1,
		AccessLevel: string(admin.AccessLevel),
		Position:    admin.Position,
		Department:  admin.Department,
		Jabatan: []models.AdminJabatan{
			{
				StrukturJabatanID: 1,
				Jabatan:           admin.Position,
			},
		},
	}

	response := &models.AdminLoginResponse{
		Result:       true,
		Success:      "Login berhasil",
		User:         adminUser,
		Token:        token,
		RefreshToken: refreshToken,
	}

	return response, nil
}

// generateAdminTokens membuat token JWT untuk admin
func generateAdminTokens(user models.User, admin models.Admin) (string, string, error) {
	// Secret key dari environment variable
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secretKey) == 0 {
		// Fallback to default key if env var not set
		secretKey = []byte("your-secret-key-here")
	}

	// Ekspirasi token (8 jam)
	expirationTime := time.Now().Add(8 * time.Hour)

	// Buat claims (payload)
	claims := jwt.MapClaims{
		"uid":          user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"role":         "Admin",
		"admin_id":     admin.ID,
		"access_level": string(admin.AccessLevel),
		"exp":          expirationTime.Unix(),
		"iat":          time.Now().Unix(),
	}

	// Buat token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	// Buat refresh token (30 hari)
	refreshExpTime := time.Now().Add(30 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"uid":      user.ID,
		"admin_id": admin.ID,
		"exp":      refreshExpTime.Unix(),
		"iat":      time.Now().Unix(),
		"type":     "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, refreshTokenString, nil
}

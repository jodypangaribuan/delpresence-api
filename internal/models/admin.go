package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin represents the admin profile model in the database
type Admin struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Position     string         `gorm:"size:100;not null" json:"position"`
	Department   string         `gorm:"size:100" json:"department"`
	AccessLevel  AccessLevel    `gorm:"type:VARCHAR(20);not null;default:'standard'" json:"access_level"`
	LastActivity *time.Time     `json:"last_activity"`
	IPAddress    string         `gorm:"size:45" json:"ip_address"`
	LoginCount   int            `gorm:"default:0" json:"login_count"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// AccessLevel defines different levels of admin access
type AccessLevel string

const (
	// SuperAdminAccess has complete access to all features
	SuperAdminAccess AccessLevel = "super"
	// StandardAdminAccess has access to most admin features
	StandardAdminAccess AccessLevel = "standard"
	// LimitedAdminAccess has restricted access
	LimitedAdminAccess AccessLevel = "limited"
)

// AdminResponse represents the admin data returned in API responses
type AdminResponse struct {
	ID           uint        `json:"id"`
	UserID       uint        `json:"user_id"`
	Username     string      `json:"username"`
	Email        string      `json:"email"`
	FullName     string      `json:"full_name"`
	Position     string      `json:"position"`
	Department   string      `json:"department"`
	AccessLevel  AccessLevel `json:"access_level"`
	IsActive     bool        `json:"is_active"`
	LastLogin    *time.Time  `json:"last_login,omitempty"`
	LastActivity *time.Time  `json:"last_activity,omitempty"`
	LoginCount   int         `json:"login_count"`
}

// AdminWithUser represents the admin data with embedded user data
type AdminWithUser struct {
	Admin *Admin `json:"admin"`
	User  *User  `json:"user"`
}

// ToAdminResponse converts Admin with User to AdminResponse
func (a *Admin) ToAdminResponse(u *User) AdminResponse {
	return AdminResponse{
		ID:           a.ID,
		UserID:       a.UserID,
		Username:     u.Username,
		Email:        u.Email,
		FullName:     u.FullName(),
		Position:     a.Position,
		Department:   a.Department,
		AccessLevel:  a.AccessLevel,
		IsActive:     a.IsActive,
		LastLogin:    u.LastLogin,
		LastActivity: a.LastActivity,
		LoginCount:   a.LoginCount,
	}
}

// AdminLoginRequest adalah struktur untuk request login admin
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse adalah struktur untuk response login admin
type AdminLoginResponse struct {
	Result       bool         `json:"result"`
	Error        string       `json:"error"`
	Success      string       `json:"success"`
	User         AdminAPIUser `json:"user"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
}

// AdminAPIUser adalah struktur data user admin untuk API response
type AdminAPIUser struct {
	UserID      uint           `json:"user_id"`
	Username    string         `json:"username"`
	Email       string         `json:"email"`
	Role        string         `json:"role"`
	Status      int            `json:"status"`
	AccessLevel string         `json:"access_level"`
	Position    string         `json:"position"`
	Department  string         `json:"department"`
	Jabatan     []AdminJabatan `json:"jabatan"`
}

// AdminJabatan adalah struktur jabatan admin
type AdminJabatan struct {
	StrukturJabatanID int    `json:"struktur_jabatan_id"`
	Jabatan           string `json:"jabatan"`
}

package models

import (
	"time"

	"gorm.io/gorm"
)

// Assistant represents the details of a teaching assistant from campus API
type Assistant struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	AssistantUserID uint   `gorm:"unique;not null" json:"assistant_user_id"` // Local app user ID
	EmployeeID      uint   `json:"pegawai_id"`                               // From campus API - pegawai_id
	IdentityNumber  string `json:"nip"`                                      // From campus API - nip
	FullName        string `json:"nama"`                                     // From campus API - nama
	Email           string `json:"email"`                                    // From campus API - email
	Username        string `json:"user_name"`                                // From campus API - user_name
	CampusUserID    uint   `json:"user_id"`                                  // Campus UserID from API - user_id
	Alias           string `json:"alias"`                                    // From campus API - alias (with space)
	Position        string `json:"posisi"`                                   // From campus API - posisi (with space)
	EmployeeStatus  string `json:"status_pegawai"`                           // From campus API - status_pegawai (A,K,S,M,P,T)
	JobTitle        string `json:"jabatan"`                                  // From jabatan[].jabatan in login response
	StrukturID      uint   `json:"struktur_jabatan_id"`                      // From jabatan[].struktur_jabatan_id in login response
	Department      string `json:"department"`                               // Department name (usually faculty)

	// User customizable fields
	Avatar      string `json:"avatar"`       // Custom avatar uploaded by user
	Biography   string `json:"biography"`    // Customizable by user
	PhoneNumber string `json:"phone_number"` // Customizable by user
	Address     string `json:"address"`      // Customizable by user

	// System fields
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	LastSyncAt time.Time      `json:"last_sync_at"` // When assistant data was last synced from campus API
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// CampusAssistantResponse represents the response from the campus API for assistant details
type CampusAssistantResponse struct {
	Result string              `json:"result"`
	Data   CampusAssistantData `json:"data"`
}

// CampusAssistantData represents the data field in the campus API response
type CampusAssistantData struct {
	Pegawai []CampusAssistantDetail `json:"pegawai"`
}

// CampusAssistantDetail represents a single assistant detail from campus API
type CampusAssistantDetail struct {
	PegawaiID     uint   `json:"pegawai_id"`     // Maps to EmployeeID
	NIP           string `json:"nip"`            // Maps to IdentityNumber
	Nama          string `json:"nama"`           // Maps to FullName
	Email         string `json:"email"`          // Maps to Email
	UserName      string `json:"user_name"`      // Maps to Username
	UserID        uint   `json:"user_id"`        // Maps to CampusUserID
	Alias         string `json:"alias "`         // Maps to Alias (space in API response)
	Posisi        string `json:"posisi "`        // Maps to Position (space in API response)
	StatusPegawai string `json:"status_pegawai"` // Maps to EmployeeStatus (A,K,S,M,P,T)
}

// TableName sets the table name for the Assistant model
func (Assistant) TableName() string {
	return "assistants"
}

// AutoMigrateAssistant automatically creates and updates the assistant table
func AutoMigrateAssistant(db *gorm.DB) error {
	return db.AutoMigrate(&Assistant{})
}

// IsUserEditable checks if a field can be edited by the user
// Fields from the campus API should not be editable by the user
func (a *Assistant) IsUserEditable(fieldName string) bool {
	// List of fields that can be edited by the user
	editableFields := map[string]bool{
		"avatar":       true,
		"biography":    true,
		"phone_number": true,
		"address":      true,
	}

	// System fields and fields from campus API are not editable
	return editableFields[fieldName]
}

// GetEditableFields returns a map of all user-editable fields
func (a *Assistant) GetEditableFields() map[string]interface{} {
	return map[string]interface{}{
		"avatar":       a.Avatar,
		"biography":    a.Biography,
		"phone_number": a.PhoneNumber,
		"address":      a.Address,
	}
}

// GetReadOnlyFields returns a map of all read-only fields (from campus API)
func (a *Assistant) GetReadOnlyFields() map[string]interface{} {
	// Konversi status pegawai ke format yang lebih mudah dibaca
	employeeStatusText := "Tidak diketahui"
	switch a.EmployeeStatus {
	case "A":
		employeeStatusText = "Aktif"
	case "K":
		employeeStatusText = "Keluar/Tidak aktif lagi"
	case "S":
		employeeStatusText = "Studi/Sedang menempuh pendidikan lanjut"
	case "M":
		employeeStatusText = "Meninggal"
	case "P":
		employeeStatusText = "Pensiun"
	case "T":
		employeeStatusText = "Tidak diketahui/Tidak jelas"
	}

	return map[string]interface{}{
		"identity_number":      a.IdentityNumber,
		"full_name":            a.FullName,
		"email":                a.Email,
		"username":             a.Username,
		"department":           a.Department,
		"position":             a.Position,
		"job_title":            a.JobTitle,
		"struktur_id":          a.StrukturID,
		"employee_id":          a.EmployeeID,
		"employee_status":      a.EmployeeStatus,
		"employee_status_text": employeeStatusText,
		"alias":                a.Alias,
	}
}

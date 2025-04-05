package models

import (
	"time"

	"gorm.io/gorm"
)

// Lecturer represents the details of a lecturer from campus API
type Lecturer struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	LecturerUserID   uint   `gorm:"unique;not null" json:"lecturer_user_id"`
	EmployeeID       uint   `json:"pegawai_id"`            // From campus API
	LecturerID       uint   `json:"dosen_id"`              // From campus API
	IdentityNumber   string `json:"nip"`                   // From campus API
	FullName         string `json:"nama"`                  // From campus API
	Email            string `json:"email"`                 // From campus API
	DepartmentID     uint   `json:"prodi_id"`              // From campus API
	Department       string `json:"prodi"`                 // From campus API
	AcademicRank     string `json:"jabatan_akademik"`      // From campus API
	AcademicRankDesc string `json:"jabatan_akademik_desc"` // From campus API
	EducationLevel   string `json:"jenjang_pendidikan"`    // From campus API
	LecturerNumber   string `json:"nidn"`                  // From campus API
	CampusUserID     uint   `json:"user_id"`               // Campus UserID from API

	// User customizable fields
	Avatar       string `json:"avatar"`       // Custom avatar uploaded by user
	Biography    string `json:"biography"`    // Customizable by user
	Publications string `json:"publications"` // Customizable by user
	PhoneNumber  string `json:"phone_number"` // Customizable by user
	Address      string `json:"address"`      // Customizable by user

	// System fields
	Status     string         `json:"status"`       // Active, Inactive
	LastSyncAt time.Time      `json:"last_sync_at"` // When lecturer data was last synced from campus API
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// CampusLecturerResponse represents the response from the campus API for lecturer details
type CampusLecturerResponse struct {
	Result string             `json:"result"`
	Data   CampusLecturerData `json:"data"`
}

// CampusLecturerData represents the data field in the campus API response
type CampusLecturerData struct {
	Dosen []CampusLecturerDetail `json:"dosen"`
}

// CampusLecturerDetail represents a single lecturer detail from campus API
type CampusLecturerDetail struct {
	PegawaiID           uint   `json:"pegawai_id"`            // Maps to EmployeeID
	DosenID             uint   `json:"dosen_id"`              // Maps to AcademicID
	NIP                 string `json:"nip"`                   // Maps to IdentityNumber
	Nama                string `json:"nama"`                  // Maps to Name
	Email               string `json:"email"`                 // Maps to Email
	ProdiID             uint   `json:"prodi_id"`              // Maps to DepartmentID
	Prodi               string `json:"prodi"`                 // Maps to Department
	JabatanAkademik     string `json:"jabatan_akademik"`      // Maps to AcademicRankCode
	JabatanAkademikDesc string `json:"jabatan_akademik_desc"` // Maps to AcademicRank
	JenjangPendidikan   string `json:"jenjang_pendidikan"`    // Maps to EducationLevel
	NIDN                string `json:"nidn"`                  // Maps to LecturerIDNumber
	UserID              uint   `json:"user_id"`               // Maps to CampusIdentifierID
}

// TableName sets the table name for the Lecturer model
func (Lecturer) TableName() string {
	return "lecturers"
}

// GetProdiName returns the name of the prodi based on the ID
func GetProdiName(prodiID uint) string {
	prodiMap := map[uint]string{
		1:  "DIII Teknologi Informasi",
		2:  "DIII Manajemen Informatika",
		3:  "DIII Teknologi Komputer",
		4:  "Sarjana Terapan Teknologi Rekayasa Perangkat Lunak",
		6:  "S1 Informatika",
		7:  "S1 Teknik Elektro",
		8:  "S1 Teknik Bioproses",
		9:  "S1 Sistem Informasi",
		10: "S1 Manajemen Rekayasa",
		15: "S1 Teknik Metalurgi",
	}

	if name, ok := prodiMap[prodiID]; ok {
		return name
	}
	return "Unknown"
}

// GetJabatanDesc returns the description of the academic position based on the code
func GetJabatanDesc(code string) string {
	jabatanMap := map[string]string{
		"-": "- (No formal academic position)",
		"A": "Tenaga Pengajar (Teaching Staff)",
		"B": "Asisten Ahli (Assistant Expert/Assistant Professor)",
		"C": "Lektor (Lecturer)",
		"D": "Lektor Kepala (Senior Lecturer/Associate Professor)",
		"E": "Guru Besar (Professor)",
	}

	if desc, ok := jabatanMap[code]; ok {
		return desc
	}
	return "Unknown"
}

// AutoMigrateLecturer automatically creates and updates the lecturer table
func AutoMigrateLecturer(db *gorm.DB) error {
	return db.AutoMigrate(&Lecturer{})
}

// IsUserEditable checks if a field can be edited by the user
// Fields from the campus API should not be editable by the user
func (l *Lecturer) IsUserEditable(fieldName string) bool {
	// List of fields that can be edited by the user
	editableFields := map[string]bool{
		"avatar":       true,
		"biography":    true,
		"publications": true,
		"phone_number": true,
		"address":      true,
	}

	// System fields and fields from campus API are not editable
	return editableFields[fieldName]
}

// GetEditableFields returns a map of all user-editable fields
func (l *Lecturer) GetEditableFields() map[string]interface{} {
	return map[string]interface{}{
		"avatar":       l.Avatar,
		"biography":    l.Biography,
		"publications": l.Publications,
		"phone_number": l.PhoneNumber,
		"address":      l.Address,
	}
}

// GetReadOnlyFields returns a map of all read-only fields (from campus API)
func (l *Lecturer) GetReadOnlyFields() map[string]interface{} {
	return map[string]interface{}{
		"identity_number":    l.IdentityNumber,
		"lecturer_number":    l.LecturerNumber,
		"full_name":          l.FullName,
		"email":              l.Email,
		"department":         l.Department,
		"academic_rank":      l.AcademicRank,
		"academic_rank_desc": l.AcademicRankDesc,
		"education_level":    l.EducationLevel,
		"status":             l.Status,
	}
}

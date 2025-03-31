package models

// CampusAuthResponse represents the authentication response from the campus API
type CampusAuthResponse struct {
	Result       bool       `json:"result"`
	Error        string     `json:"error"`
	Success      string     `json:"success"`
	User         CampusUser `json:"user"`
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
}

// CampusUser represents the user information in the authentication response
type CampusUser struct {
	UserID   int         `json:"user_id"`
	Username string      `json:"username"`
	Email    string      `json:"email"`
	Role     string      `json:"role"`
	Status   int         `json:"status"`
	Jabatan  interface{} `json:"jabatan"`
}

// MahasiswaListResponse represents the response from the get-mahasiswa endpoint
type MahasiswaListResponse struct {
	Result string `json:"result"`
	Data   struct {
		Mahasiswa []MahasiswaInfo `json:"mahasiswa"`
	} `json:"data"`
}

// MahasiswaInfo represents a student's information from the mahasiswa list
type MahasiswaInfo struct {
	DimID     int    `json:"dim_id"`
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	Nim       string `json:"nim"`
	Nama      string `json:"nama"`
	Email     string `json:"email"`
	ProdiID   int    `json:"prodi_id"`
	ProdiName string `json:"prodi_name"`
	Fakultas  string `json:"fakultas"`
	Angkatan  int    `json:"angkatan"`
	Status    string `json:"status"`
	Asrama    string `json:"asrama"`
}

// MahasiswaDetailResponse represents the response from the get-student-by-nim endpoint
type MahasiswaDetailResponse struct {
	Result string          `json:"result"`
	Data   MahasiswaDetail `json:"data"`
}

// MahasiswaDetail represents a student's detailed information
type MahasiswaDetail struct {
	Nim          string `json:"nim"`
	Nama         string `json:"nama"`
	Email        string `json:"email"`
	TempatLahir  string `json:"tempat_lahir"`
	TglLahir     string `json:"tgl_lahir"`
	JenisKelamin string `json:"jenis_kelamin"`
	Alamat       string `json:"alamat"`
	Hp           string `json:"hp"`
	Prodi        string `json:"prodi"`
	Fakultas     string `json:"fakultas"`
	Sem          int    `json:"sem"`
	SemTa        int    `json:"sem_ta"`
	Ta           string `json:"ta"`
	TahunMasuk   int    `json:"tahun_masuk"`
	Kelas        string `json:"kelas"`
	DosenWali    string `json:"dosen_wali"`
	Asrama       string `json:"asrama"`
	NamaAyah     string `json:"nama_ayah"`
	NamaIbu      string `json:"nama_ibu"`
	NoHpAyah     string `json:"no_hp_ayah"`
	NoHpIbu      string `json:"no_hp_ibu"`
}

// MahasiswaComplete combines basic student information with detailed information
type MahasiswaComplete struct {
	BasicInfo MahasiswaInfo   `json:"basic_info"`
	Details   MahasiswaDetail `json:"details"`
}

# DelPresence API

API backend dengan JWT authentication untuk sistem login dan register aplikasi DelPresence.

## Teknologi

- Go (Golang)
- Gin Web Framework
- GORM (ORM)
- MySQL
- JWT Authentication

## Prasyarat

- Go 1.21 atau versi lebih tinggi
- MySQL

## Cara Instalasi

1. Clone repository ini

   ```bash
   git clone https://github.com/username/delpresence-api.git
   cd delpresence-api
   ```

2. Instal dependensi

   ```bash
   go mod tidy
   ```

3. Konfigurasi database

   - Buat database MySQL dengan nama `delpresence`
   - Sesuaikan konfigurasi database pada file `.env`

4. Jalankan aplikasi
   ```bash
   go run cmd/api/main.go
   ```

## Struktur Proyek

```
delpresence-api/
├── cmd/                # Entry points aplikasi
│   └── api/            # API server
├── internal/           # Private application code
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Middleware components
│   ├── models/         # Data models
│   ├── repository/     # Database operations
│   └── utils/          # Utility functions
├── pkg/                # Public libraries
│   ├── database/       # Database connection
│   └── jwt/            # JWT utilities
├── .env                # Environment configuration
├── go.mod              # Go modules
└── README.md           # Documentation
```

## Endpoints API

### Auth

- **POST /api/v1/auth/register**

  - Deskripsi: Mendaftarkan pengguna baru
  - Body:
    ```json
    {
      "nim_nip": "string",
      "name": "string",
      "email": "string",
      "password": "string",
      "user_type": "student / staff",
      "major": "string (optional)",
      "faculty": "string (optional)",
      "position": "string (optional)"
    }
    ```

- **POST /api/v1/auth/login**

  - Deskripsi: Login pengguna
  - Body:
    ```json
    {
      "nim_nip": "string",
      "password": "string"
    }
    ```
  - Response:
    ```json
    {
      "success": true,
      "message": "Login successful",
      "data": {
        "user": {
          "id": 1,
          "nim_nip": "string",
          "name": "string",
          "email": "string",
          "user_type": "student / staff",
          "major": "string (if exists)",
          "faculty": "string (if exists)",
          "position": "string (if exists)",
          "verified": false
        },
        "tokens": {
          "access_token": "string",
          "refresh_token": "string",
          "expires_in": 86400
        }
      }
    }
    ```

- **POST /api/v1/auth/refresh**

  - Deskripsi: Memperbaharui access token
  - Body:
    ```json
    {
      "refresh_token": "string"
    }
    ```
  - Response:
    ```json
    {
      "success": true,
      "message": "Token refreshed successfully",
      "data": {
        "access_token": "string",
        "expires_in": 86400
      }
    }
    ```

- **POST /api/v1/auth/logout**

  - Deskripsi: Logout pengguna (invalidate token)
  - Body:
    ```json
    {
      "refresh_token": "string"
    }
    ```

- **GET /api/v1/auth/me**
  - Deskripsi: Mendapatkan data pengguna yang sedang login
  - Headers:
    ```
    Authorization: Bearer {access_token}
    ```
  - Response:
    ```json
    {
      "success": true,
      "message": "User details retrieved successfully",
      "data": {
        "id": 1,
        "nim_nip": "string",
        "name": "string",
        "email": "string",
        "user_type": "student / staff",
        "major": "string (if exists)",
        "faculty": "string (if exists)",
        "position": "string (if exists)",
        "verified": false
      }
    }
    ```

## Autentikasi

API ini menggunakan JWT (JSON Web Tokens) untuk autentikasi. Token akses harus disertakan pada header Authorization untuk endpoint yang memerlukan autentikasi.

Format: `Authorization: Bearer {access_token}`

## Penanganan Error

API ini menggunakan format error yang konsisten:

```json
{
  "success": false,
  "message": "Error message",
  "error": "Error details (if available)"
}
```

## Pengembangan dan Kontribusi

1. Fork repository
2. Buat branch baru: `git checkout -b fitur-baru`
3. Commit perubahan: `git commit -am 'Menambahkan fitur baru'`
4. Push ke branch: `git push origin fitur-baru`
5. Submit pull request

## Lisensi

MIT

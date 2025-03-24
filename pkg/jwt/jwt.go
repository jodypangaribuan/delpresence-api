package jwt

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Custom error definitions
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// CustomClaims defines the claims for JWT
type CustomClaims struct {
	UserID     uint   `json:"user_id"`
	NimNip     string `json:"nim_nip"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new JWT access token
func GenerateAccessToken(userID uint, nimNip string, firstName string, middleName string, lastName string, email string) (string, time.Time, error) {
	// Get secret key from environment
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", time.Time{}, errors.New("JWT_SECRET environment variable not set")
	}

	// Parse expiry duration
	expiryStr := os.Getenv("JWT_EXPIRY")
	if expiryStr == "" {
		expiryStr = "24h" // Default to 24 hours
	}

	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid JWT_EXPIRY format: %v", err)
	}

	expiryTime := time.Now().Add(expiry)

	// Create the Claims
	claims := CustomClaims{
		UserID:     userID,
		NimNip:     nimNip,
		FirstName:  firstName,
		MiddleName: middleName,
		LastName:   lastName,
		Email:      email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "delpresence-api",
			Subject:   strconv.Itoa(int(userID)),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryTime, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*CustomClaims, error) {
	// Get secret key from environment
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET environment variable not set")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		// Check if the error is because the token is expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

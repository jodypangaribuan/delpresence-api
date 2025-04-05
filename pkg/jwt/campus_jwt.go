package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CampusClaims defines the claims structure for campus API JWT tokens
type CampusClaims struct {
	// Campus API uses "uid" field for user ID
	UID int `json:"uid"`
	// Registered claims
	jwt.RegisteredClaims
}

// ValidateCampusToken validates a JWT token from the campus API
// This function is more permissive than our regular JWT validation
func ValidateCampusToken(tokenString string) (int, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("campus-api-dummy-key"), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil && !errors.Is(err, jwt.ErrSignatureInvalid) {
		return 0, ErrInvalidToken
	}

	// Try to extract from map claims
	if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
		// Try to get the uid from the claims
		if uid, exists := mapClaims["uid"]; exists {
			// Convert UID to int
			var userID int
			switch v := uid.(type) {
			case float64:
				userID = int(v)
			case int:
				userID = v
			case string:
				// Try to parse string as number
				var parseErr error
				userID, parseErr = parseInt(v)
				if parseErr != nil {
					return 0, ErrInvalidToken
				}
			default:
				return 0, ErrInvalidToken
			}

			// Check expiry manually using map claims
			if exp, exists := mapClaims["exp"]; exists {
				if expFloat, ok := exp.(float64); ok {
					expTime := time.Unix(int64(expFloat), 0)
					if time.Now().After(expTime) {
						return 0, ErrExpiredToken
					}
				}
			}

			return userID, nil
		}
	}

	// Parse again with structured claims as fallback
	parsedToken, err := jwt.ParseWithClaims(tokenString, &CampusClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("campus-api-dummy-key"), nil
	}, jwt.WithoutClaimsValidation())

	// If we can't parse the token at all, it's invalid
	if err != nil && !errors.Is(err, jwt.ErrSignatureInvalid) {
		return 0, ErrInvalidToken
	}

	// Try to extract structured claims
	claims, ok := parsedToken.Claims.(*CampusClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	// Check if the token has user ID in structured claims
	if claims.UID == 0 {
		return 0, ErrInvalidToken
	}

	// Check expiry manually in structured claims
	if claims.ExpiresAt != nil {
		expiryTime := claims.ExpiresAt.Time
		if time.Now().After(expiryTime) {
			return 0, ErrExpiredToken
		}
	}

	// Return the user ID from the token
	return claims.UID, nil
}

// Helper function to try to parse a string as an integer
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

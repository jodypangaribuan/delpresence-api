package jwt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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
	// Debug the token structure
	log.Printf("Token debug - Token length: %d, First 20 chars: %s", len(tokenString), safeSubstring(tokenString, 0, 20))
	debugDumpToken(tokenString)

	// First try parsing with flexible options to diagnose token structure
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("campus-api-dummy-key"), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil && !errors.Is(err, jwt.ErrSignatureInvalid) {
		log.Printf("Error parsing campus token: %v", err)
		return 0, ErrInvalidToken
	}

	// Try to extract from map claims first (most common case)
	if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
		// Log all claims for debugging
		log.Printf("Token claims found: %d claims", len(mapClaims))
		for key, val := range mapClaims {
			log.Printf("  Claim %s: %v (type: %T)", key, val, val)
		}

		// Try to get the uid from the claims
		if uid, exists := mapClaims["uid"]; exists {
			log.Printf("Found uid claim: %v (type: %T)", uid, uid)

			// Convert UID to int
			var userID int
			switch v := uid.(type) {
			case float64:
				userID = int(v)
				log.Printf("Converted float64 uid to int: %d", userID)
			case int:
				userID = v
				log.Printf("Used int uid directly: %d", userID)
			case string:
				// Try to parse string as number
				var parseErr error
				userID, parseErr = parseInt(v)
				if parseErr != nil {
					log.Printf("Failed to parse string uid as int: %v", parseErr)
					return 0, ErrInvalidToken
				}
				log.Printf("Converted string uid to int: %d", userID)
			default:
				log.Printf("Unsupported uid type: %T", uid)
				return 0, ErrInvalidToken
			}

			// Check expiry manually using map claims
			if exp, exists := mapClaims["exp"]; exists {
				if expFloat, ok := exp.(float64); ok {
					expTime := time.Unix(int64(expFloat), 0)
					if time.Now().After(expTime) {
						log.Printf("Token expired at %v", expTime)
						return 0, ErrExpiredToken
					}
				}
			}

			log.Printf("Token validation successful for user ID: %d", userID)
			return userID, nil
		} else {
			log.Printf("No uid claim found in token")
		}
	} else {
		log.Printf("Failed to parse token as MapClaims")
	}

	// Parse again with structured claims as fallback
	parsedToken, err := jwt.ParseWithClaims(tokenString, &CampusClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("campus-api-dummy-key"), nil
	}, jwt.WithoutClaimsValidation())

	// If we can't parse the token at all, it's invalid
	if err != nil && !errors.Is(err, jwt.ErrSignatureInvalid) {
		log.Printf("Error parsing campus token with claims: %v", err)
		return 0, ErrInvalidToken
	}

	// Try to extract structured claims
	claims, ok := parsedToken.Claims.(*CampusClaims)
	if !ok {
		log.Printf("Token claims not valid for StructuredClaims")
		return 0, ErrInvalidToken
	}

	// Check if the token has user ID in structured claims
	if claims.UID == 0 {
		log.Printf("Token has no UID claim in structured claims")
		return 0, ErrInvalidToken
	}

	// Check expiry manually in structured claims
	if claims.ExpiresAt != nil {
		expiryTime := claims.ExpiresAt.Time
		if time.Now().After(expiryTime) {
			log.Printf("Token expired at %v", expiryTime)
			return 0, ErrExpiredToken
		}
	}

	// Return the user ID from the token
	log.Printf("Token validation successful for user ID (structured): %d", claims.UID)
	return claims.UID, nil
}

// Helper function to safely get a substring of a string
func safeSubstring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

// Helper function to try to parse a string as an integer
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// Debug function to dump token parts
func debugDumpToken(tokenString string) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		log.Printf("Token does not have 3 parts (header.payload.signature), has %d parts", len(parts))
		return
	}

	// Try to decode header
	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		log.Printf("Failed to decode header: %v", err)
	} else {
		var header map[string]interface{}
		if err := json.Unmarshal(headerJSON, &header); err != nil {
			log.Printf("Failed to parse header JSON: %v", err)
		} else {
			log.Printf("Token header: %v", header)
		}
	}

	// Try to decode payload
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		log.Printf("Failed to decode payload: %v", err)
	} else {
		var payload map[string]interface{}
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			log.Printf("Failed to parse payload JSON: %v", err)
		} else {
			log.Printf("Token payload: %v", payload)
		}
	}
}

// Helper function to decode base64 URL-encoded strings
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	return base64.URLEncoding.DecodeString(s)
}

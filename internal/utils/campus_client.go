package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"delpresence-api/internal/models"
)

const (
	campusAPIBaseURL = "https://cis.del.ac.id/api"
	campusAuthURL    = "https://cis-dev.del.ac.id/api/jwt-api/do-auth"
	defaultUsername  = "johannes"
	defaultPassword  = "Del@2022"
)

// TokenCache stores the authentication tokens
type TokenCache struct {
	AuthToken    string
	RefreshToken string
	ExpiresAt    time.Time
	mutex        sync.RWMutex
}

// CampusClient is a client for interacting with the campus API
type CampusClient struct {
	httpClient *http.Client
	tokenCache *TokenCache
}

// AuthRoundTripper is a custom RoundTripper that adds authentication headers to requests
type AuthRoundTripper struct {
	BaseTransport http.RoundTripper
	TokenCache    *TokenCache
}

// RoundTrip implements the http.RoundTripper interface
func (rt *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("[TOKEN_DEBUG] Processing request to: %s", req.URL.String())

	// Skip token check for authentication requests
	if req.URL.String() == campusAuthURL {
		log.Printf("[TOKEN_DEBUG] Direct auth request to: %s", campusAuthURL)
		return rt.BaseTransport.RoundTrip(req)
	}

	// Check if we need to refresh token
	rt.TokenCache.mutex.RLock()
	token := rt.TokenCache.AuthToken
	refreshToken := rt.TokenCache.RefreshToken
	expiresAt := rt.TokenCache.ExpiresAt
	rt.TokenCache.mutex.RUnlock()

	// Get a new token if needed (none exists or is about to expire)
	tokenIsExpiredOrMissing := token == "" || time.Now().Add(30*time.Second).After(expiresAt)

	if tokenIsExpiredOrMissing {
		log.Printf("[TOKEN_DEBUG] Token is missing or about to expire. Current token: %s... Expiry: %v",
			safeSubstring(token, 0, 10), expiresAt)

		// Try to use refresh token if available
		if refreshToken != "" {
			// TODO: Implement refresh token flow if campus API supports it
			log.Println("[TOKEN_DEBUG] Refresh token available but refresh flow not implemented, falling back to new auth")
		}

		// Get a new token with full authentication
		newToken, newRefreshToken, expiryTime, err := getNewToken()
		if err != nil {
			log.Printf("[TOKEN_DEBUG] Failed to get authentication token: %v", err)
			return nil, fmt.Errorf("failed to get authentication token: %w", err)
		}

		// Update token cache
		rt.TokenCache.mutex.Lock()
		rt.TokenCache.AuthToken = newToken
		rt.TokenCache.RefreshToken = newRefreshToken
		rt.TokenCache.ExpiresAt = expiryTime
		rt.TokenCache.mutex.Unlock()

		token = newToken
		log.Printf("[TOKEN_DEBUG] Successfully obtained new token, expires at: %v", expiryTime)
	} else {
		log.Printf("[TOKEN_DEBUG] Using existing token, expires at: %v", expiresAt)
	}

	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+token)
	log.Printf("[TOKEN_DEBUG] Request to %s with token (first 15 chars): %s...",
		reqClone.URL.String(),
		safeSubstring(token, 0, 15))

	// Send the request with the token
	resp, err := rt.BaseTransport.RoundTrip(reqClone)
	if err != nil {
		log.Printf("[TOKEN_DEBUG] Campus API request failed: %v", err)
		return nil, err
	}

	log.Printf("[TOKEN_DEBUG] Response from %s: %d", reqClone.URL.String(), resp.StatusCode)

	// If we get a 401 Unauthorized, our token might be expired
	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("[TOKEN_DEBUG] Got 401 Unauthorized, token might be expired")

		// Close the current response body
		resp.Body.Close()

		// Force get a new token
		newToken, newRefreshToken, expiryTime, err := getNewToken()
		if err != nil {
			log.Printf("[TOKEN_DEBUG] Failed to refresh authentication token: %v", err)
			return nil, fmt.Errorf("failed to refresh authentication token: %w", err)
		}

		// Update token cache
		rt.TokenCache.mutex.Lock()
		rt.TokenCache.AuthToken = newToken
		rt.TokenCache.RefreshToken = newRefreshToken
		rt.TokenCache.ExpiresAt = expiryTime
		rt.TokenCache.mutex.Unlock()

		// Create a new request with the new token
		reqClone = req.Clone(req.Context())
		reqClone.Header.Set("Authorization", "Bearer "+newToken)
		log.Printf("[TOKEN_DEBUG] Retrying request with new token (first 15 chars): %s...", safeSubstring(newToken, 0, 15))

		// Retry the request with the new token
		return rt.BaseTransport.RoundTrip(reqClone)
	}

	return resp, nil
}

// getNewToken authenticates and gets a new token from the campus API
// Returns token, refresh token, expiry time, and error
func getNewToken() (string, string, time.Time, error) {
	log.Printf("Authenticating with campus API using account: %s", defaultUsername)

	// Create a multipart form data request (matching Flutter's http.MultipartRequest)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	if err := writer.WriteField("username", defaultUsername); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to add username field: %w", err)
	}
	if err := writer.WriteField("password", defaultPassword); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to add password field: %w", err)
	}

	// Close writer to finalize the form
	if err := writer.Close(); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", campusAuthURL, body)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type and other headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "*/*")

	// Create client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Log request info
	log.Printf("Sending auth request to %s", campusAuthURL)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response info
	log.Printf("Auth response status: %d", resp.StatusCode)
	log.Printf("Auth response body: %s", string(respBody))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", "", time.Time{}, fmt.Errorf("auth failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var authResp models.CampusAuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result
	if !authResp.Result {
		return "", "", time.Time{}, fmt.Errorf("authentication failed: %s", authResp.Error)
	}

	// Validate token
	if authResp.Token == "" {
		return "", "", time.Time{}, fmt.Errorf("empty token received")
	}

	// Extract expiry time from token
	expiryTime := extractExpiryFromToken(authResp.Token)
	log.Printf("Got new token with expiry: %v", expiryTime)

	return authResp.Token, authResp.RefreshToken, expiryTime, nil
}

// extractExpiryFromToken tries to extract the expiry time from a JWT token
func extractExpiryFromToken(tokenString string) time.Time {
	// Default expiry time (30 minutes from now) in case we can't extract it
	defaultExpiry := time.Now().Add(30 * time.Minute)

	// Split the token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return defaultExpiry
	}

	// Try to decode the payload
	payload, err := decodeTokenPart(parts[1])
	if err != nil {
		return defaultExpiry
	}

	// Try to parse the expiry claim
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return defaultExpiry
	}

	// Check for the "exp" claim
	if expValue, ok := claims["exp"]; ok {
		// Try to convert to a number
		if expFloat, ok := expValue.(float64); ok {
			return time.Unix(int64(expFloat), 0)
		}
	}

	return defaultExpiry
}

// decodeTokenPart decodes a base64url encoded token part
func decodeTokenPart(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	return base64.URLEncoding.DecodeString(s)
}

// safeSubstring returns a substring of s, handling bounds safely
func safeSubstring(s string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

// NewCampusClient creates a new client for the campus API
func NewCampusClient() *CampusClient {
	tokenCache := &TokenCache{
		mutex: sync.RWMutex{},
	}

	transport := &AuthRoundTripper{
		BaseTransport: http.DefaultTransport,
		TokenCache:    tokenCache,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Pre-fetch a token asynchronously
	go func() {
		token, refreshToken, expiresAt, err := getNewToken()
		if err != nil {
			log.Printf("Initial token fetch failed: %v", err)
			return
		}

		tokenCache.mutex.Lock()
		tokenCache.AuthToken = token
		tokenCache.RefreshToken = refreshToken
		tokenCache.ExpiresAt = expiresAt
		tokenCache.mutex.Unlock()

		log.Printf("Initial token pre-fetched successfully")
	}()

	return &CampusClient{
		httpClient: httpClient,
		tokenCache: tokenCache,
	}
}

// GetMahasiswaByUserID fetches student information by user ID
func (c *CampusClient) GetMahasiswaByUserID(userID int) (*models.MahasiswaInfo, error) {
	url := fmt.Sprintf("%s/library-api/mahasiswa?userid=%d", campusAPIBaseURL, userID)
	log.Printf("Fetching student info for user ID: %d from URL: %s", userID, url)

	// Send the request
	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("Error fetching student info: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log a summary of the response
	respSummary := safeSubstring(string(body), 0, 100)
	log.Printf("Student info response (first 100 chars): %s...", respSummary)

	// Parse response
	var mahasiswaResp models.MahasiswaListResponse
	if err := json.Unmarshal(body, &mahasiswaResp); err != nil {
		log.Printf("Error parsing student info response: %v", err)
		return nil, err
	}

	// Check if response is valid
	if mahasiswaResp.Result != "Ok" {
		log.Printf("Campus API returned non-Ok result for user ID %d: %s", userID, mahasiswaResp.Result)
		return nil, fmt.Errorf("API returned non-Ok result: %s", mahasiswaResp.Result)
	}

	// Check if any mahasiswa data was returned
	if len(mahasiswaResp.Data.Mahasiswa) == 0 {
		log.Printf("No student found with user ID: %d", userID)
		return nil, fmt.Errorf("no student found with user ID: %d", userID)
	}

	log.Printf("Found student: %s (NIM: %s)",
		mahasiswaResp.Data.Mahasiswa[0].Nama,
		mahasiswaResp.Data.Mahasiswa[0].Nim)
	return &mahasiswaResp.Data.Mahasiswa[0], nil
}

// GetMahasiswaDetailByNIM fetches detailed student information by NIM
func (c *CampusClient) GetMahasiswaDetailByNIM(nim string) (*models.MahasiswaDetail, error) {
	url := fmt.Sprintf("%s/library-api/get-student-by-nim?nim=%s", campusAPIBaseURL, nim)
	log.Printf("Fetching student details for NIM: %s from URL: %s", nim, url)

	// Send the request
	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("Error fetching student details: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading student details response: %v", err)
		return nil, err
	}

	// Log a summary of the response
	respSummary := safeSubstring(string(body), 0, 100)
	log.Printf("Student details response for NIM %s (first 100 chars): %s...", nim, respSummary)

	// Parse response
	var detailResp models.MahasiswaDetailResponse
	if err := json.Unmarshal(body, &detailResp); err != nil {
		log.Printf("Error parsing student details response: %v", err)
		return nil, err
	}

	// Check if response is valid
	if detailResp.Result != "OK" {
		log.Printf("Campus API returned non-OK result for NIM %s: %s", nim, detailResp.Result)
		return nil, fmt.Errorf("failed to get student details for NIM: %s", nim)
	}

	log.Printf("Successfully retrieved details for student with NIM: %s, Name: %s",
		nim, detailResp.Data.Nama)
	return &detailResp.Data, nil
}

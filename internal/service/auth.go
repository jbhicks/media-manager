package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	return []byte("media-manager-secret-key-change-in-production")
}

// IsPublicAPIPath returns true for API routes that do not require authentication.
func IsPublicAPIPath(path string) bool {
	switch path {
	case "/api/auth/login", "/api/health":
		return true
	case "/api/auth/register":
		return os.Getenv("ALLOW_REGISTRATION") == "true"
	default:
		return false
	}
}

// APIAuthMiddleware protects all /api/* routes except public endpoints.
func APIAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") || IsPublicAPIPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})(w, r)
	})
}

// AuthMiddleware validates JWT tokens and injects user into request context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			jsonError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			jsonError(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret(), nil
		})

		if err != nil || !token.Valid {
			jsonError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			jsonError(w, "Token expired", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		r.Header.Set("X-User-Name", claims.Username)
		r.Header.Set("X-User-Role", claims.Role)

		next(w, r)
	}
}

// OptionalAuthMiddleware allows requests through but adds user info if token is present
func OptionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				claims := &Claims{}
				token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret(), nil
				})
				if err == nil && token.Valid && claims.ExpiresAt != nil && !claims.ExpiresAt.Time.Before(time.Now()) {
					r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
					r.Header.Set("X-User-Name", claims.Username)
					r.Header.Set("X-User-Role", claims.Role)
				}
			}
		}
		next(w, r)
	}
}

// GetUserIDFromRequest extracts user ID from request context
func GetUserIDFromRequest(r *http.Request) uint {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0
	}
	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)
	return userID
}

// GetUserRoleFromRequest extracts user role from request context
func GetUserRoleFromRequest(r *http.Request) string {
	return r.Header.Get("X-User-Role")
}

// HashPassword hashes a password using SHA-256
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

// AuthHandler handles login and registration
type AuthHandler struct {
	db *db.Database
}

func NewAuthHandler(db *db.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/register", h.handleRegister)
	mux.HandleFunc("/api/auth/login", h.handleLogin)
	mux.HandleFunc("/api/auth/me", AuthMiddleware(h.handleMe))
	mux.HandleFunc("/api/auth/logout", AuthMiddleware(h.handleLogout))
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("ALLOW_REGISTRATION") != "true" {
		jsonError(w, "Registration is disabled", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	var existingUser models.User
	if err := h.db.GetDB().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		jsonError(w, "Username already exists", http.StatusConflict)
		return
	}

	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: HashPassword(req.Password),
		Role:         "user",
		IsActive:     true,
	}

	if err := h.db.GetDB().Create(&user).Error; err != nil {
		jsonError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(&user)
	if err != nil {
		jsonError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"token":   token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := h.db.GetDB().Where("username = ?", req.Username).First(&user).Error; err != nil {
		jsonError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if user.PasswordHash != HashPassword(req.Password) {
		jsonError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		jsonError(w, "Account is disabled", http.StatusForbidden)
		return
	}

	now := time.Now()
	user.LastLoginAt = &now
	h.db.GetDB().Save(&user)

	token, err := GenerateToken(&user)
	if err != nil {
		jsonError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := h.db.GetDB().First(&user, userID).Error; err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"role":         user.Role,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarURL,
	})
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}
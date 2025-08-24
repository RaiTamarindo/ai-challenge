package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents JWT claims structure
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// TokenService defines the interface for JWT operations
type TokenService interface {
	GenerateToken(userID int, username, email string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

// PasswordService defines the interface for password operations
type PasswordService interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

// JWTService implements TokenService using JWT
type JWTService struct {
	secret string
}

// NewJWTService creates a new JWT service
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: secret,
	}
}

// GenerateToken generates a new JWT token
func (s *JWTService) GenerateToken(userID int, username, email string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ValidateToken validates a JWT token and returns claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// BCryptPasswordService implements PasswordService using bcrypt
type BCryptPasswordService struct{}

// NewBCryptPasswordService creates a new bcrypt password service
func NewBCryptPasswordService() *BCryptPasswordService {
	return &BCryptPasswordService{}
}

// HashPassword hashes a password using bcrypt
func (s *BCryptPasswordService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash checks if password matches the hash
func (s *BCryptPasswordService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
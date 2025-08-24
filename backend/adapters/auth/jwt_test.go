package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		userID   int
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "valid token generation",
			secret:   "test-secret",
			userID:   123,
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
		{
			name:     "empty secret",
			secret:   "",
			userID:   123,
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false, // JWT library allows empty secrets
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewJWTService(tt.secret)
			
			token, err := service.GenerateToken(tt.userID, tt.username, tt.email)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
				return
			}
			
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			
			// Verify token can be parsed
			claims, err := service.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, tt.email, claims.Email)
			assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
			assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	secret := "test-secret"
	service := NewJWTService(secret)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		setupFunc func() string
	}{
		{
			name:    "valid token",
			wantErr: false,
			setupFunc: func() string {
				token, _ := service.GenerateToken(123, "testuser", "test@example.com")
				return token
			},
		},
		{
			name:    "invalid token format",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "token with wrong secret",
			wantErr: true,
			setupFunc: func() string {
				wrongService := NewJWTService("wrong-secret")
				token, _ := wrongService.GenerateToken(123, "testuser", "test@example.com")
				return token
			},
		},
		{
			name:    "expired token",
			wantErr: true,
			setupFunc: func() string {
				// Create an expired token
				claims := &Claims{
					UserID:   123,
					Username: "testuser",
					Email:    "test@example.com",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.token
			if tt.setupFunc != nil {
				token = tt.setupFunc()
			}

			claims, err := service.ValidateToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, 123, claims.UserID)
			assert.Equal(t, "testuser", claims.Username)
			assert.Equal(t, "test@example.com", claims.Email)
		})
	}
}

func TestJWTService_ValidateToken_WrongSigningMethod(t *testing.T) {
	service := NewJWTService("test-secret")

	// Create token with wrong signing method
	claims := &Claims{
		UserID:   123,
		Username: "testuser",
		Email:    "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Use RS256 instead of HS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	// This will fail because RS256 requires a different key format
	assert.Error(t, err)

	// If we could create the token, validation should fail
	if tokenString != "" {
		claims, err := service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	}
}

func TestBCryptPasswordService_HashPassword(t *testing.T) {
	service := NewBCryptPasswordService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "testpassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "long password",
			password: string(make([]byte, 100)), // 100 bytes of null characters
			wantErr:  true, // bcrypt has a 72 byte limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash) // Hash should be different from password
			assert.True(t, len(hash) > 50) // bcrypt hashes are typically 60 characters

			// Verify the hash can be used to check the password
			assert.True(t, service.CheckPasswordHash(tt.password, hash))
		})
	}
}

func TestBCryptPasswordService_CheckPasswordHash(t *testing.T) {
	service := NewBCryptPasswordService()
	password := "testpassword123"
	hash, err := service.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password with valid hash",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: password,
			hash:     "invalid-hash",
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBCryptPasswordService_Integration(t *testing.T) {
	service := NewBCryptPasswordService()

	// Test multiple password/hash cycles
	passwords := []string{
		"simple",
		"complex!@#$%^&*()_+",
		"unicode测试密码",
		"very-long-password-with-many-characters-to-test-boundaries",
	}

	for _, password := range passwords {
		t.Run("password_"+password, func(t *testing.T) {
			// Hash the password
			hash1, err := service.HashPassword(password)
			require.NoError(t, err)

			// Hash the same password again - should get different hash
			hash2, err := service.HashPassword(password)
			require.NoError(t, err)
			assert.NotEqual(t, hash1, hash2, "Hashing the same password twice should produce different hashes")

			// Both hashes should validate the original password
			assert.True(t, service.CheckPasswordHash(password, hash1))
			assert.True(t, service.CheckPasswordHash(password, hash2))

			// Wrong passwords should not validate
			assert.False(t, service.CheckPasswordHash(password+"wrong", hash1))
			assert.False(t, service.CheckPasswordHash(password+"wrong", hash2))
		})
	}
}
package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	authmocks "github.com/feature-voting-platform/backend/adapters/auth/mocks"
	logsmocks "github.com/feature-voting-platform/backend/adapters/logs/mocks"
	"github.com/feature-voting-platform/backend/domain/users"
	usersmocks "github.com/feature-voting-platform/backend/domain/users/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Note: Register functionality would be tested here if implemented
// Currently only Login and GetProfile are implemented in auth_handler.go

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*usersmocks.MockRepository, *authmocks.MockTokenService, *authmocks.MockPasswordService, *logsmocks.MockLogger)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful login with email",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMocks: func(userRepo *usersmocks.MockRepository, tokenService *authmocks.MockTokenService, passwordService *authmocks.MockPasswordService, logger *logsmocks.MockLogger) {
				user := &users.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
				}
				userRepo.On("GetByEmail", "test@example.com").Return(user, nil)
				passwordService.On("CheckPasswordHash", "password123", "hashed_password").Return(true)
				tokenService.On("GenerateToken", 1, "testuser", "test@example.com").Return("jwt_token", nil)
				logger.On("Info", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "jwt_token", response["token"])
				user := response["user"].(map[string]interface{})
				assert.Equal(t, float64(1), user["id"])
				assert.Equal(t, "testuser", user["username"])
				assert.Equal(t, "test@example.com", user["email"])
			},
		},
		{
			name: "successful login with username",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			setupMocks: func(userRepo *usersmocks.MockRepository, tokenService *authmocks.MockTokenService, passwordService *authmocks.MockPasswordService, logger *logsmocks.MockLogger) {
				user := &users.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
				}
				userRepo.On("GetByUsername", "testuser").Return(user, nil)
				passwordService.On("CheckPasswordHash", "password123", "hashed_password").Return(true)
				tokenService.On("GenerateToken", 1, "testuser", "test@example.com").Return("jwt_token", nil)
				logger.On("Info", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "jwt_token", response["token"])
			},
		},
		{
			name: "invalid credentials - wrong password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			setupMocks: func(userRepo *usersmocks.MockRepository, tokenService *authmocks.MockTokenService, passwordService *authmocks.MockPasswordService, logger *logsmocks.MockLogger) {
				user := &users.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
				}
				userRepo.On("GetByEmail", "test@example.com").Return(user, nil)
				passwordService.On("CheckPasswordHash", "wrongpassword", "hashed_password").Return(false)
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Invalid credentials", response["error"])
			},
		},
		{
			name: "user not found",
			requestBody: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			setupMocks: func(userRepo *usersmocks.MockRepository, tokenService *authmocks.MockTokenService, passwordService *authmocks.MockPasswordService, logger *logsmocks.MockLogger) {
				userRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, fmt.Errorf("user not found"))
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Invalid credentials", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := usersmocks.NewMockRepository(t)
			tokenService := authmocks.NewMockTokenService(t)
			passwordService := authmocks.NewMockPasswordService(t)
			logger := logsmocks.NewMockLogger(t)

			handler := NewAuthHandler(userRepo, tokenService, passwordService, logger)

			tt.setupMocks(userRepo, tokenService, passwordService, logger)

			var requestBody []byte
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			router.POST("/login", handler.Login)
			
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			
			c.Request = req
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
		})
	}
}
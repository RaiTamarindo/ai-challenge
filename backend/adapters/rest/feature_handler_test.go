package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	logsmocks "github.com/feature-voting-platform/backend/adapters/logs/mocks"
	"github.com/feature-voting-platform/backend/domain/features"
	featuresmocks "github.com/feature-voting-platform/backend/domain/features/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFeatureHandler_CreateFeature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int
		requestBody    interface{}
		setupMocks     func(*featuresmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful feature creation",
			userID: 1,
			requestBody: map[string]string{
				"title":       "New Feature",
				"description": "Feature Description",
			},
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				repo.On("Create", mock.MatchedBy(func(f *features.Feature) bool {
					return f.Title == "New Feature" && f.Description == "Feature Description" && f.CreatedBy == 1
				})).Return(nil).Run(func(args mock.Arguments) {
					f := args.Get(0).(*features.Feature)
					f.ID = 1
					f.CreatedAt = time.Now()
					f.UpdatedAt = time.Now()
				})
				logger.On("Info", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"message": "Feature created successfully",
			},
		},
		{
			name:   "missing title",
			userID: 1,
			requestBody: map[string]string{
				"description": "Feature Description",
			},
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Title and description are required",
			},
		},
		{
			name:        "invalid JSON",
			userID:      1,
			requestBody: "invalid json",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := featuresmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewFeatureHandler(repo, logger)

			tt.setupMocks(repo, logger)

			var requestBody []byte
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			c.Set("user_id", tt.userID)
			router.POST("/features", handler.CreateFeature)

			req, _ := http.NewRequest(http.MethodPost, "/features", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			c.Request = req
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestFeatureHandler_GetFeatures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()

	tests := []struct {
		name           string
		userID         *int
		queryParams    string
		setupMocks     func(*featuresmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:        "successful retrieval with defaults",
			userID:      intPtr(1),
			queryParams: "",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				mockFeatures := []features.Feature{
					{
						ID:              1,
						Title:           "Feature 1",
						Description:     "Description 1",
						CreatedBy:       1,
						CreatedByUser:   "user1",
						VoteCount:       3,
						CreatedAt:       now,
						UpdatedAt:       now,
						HasUserVoted:    true,
					},
				}
				repo.On("GetAll", 1, 10, intPtr(1)).Return(mockFeatures, 1, nil)
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, float64(1), response["total"])
				assert.Equal(t, float64(1), response["page"])
				assert.Equal(t, float64(10), response["per_page"])
				
				featuresData := response["features"].([]interface{})
				assert.Len(t, featuresData, 1)
				
				feature := featuresData[0].(map[string]interface{})
				assert.Equal(t, float64(1), feature["id"])
				assert.Equal(t, "Feature 1", feature["title"])
				assert.Equal(t, true, feature["has_user_voted"])
			},
		},
		{
			name:        "with pagination parameters",
			userID:      nil,
			queryParams: "?page=2&per_page=5",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				repo.On("GetAll", 2, 5, (*int)(nil)).Return([]features.Feature{}, 0, nil)
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, float64(2), response["page"])
				assert.Equal(t, float64(5), response["per_page"])
			},
		},
		{
			name:        "repository error",
			userID:      nil,
			queryParams: "",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				repo.On("GetAll", 1, 10, (*int)(nil)).Return(nil, 0, fmt.Errorf("database error"))
				logger.On("Error", mock.AnythingOfType("string"), mock.Anything, mock.Anything)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Internal server error", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := featuresmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewFeatureHandler(repo, logger)

			tt.setupMocks(repo, logger)

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}
			router.GET("/features", handler.GetFeatures)

			url := "/features" + tt.queryParams
			req, _ := http.NewRequest(http.MethodGet, url, nil)

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

func TestFeatureHandler_GetFeature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()

	tests := []struct {
		name           string
		userID         *int
		featureID      string
		setupMocks     func(*featuresmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:      "successful retrieval",
			userID:    intPtr(1),
			featureID: "1",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				feature := &features.Feature{
					ID:              1,
					Title:           "Test Feature",
					Description:     "Test Description",
					CreatedBy:       1,
					CreatedByUser:   "testuser",
					VoteCount:       5,
					CreatedAt:       now,
					UpdatedAt:       now,
					HasUserVoted:    true,
				}
				repo.On("GetByID", 1, intPtr(1)).Return(feature, nil)
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, float64(1), response["id"])
				assert.Equal(t, "Test Feature", response["title"])
				assert.Equal(t, true, response["has_user_voted"])
			},
		},
		{
			name:      "feature not found",
			userID:    nil,
			featureID: "999",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				repo.On("GetByID", 999, (*int)(nil)).Return(nil, fmt.Errorf("feature not found"))
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Feature not found", response["error"])
			},
		},
		{
			name:      "invalid feature ID",
			userID:    nil,
			featureID: "invalid",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Invalid feature ID", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := featuresmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewFeatureHandler(repo, logger)

			tt.setupMocks(repo, logger)

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			if tt.userID != nil {
				c.Set("user_id", *tt.userID)
			}
			router.GET("/features/:id", handler.GetFeature)

			url := "/features/" + tt.featureID
			req, _ := http.NewRequest(http.MethodGet, url, nil)

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

func TestFeatureHandler_UpdateFeature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int
		featureID      string
		requestBody    interface{}
		setupMocks     func(*featuresmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful update",
			userID:    1,
			featureID: "1",
			requestBody: map[string]string{
				"title":       "Updated Title",
				"description": "Updated Description",
			},
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				feature := &features.Feature{
					ID:        1,
					CreatedBy: 1,
				}
				repo.On("GetByID", 1, (*int)(nil)).Return(feature, nil)
				repo.On("Update", 1, stringPtr("Updated Title"), stringPtr("Updated Description")).Return(nil)
				logger.On("Info", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Feature updated successfully",
			},
		},
		{
			name:      "unauthorized - not creator",
			userID:    2,
			featureID: "1",
			requestBody: map[string]string{
				"title": "Updated Title",
			},
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				feature := &features.Feature{
					ID:        1,
					CreatedBy: 1,
				}
				repo.On("GetByID", 1, (*int)(nil)).Return(feature, nil)
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "You can only update features you created",
			},
		},
		{
			name:      "feature not found",
			userID:    1,
			featureID: "999",
			requestBody: map[string]string{
				"title": "Updated Title",
			},
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				repo.On("GetByID", 999, (*int)(nil)).Return(nil, fmt.Errorf("feature not found"))
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Feature not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := featuresmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewFeatureHandler(repo, logger)

			tt.setupMocks(repo, logger)

			var requestBody []byte
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			c.Set("user_id", tt.userID)
			router.PUT("/features/:id", handler.UpdateFeature)

			url := "/features/" + tt.featureID
			req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			c.Request = req
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestFeatureHandler_DeleteFeature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int
		featureID      string
		setupMocks     func(*featuresmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful deletion",
			userID:    1,
			featureID: "1",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				feature := &features.Feature{
					ID:        1,
					CreatedBy: 1,
				}
				repo.On("GetByID", 1, (*int)(nil)).Return(feature, nil)
				repo.On("Delete", 1).Return(nil)
				logger.On("Info", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Feature deleted successfully",
			},
		},
		{
			name:      "unauthorized - not creator",
			userID:    2,
			featureID: "1",
			setupMocks: func(repo *featuresmocks.MockRepository, logger *logsmocks.MockLogger) {
				feature := &features.Feature{
					ID:        1,
					CreatedBy: 1,
				}
				repo.On("GetByID", 1, (*int)(nil)).Return(feature, nil)
				logger.On("Warning", mock.AnythingOfType("string"), mock.Anything)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]interface{}{
				"error": "You can only delete features you created",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := featuresmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewFeatureHandler(repo, logger)

			tt.setupMocks(repo, logger)

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			c.Set("user_id", tt.userID)
			router.DELETE("/features/:id", handler.DeleteFeature)

			url := "/features/" + tt.featureID
			req, _ := http.NewRequest(http.MethodDelete, url, nil)

			c.Request = req
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
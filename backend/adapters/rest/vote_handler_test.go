package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	logsmocks "github.com/feature-voting-platform/backend/adapters/logs/mocks"
	featuresmocks "github.com/feature-voting-platform/backend/domain/features/mocks"
	"github.com/feature-voting-platform/backend/domain/votes"
	votesmocks "github.com/feature-voting-platform/backend/domain/votes/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVoteHandler_VoteForFeature_Simple(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int
		featureID      string
		setupMocks     func(*featuresmocks.MockRepository, *votesmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:      "successful vote",
			userID:    1,
			featureID: "1",
			setupMocks: func(featureRepo *featuresmocks.MockRepository, voteRepo *votesmocks.MockRepository, logger *logsmocks.MockLogger) {
				featureRepo.On("FeatureExists", 1).Return(true, nil)
				voteRepo.On("HasUserVoted", 1, 1).Return(false, nil)
				voteRepo.On("AddVote", 1, 1).Return(nil)
				// Simple logger expectations - just expect any calls
				logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Vote cast successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			featureRepo := featuresmocks.NewMockRepository(t)
			voteRepo := votesmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewVoteHandler(featureRepo, voteRepo, logger)

			tt.setupMocks(featureRepo, voteRepo, logger)

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			c.Set("user_id", tt.userID)
			router.POST("/features/:id/vote", handler.VoteForFeature)

			url := "/features/" + tt.featureID + "/vote"
			req, _ := http.NewRequest(http.MethodPost, url, nil)

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

func TestVoteHandler_GetUserVotes_Simple(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()

	tests := []struct {
		name           string
		userID         int
		setupMocks     func(*featuresmocks.MockRepository, *votesmocks.MockRepository, *logsmocks.MockLogger)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successful retrieval with votes",
			userID: 1,
			setupMocks: func(featureRepo *featuresmocks.MockRepository, voteRepo *votesmocks.MockRepository, logger *logsmocks.MockLogger) {
				mockVotes := []votes.Vote{
					{
						ID:        1,
						UserID:    1,
						FeatureID: 10,
						CreatedAt: now,
					},
				}
				voteRepo.On("GetUserVotes", 1).Return(mockVotes, nil)
				// Simple logger expectations
				logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
				logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				votes := response["votes"].([]interface{})
				assert.Len(t, votes, 1)
				assert.Equal(t, float64(1), response["count"])

				vote1 := votes[0].(map[string]interface{})
				assert.Equal(t, float64(1), vote1["id"])
				assert.Equal(t, float64(10), vote1["feature_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			featureRepo := featuresmocks.NewMockRepository(t)
			voteRepo := votesmocks.NewMockRepository(t)
			logger := logsmocks.NewMockLogger(t)
			handler := NewVoteHandler(featureRepo, voteRepo, logger)

			tt.setupMocks(featureRepo, voteRepo, logger)

			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			c.Set("user_id", tt.userID)
			router.GET("/votes", handler.GetUserVotes)

			req, _ := http.NewRequest(http.MethodGet, "/votes", nil)

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
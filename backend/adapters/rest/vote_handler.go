package rest

import (
	"net/http"
	"strconv"

	"github.com/feature-voting-platform/backend/adapters/logs"
	"github.com/feature-voting-platform/backend/domain/features"
	"github.com/feature-voting-platform/backend/domain/votes"
	"github.com/gin-gonic/gin"
)

// VoteHandler handles vote-related HTTP requests
type VoteHandler struct {
	featureRepo features.Repository
	voteRepo    votes.Repository
	logger      logs.Logger
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(featureRepo features.Repository, voteRepo votes.Repository, logger logs.Logger) *VoteHandler {
	return &VoteHandler{
		featureRepo: featureRepo,
		voteRepo:    voteRepo,
		logger:      logger,
	}
}

// VoteForFeature godoc
// @Summary Vote for a feature
// @Description Add a vote for a specific feature
// @Tags votes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} map[string]interface{} "Vote added successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 409 {object} map[string]interface{} "Already voted"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id}/vote [post]
func (h *VoteHandler) VoteForFeature(c *gin.Context) {
	h.logger.Info("Vote for feature request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warning("Invalid feature ID for voting",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest),
			logs.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Vote attempt without authentication",
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Info("Processing vote request",
		logs.WithUserID(userID),
		logs.WithFeatureID(featureID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		h.logger.Error("Failed to check feature existence for voting", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		h.logger.Info("Vote attempt on non-existent feature",
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check if user has already voted
	hasVoted, err := h.voteRepo.HasUserVoted(userID, featureID)
	if err != nil {
		h.logger.Error("Failed to check user vote status", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check vote status"})
		return
	}

	if hasVoted {
		h.logger.Info("Duplicate vote attempt",
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusConflict))
		c.JSON(http.StatusConflict, gin.H{"error": "You have already voted for this feature"})
		return
	}

	// Add vote
	if err := h.voteRepo.AddVote(userID, featureID); err != nil {
		h.logger.Error("Failed to add vote to database", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote"})
		return
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		h.logger.Error("Failed to get updated feature after voting", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	h.logger.Info("Vote added successfully",
		logs.WithUserID(userID),
		logs.WithFeatureID(featureID),
		logs.WithVoteCount(updatedFeature.VoteCount),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Vote added successfully",
		"feature_id": featureID,
		"vote_count": updatedFeature.VoteCount,
		"has_voted":  true,
	})
}

// RemoveVoteFromFeature godoc
// @Summary Remove vote from a feature
// @Description Remove user's vote from a specific feature
// @Tags votes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} map[string]interface{} "Vote removed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Feature or vote not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id}/vote [delete]
func (h *VoteHandler) RemoveVoteFromFeature(c *gin.Context) {
	h.logger.Info("Remove vote from feature request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warning("Invalid feature ID for vote removal",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest),
			logs.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Vote removal attempt without authentication",
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Info("Processing vote removal request",
		logs.WithUserID(userID),
		logs.WithFeatureID(featureID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		h.logger.Error("Failed to check feature existence for vote removal", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		h.logger.Info("Vote removal attempt on non-existent feature",
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Remove vote
	if err := h.voteRepo.RemoveVote(userID, featureID); err != nil {
		if err.Error() == "vote not found" {
			h.logger.Info("Vote removal attempt on non-existent vote",
				logs.WithUserID(userID),
				logs.WithFeatureID(featureID),
				logs.WithMethod(c.Request.Method),
				logs.WithPath(c.Request.URL.Path),
				logs.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Vote not found"})
			return
		}
		h.logger.Error("Failed to remove vote from database", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
		return
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		h.logger.Error("Failed to get updated feature after vote removal", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	h.logger.Info("Vote removed successfully",
		logs.WithUserID(userID),
		logs.WithFeatureID(featureID),
		logs.WithVoteCount(updatedFeature.VoteCount),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Vote removed successfully",
		"feature_id": featureID,
		"vote_count": updatedFeature.VoteCount,
		"has_voted":  false,
	})
}

// GetUserVotes godoc
// @Summary Get user's votes
// @Description Get all votes made by the authenticated user
// @Tags votes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User's votes"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /votes/my [get]
func (h *VoteHandler) GetUserVotes(c *gin.Context) {
	h.logger.Info("Get user votes request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Get user votes attempt without authentication",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Debug("Fetching user's votes",
		logs.WithUserID(userID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	votesList, err := h.voteRepo.GetUserVotes(userID)
	if err != nil {
		h.logger.Error("Failed to get user votes from database", err,
			logs.WithUserID(userID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user votes"})
		return
	}

	h.logger.Info("User votes retrieved successfully",
		logs.WithUserID(userID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("vote_count", len(votesList)))

	c.JSON(http.StatusOK, gin.H{
		"votes": votesList,
		"count": len(votesList),
	})
}

// ToggleVote godoc
// @Summary Toggle vote for a feature
// @Description Add vote if not voted, remove vote if already voted
// @Tags votes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} map[string]interface{} "Vote toggled successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id}/toggle-vote [post]
func (h *VoteHandler) ToggleVote(c *gin.Context) {
	h.logger.Info("Toggle vote request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warning("Invalid feature ID for toggle vote",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest),
			logs.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Toggle vote attempt without authentication",
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Info("Processing toggle vote request",
		logs.WithUserID(userID),
		logs.WithFeatureID(featureID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		h.logger.Error("Failed to check feature existence for toggle vote", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		h.logger.Info("Toggle vote attempt on non-existent feature",
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check if user has already voted
	hasVoted, err := h.voteRepo.HasUserVoted(userID, featureID)
	if err != nil {
		h.logger.Error("Failed to check user vote status for toggle", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check vote status"})
		return
	}

	var message string
	var action string
	if hasVoted {
		// Remove vote
		if err := h.voteRepo.RemoveVote(userID, featureID); err != nil {
			h.logger.Error("Failed to remove vote during toggle", err,
				logs.WithUserID(userID),
				logs.WithFeatureID(featureID),
				logs.WithMethod(c.Request.Method),
				logs.WithPath(c.Request.URL.Path),
				logs.WithStatusCode(http.StatusInternalServerError))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
			return
		}
		message = "Vote removed successfully"
		action = "removed"
		hasVoted = false
	} else {
		// Add vote
		if err := h.voteRepo.AddVote(userID, featureID); err != nil {
			h.logger.Error("Failed to add vote during toggle", err,
				logs.WithUserID(userID),
				logs.WithFeatureID(featureID),
				logs.WithMethod(c.Request.Method),
				logs.WithPath(c.Request.URL.Path),
				logs.WithStatusCode(http.StatusInternalServerError))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote"})
			return
		}
		message = "Vote added successfully"
		action = "added"
		hasVoted = true
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		h.logger.Error("Failed to get updated feature after toggle", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(featureID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	h.logger.Info("Vote toggled successfully",
		logs.WithUserID(userID),
		logs.WithFeatureID(featureID),
		logs.WithVoteCount(updatedFeature.VoteCount),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("vote_action", action),
		logs.WithMetadata("has_voted", hasVoted))

	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"feature_id": featureID,
		"vote_count": updatedFeature.VoteCount,
		"has_voted":  hasVoted,
	})
}
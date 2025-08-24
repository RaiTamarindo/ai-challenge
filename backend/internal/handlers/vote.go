package handlers

import (
	"net/http"
	"strconv"

	"github.com/feature-voting-platform/backend/internal/middleware"
	"github.com/feature-voting-platform/backend/internal/repository"
	"github.com/feature-voting-platform/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type VoteHandler struct {
	featureRepo *repository.FeatureRepository
}

func NewVoteHandler(featureRepo *repository.FeatureRepository) *VoteHandler {
	return &VoteHandler{
		featureRepo: featureRepo,
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
	utils.LogInfo("Vote for feature request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogWarning("Invalid feature ID for voting",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest),
			utils.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Vote attempt without authentication",
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogInfo("Processing vote request",
		utils.WithUserID(userID),
		utils.WithFeatureID(featureID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		utils.LogError("Failed to check feature existence for voting", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		utils.LogInfo("Vote attempt on non-existent feature",
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check if user has already voted
	hasVoted, err := h.featureRepo.HasUserVoted(userID, featureID)
	if err != nil {
		utils.LogError("Failed to check user vote status", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check vote status"})
		return
	}

	if hasVoted {
		utils.LogInfo("Duplicate vote attempt",
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusConflict))
		c.JSON(http.StatusConflict, gin.H{"error": "You have already voted for this feature"})
		return
	}

	// Add vote
	if err := h.featureRepo.AddVote(userID, featureID); err != nil {
		utils.LogError("Failed to add vote to database", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote"})
		return
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		utils.LogError("Failed to get updated feature after voting", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	utils.LogInfo("Vote added successfully",
		utils.WithUserID(userID),
		utils.WithFeatureID(featureID),
		utils.WithVoteCount(updatedFeature.VoteCount),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK))

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
	utils.LogInfo("Remove vote from feature request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogWarning("Invalid feature ID for vote removal",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest),
			utils.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Vote removal attempt without authentication",
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogInfo("Processing vote removal request",
		utils.WithUserID(userID),
		utils.WithFeatureID(featureID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		utils.LogError("Failed to check feature existence for vote removal", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		utils.LogInfo("Vote removal attempt on non-existent feature",
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Remove vote
	if err := h.featureRepo.RemoveVote(userID, featureID); err != nil {
		if err.Error() == "vote not found" {
			utils.LogInfo("Vote removal attempt on non-existent vote",
				utils.WithUserID(userID),
				utils.WithFeatureID(featureID),
				utils.WithMethod(c.Request.Method),
				utils.WithPath(c.Request.URL.Path),
				utils.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Vote not found"})
			return
		}
		utils.LogError("Failed to remove vote from database", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
		return
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		utils.LogError("Failed to get updated feature after vote removal", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	utils.LogInfo("Vote removed successfully",
		utils.WithUserID(userID),
		utils.WithFeatureID(featureID),
		utils.WithVoteCount(updatedFeature.VoteCount),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK))

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
	utils.LogInfo("Get user votes request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Get user votes attempt without authentication",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogDebug("Fetching user's votes",
		utils.WithUserID(userID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	votes, err := h.featureRepo.GetUserVotes(userID)
	if err != nil {
		utils.LogError("Failed to get user votes from database", err,
			utils.WithUserID(userID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user votes"})
		return
	}

	utils.LogInfo("User votes retrieved successfully",
		utils.WithUserID(userID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("vote_count", len(votes)))

	c.JSON(http.StatusOK, gin.H{
		"votes": votes,
		"count": len(votes),
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
	utils.LogInfo("Toggle vote request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogWarning("Invalid feature ID for toggle vote",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest),
			utils.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Toggle vote attempt without authentication",
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogInfo("Processing toggle vote request",
		utils.WithUserID(userID),
		utils.WithFeatureID(featureID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		utils.LogError("Failed to check feature existence for toggle vote", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		utils.LogInfo("Toggle vote attempt on non-existent feature",
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check if user has already voted
	hasVoted, err := h.featureRepo.HasUserVoted(userID, featureID)
	if err != nil {
		utils.LogError("Failed to check user vote status for toggle", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check vote status"})
		return
	}

	var message string
	var action string
	if hasVoted {
		// Remove vote
		if err := h.featureRepo.RemoveVote(userID, featureID); err != nil {
			utils.LogError("Failed to remove vote during toggle", err,
				utils.WithUserID(userID),
				utils.WithFeatureID(featureID),
				utils.WithMethod(c.Request.Method),
				utils.WithPath(c.Request.URL.Path),
				utils.WithStatusCode(http.StatusInternalServerError))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
			return
		}
		message = "Vote removed successfully"
		action = "removed"
		hasVoted = false
	} else {
		// Add vote
		if err := h.featureRepo.AddVote(userID, featureID); err != nil {
			utils.LogError("Failed to add vote during toggle", err,
				utils.WithUserID(userID),
				utils.WithFeatureID(featureID),
				utils.WithMethod(c.Request.Method),
				utils.WithPath(c.Request.URL.Path),
				utils.WithStatusCode(http.StatusInternalServerError))
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
		utils.LogError("Failed to get updated feature after toggle", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(featureID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	utils.LogInfo("Vote toggled successfully",
		utils.WithUserID(userID),
		utils.WithFeatureID(featureID),
		utils.WithVoteCount(updatedFeature.VoteCount),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("vote_action", action),
		utils.WithMetadata("has_voted", hasVoted))

	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"feature_id": featureID,
		"vote_count": updatedFeature.VoteCount,
		"has_voted":  hasVoted,
	})
}
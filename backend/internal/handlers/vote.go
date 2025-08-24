package handlers

import (
	"net/http"
	"strconv"

	"github.com/feature-voting-platform/backend/internal/middleware"
	"github.com/feature-voting-platform/backend/internal/repository"
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
	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check if user has already voted
	hasVoted, err := h.featureRepo.HasUserVoted(userID, featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check vote status"})
		return
	}

	if hasVoted {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already voted for this feature"})
		return
	}

	// Add vote
	if err := h.featureRepo.AddVote(userID, featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote"})
		return
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

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
	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Remove vote
	if err := h.featureRepo.RemoveVote(userID, featureID); err != nil {
		if err.Error() == "vote not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vote not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
		return
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

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
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	votes, err := h.featureRepo.GetUserVotes(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user votes"})
		return
	}

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
	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if feature exists
	exists, err = h.featureRepo.FeatureExists(featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check feature existence"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Check if user has already voted
	hasVoted, err := h.featureRepo.HasUserVoted(userID, featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check vote status"})
		return
	}

	var message string
	if hasVoted {
		// Remove vote
		if err := h.featureRepo.RemoveVote(userID, featureID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
			return
		}
		message = "Vote removed successfully"
		hasVoted = false
	} else {
		// Add vote
		if err := h.featureRepo.AddVote(userID, featureID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote"})
			return
		}
		message = "Vote added successfully"
		hasVoted = true
	}

	// Get updated feature to return current vote count
	updatedFeature, err := h.featureRepo.GetByID(featureID, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    message,
		"feature_id": featureID,
		"vote_count": updatedFeature.VoteCount,
		"has_voted":  hasVoted,
	})
}
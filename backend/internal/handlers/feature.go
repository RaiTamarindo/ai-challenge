package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/feature-voting-platform/backend/internal/middleware"
	"github.com/feature-voting-platform/backend/internal/models"
	"github.com/feature-voting-platform/backend/internal/repository"
)

type FeatureHandler struct {
	featureRepo *repository.FeatureRepository
}

func NewFeatureHandler(featureRepo *repository.FeatureRepository) *FeatureHandler {
	return &FeatureHandler{
		featureRepo: featureRepo,
	}
}

// CreateFeature godoc
// @Summary Create a new feature
// @Description Create a new feature request
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param feature body models.CreateFeatureRequest true "Feature data"
// @Success 201 {object} models.Feature "Feature created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features [post]
func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	var req models.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	feature := &models.Feature{
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.featureRepo.Create(feature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feature"})
		return
	}

	// Get the created feature with user info
	createdFeature, err := h.featureRepo.GetByID(feature.ID, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get created feature"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feature created successfully",
		"feature": createdFeature,
	})
}

// GetFeatures godoc
// @Summary Get all features
// @Description Get a paginated list of all features
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} models.FeatureListResponse "List of features"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features [get]
func (h *FeatureHandler) GetFeatures(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	perPage := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	// Get optional user ID for vote status
	userID := middleware.GetOptionalUserID(c)

	features, total, err := h.featureRepo.GetAll(page, perPage, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get features"})
		return
	}

	response := models.FeatureListResponse{
		Features: features,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}

	c.JSON(http.StatusOK, response)
}

// GetFeature godoc
// @Summary Get a feature by ID
// @Description Get detailed information about a specific feature
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} models.Feature "Feature details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id} [get]
func (h *FeatureHandler) GetFeature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// Get optional user ID for vote status
	userID := middleware.GetOptionalUserID(c)

	feature, err := h.featureRepo.GetByID(id, userID)
	if err != nil {
		if err.Error() == "feature not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feature": feature,
	})
}

// UpdateFeature godoc
// @Summary Update a feature
// @Description Update an existing feature (only by creator)
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feature ID"
// @Param feature body models.UpdateFeatureRequest true "Updated feature data"
// @Success 200 {object} models.Feature "Updated feature"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id} [put]
func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if feature exists and user is the creator
	feature, err := h.featureRepo.GetByID(id, nil)
	if err != nil {
		if err.Error() == "feature not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	if feature.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own features"})
		return
	}

	// Update feature
	if err := h.featureRepo.Update(id, req.Title, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature"})
		return
	}

	// Get updated feature
	updatedFeature, err := h.featureRepo.GetByID(id, &userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature updated successfully",
		"feature": updatedFeature,
	})
}

// DeleteFeature godoc
// @Summary Delete a feature
// @Description Delete an existing feature (only by creator)
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} map[string]interface{} "Feature deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id} [delete]
func (h *FeatureHandler) DeleteFeature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if feature exists and user is the creator
	feature, err := h.featureRepo.GetByID(id, nil)
	if err != nil {
		if err.Error() == "feature not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	if feature.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own features"})
		return
	}

	// Delete feature
	if err := h.featureRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feature deleted successfully",
	})
}

// GetMyFeatures godoc
// @Summary Get user's features
// @Description Get all features created by the authenticated user
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User's features"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/my [get]
func (h *FeatureHandler) GetMyFeatures(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	features, err := h.featureRepo.GetByCreatedBy(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user features"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"count":    len(features),
	})
}
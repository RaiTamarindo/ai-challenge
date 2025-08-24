package handlers

import (
	"net/http"
	"strconv"

	"github.com/feature-voting-platform/backend/internal/middleware"
	"github.com/feature-voting-platform/backend/internal/models"
	"github.com/feature-voting-platform/backend/internal/repository"
	"github.com/feature-voting-platform/backend/pkg/utils"
	"github.com/gin-gonic/gin"
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
	utils.LogInfo("Create feature request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	var req models.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Create feature request validation failed", err,
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Create feature attempt without authentication",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogInfo("Creating new feature",
		utils.WithUserID(userID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithMetadata("feature_title", req.Title),
		utils.WithMetadata("description_length", len(req.Description)))

	feature := &models.Feature{
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.featureRepo.Create(feature); err != nil {
		utils.LogError("Failed to create feature in database", err,
			utils.WithUserID(userID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError),
			utils.WithMetadata("feature_title", req.Title))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feature"})
		return
	}

	// Get the created feature with user info
	createdFeature, err := h.featureRepo.GetByID(feature.ID, &userID)
	if err != nil {
		utils.LogError("Failed to get created feature", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(feature.ID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get created feature"})
		return
	}

	utils.LogInfo("Feature created successfully",
		utils.WithUserID(userID),
		utils.WithFeatureID(createdFeature.ID),
		utils.WithVoteCount(createdFeature.VoteCount),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusCreated),
		utils.WithMetadata("feature_title", createdFeature.Title))

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
	utils.LogInfo("Get features request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

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

	logFields := []utils.LogField{
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithMetadata("page", page),
		utils.WithMetadata("per_page", perPage),
	}
	if userID != nil {
		logFields = append(logFields, utils.WithUserID(*userID))
	}

	utils.LogDebug("Fetching features with pagination", logFields...)

	features, total, err := h.featureRepo.GetAll(page, perPage, userID)
	if err != nil {
		utils.LogError("Failed to get features from database", err,
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError),
			utils.WithMetadata("page", page),
			utils.WithMetadata("per_page", perPage))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get features"})
		return
	}

	response := models.FeatureListResponse{
		Features: features,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}

	logFields = append(logFields, 
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("total_features", total),
		utils.WithMetadata("returned_count", len(features)))

	utils.LogInfo("Features retrieved successfully", logFields...)

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
	utils.LogInfo("Get single feature request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogWarning("Invalid feature ID provided",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest),
			utils.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// Get optional user ID for vote status
	userID := middleware.GetOptionalUserID(c)

	logFields := []utils.LogField{
		utils.WithFeatureID(id),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
	}
	if userID != nil {
		logFields = append(logFields, utils.WithUserID(*userID))
	}

	utils.LogDebug("Fetching feature by ID", logFields...)

	feature, err := h.featureRepo.GetByID(id, userID)
	if err != nil {
		if err.Error() == "feature not found" {
			utils.LogInfo("Feature not found",
				utils.WithFeatureID(id),
				utils.WithMethod(c.Request.Method),
				utils.WithPath(c.Request.URL.Path),
				utils.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		utils.LogError("Failed to get feature from database", err,
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	utils.LogInfo("Feature retrieved successfully",
		utils.WithFeatureID(feature.ID),
		utils.WithVoteCount(feature.VoteCount),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("feature_title", feature.Title),
		utils.WithMetadata("created_by", feature.CreatedBy))

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
	utils.LogInfo("Update feature request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogWarning("Invalid feature ID for update",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest),
			utils.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Update feature attempt without authentication",
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Update feature request validation failed", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.LogInfo("Processing feature update request",
		utils.WithUserID(userID),
		utils.WithFeatureID(id),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithMetadata("new_title", req.Title))

	// Check if feature exists and user is the creator
	feature, err := h.featureRepo.GetByID(id, nil)
	if err != nil {
		if err.Error() == "feature not found" {
			utils.LogInfo("Update attempt on non-existent feature",
				utils.WithUserID(userID),
				utils.WithFeatureID(id),
				utils.WithMethod(c.Request.Method),
				utils.WithPath(c.Request.URL.Path),
				utils.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		utils.LogError("Failed to get feature for update validation", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	if feature.CreatedBy != userID {
		utils.LogWarning("Unauthorized feature update attempt",
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusForbidden),
			utils.WithMetadata("feature_owner_id", feature.CreatedBy))
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own features"})
		return
	}

	// Update feature
	if err := h.featureRepo.Update(id, req.Title, req.Description); err != nil {
		utils.LogError("Failed to update feature in database", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError),
			utils.WithMetadata("new_title", req.Title))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature"})
		return
	}

	// Get updated feature
	updatedFeature, err := h.featureRepo.GetByID(id, &userID)
	if err != nil {
		utils.LogError("Failed to get updated feature", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	utils.LogInfo("Feature updated successfully",
		utils.WithUserID(userID),
		utils.WithFeatureID(updatedFeature.ID),
		utils.WithVoteCount(updatedFeature.VoteCount),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("updated_title", updatedFeature.Title),
		utils.WithMetadata("description_length", len(updatedFeature.Description)))

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
	utils.LogInfo("Delete feature request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogWarning("Invalid feature ID for deletion",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest),
			utils.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Delete feature attempt without authentication",
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogInfo("Processing feature deletion request",
		utils.WithUserID(userID),
		utils.WithFeatureID(id),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	// Check if feature exists and user is the creator
	feature, err := h.featureRepo.GetByID(id, nil)
	if err != nil {
		if err.Error() == "feature not found" {
			utils.LogInfo("Delete attempt on non-existent feature",
				utils.WithUserID(userID),
				utils.WithFeatureID(id),
				utils.WithMethod(c.Request.Method),
				utils.WithPath(c.Request.URL.Path),
				utils.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		utils.LogError("Failed to get feature for deletion validation", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	if feature.CreatedBy != userID {
		utils.LogWarning("Unauthorized feature deletion attempt",
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusForbidden),
			utils.WithMetadata("feature_owner_id", feature.CreatedBy),
			utils.WithMetadata("feature_title", feature.Title))
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own features"})
		return
	}

	// Delete feature
	if err := h.featureRepo.Delete(id); err != nil {
		utils.LogError("Failed to delete feature from database", err,
			utils.WithUserID(userID),
			utils.WithFeatureID(id),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError),
			utils.WithMetadata("feature_title", feature.Title))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature"})
		return
	}

	utils.LogInfo("Feature deleted successfully",
		utils.WithUserID(userID),
		utils.WithFeatureID(id),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("deleted_title", feature.Title),
		utils.WithMetadata("deleted_vote_count", feature.VoteCount))

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
	utils.LogInfo("Get my features request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.LogWarning("Get my features attempt without authentication",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	utils.LogDebug("Fetching user's created features",
		utils.WithUserID(userID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	features, err := h.featureRepo.GetByCreatedBy(userID)
	if err != nil {
		utils.LogError("Failed to get user features from database", err,
			utils.WithUserID(userID),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user features"})
		return
	}

	utils.LogInfo("User features retrieved successfully",
		utils.WithUserID(userID),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK),
		utils.WithMetadata("feature_count", len(features)))

	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"count":    len(features),
	})
}
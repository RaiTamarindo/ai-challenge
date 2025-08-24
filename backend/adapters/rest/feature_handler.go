package rest

import (
	"net/http"
	"strconv"

	"github.com/feature-voting-platform/backend/adapters/logs"
	"github.com/feature-voting-platform/backend/domain/features"
	"github.com/gin-gonic/gin"
)

// FeatureHandler handles feature-related HTTP requests
type FeatureHandler struct {
	featureRepo features.Repository
	logger      logs.Logger
}

// NewFeatureHandler creates a new feature handler
func NewFeatureHandler(featureRepo features.Repository, logger logs.Logger) *FeatureHandler {
	return &FeatureHandler{
		featureRepo: featureRepo,
		logger:      logger,
	}
}

// CreateFeature godoc
// @Summary Create a new feature
// @Description Create a new feature request
// @Tags features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param feature body features.CreateFeatureRequest true "Feature data"
// @Success 201 {object} features.Feature "Feature created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features [post]
func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	h.logger.Info("Create feature request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	var req features.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Create feature request validation failed", err,
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Create feature attempt without authentication",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Info("Creating new feature",
		logs.WithUserID(userID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithMetadata("feature_title", req.Title),
		logs.WithMetadata("description_length", len(req.Description)))

	feature := &features.Feature{
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.featureRepo.Create(feature); err != nil {
		h.logger.Error("Failed to create feature in database", err,
			logs.WithUserID(userID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError),
			logs.WithMetadata("feature_title", req.Title))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feature"})
		return
	}

	// Get the created feature with user info
	createdFeature, err := h.featureRepo.GetByID(feature.ID, &userID)
	if err != nil {
		h.logger.Error("Failed to get created feature", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(feature.ID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get created feature"})
		return
	}

	h.logger.Info("Feature created successfully",
		logs.WithUserID(userID),
		logs.WithFeatureID(createdFeature.ID),
		logs.WithVoteCount(createdFeature.VoteCount),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusCreated),
		logs.WithMetadata("feature_title", createdFeature.Title))

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
// @Success 200 {object} features.FeatureListResponse "List of features"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features [get]
func (h *FeatureHandler) GetFeatures(c *gin.Context) {
	h.logger.Info("Get features request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

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
	userID := getOptionalUserID(c)

	logFields := []logs.LogField{
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithMetadata("page", page),
		logs.WithMetadata("per_page", perPage),
	}
	if userID != nil {
		logFields = append(logFields, logs.WithUserID(*userID))
	}

	h.logger.Debug("Fetching features with pagination", logFields...)

	featuresList, total, err := h.featureRepo.GetAll(page, perPage, userID)
	if err != nil {
		h.logger.Error("Failed to get features from database", err,
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError),
			logs.WithMetadata("page", page),
			logs.WithMetadata("per_page", perPage))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get features"})
		return
	}

	response := features.FeatureListResponse{
		Features: featuresList,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}

	logFields = append(logFields,
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("total_features", total),
		logs.WithMetadata("returned_count", len(featuresList)))

	h.logger.Info("Features retrieved successfully", logFields...)

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
// @Success 200 {object} features.Feature "Feature details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id} [get]
func (h *FeatureHandler) GetFeature(c *gin.Context) {
	h.logger.Info("Get single feature request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warning("Invalid feature ID provided",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest),
			logs.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// Get optional user ID for vote status
	userID := getOptionalUserID(c)

	logFields := []logs.LogField{
		logs.WithFeatureID(id),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
	}
	if userID != nil {
		logFields = append(logFields, logs.WithUserID(*userID))
	}

	h.logger.Debug("Fetching feature by ID", logFields...)

	feature, err := h.featureRepo.GetByID(id, userID)
	if err != nil {
		if err.Error() == "feature not found" {
			h.logger.Info("Feature not found",
				logs.WithFeatureID(id),
				logs.WithMethod(c.Request.Method),
				logs.WithPath(c.Request.URL.Path),
				logs.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		h.logger.Error("Failed to get feature from database", err,
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	h.logger.Info("Feature retrieved successfully",
		logs.WithFeatureID(feature.ID),
		logs.WithVoteCount(feature.VoteCount),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("feature_title", feature.Title),
		logs.WithMetadata("created_by", feature.CreatedBy))

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
// @Param feature body features.UpdateFeatureRequest true "Updated feature data"
// @Success 200 {object} features.Feature "Updated feature"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Feature not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /features/{id} [put]
func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	h.logger.Info("Update feature request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warning("Invalid feature ID for update",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest),
			logs.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Update feature attempt without authentication",
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req features.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Update feature request validation failed", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logFields := []logs.LogField{
		logs.WithUserID(userID),
		logs.WithFeatureID(id),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
	}
	if req.Title != nil {
		logFields = append(logFields, logs.WithMetadata("new_title", *req.Title))
	}
	if req.Description != nil {
		logFields = append(logFields, logs.WithMetadata("description_length", len(*req.Description)))
	}

	h.logger.Info("Processing feature update request", logFields...)

	// Check if feature exists and user is the creator
	feature, err := h.featureRepo.GetByID(id, nil)
	if err != nil {
		if err.Error() == "feature not found" {
			h.logger.Info("Update attempt on non-existent feature",
				logs.WithUserID(userID),
				logs.WithFeatureID(id),
				logs.WithMethod(c.Request.Method),
				logs.WithPath(c.Request.URL.Path),
				logs.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		h.logger.Error("Failed to get feature for update validation", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	if feature.CreatedBy != userID {
		h.logger.Warning("Unauthorized feature update attempt",
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusForbidden),
			logs.WithMetadata("feature_owner_id", feature.CreatedBy))
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own features"})
		return
	}

	// Update feature
	if err := h.featureRepo.Update(id, req.Title, req.Description); err != nil {
		h.logger.Error("Failed to update feature in database", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feature"})
		return
	}

	// Get updated feature
	updatedFeature, err := h.featureRepo.GetByID(id, &userID)
	if err != nil {
		h.logger.Error("Failed to get updated feature", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated feature"})
		return
	}

	h.logger.Info("Feature updated successfully",
		logs.WithUserID(userID),
		logs.WithFeatureID(updatedFeature.ID),
		logs.WithVoteCount(updatedFeature.VoteCount),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("updated_title", updatedFeature.Title),
		logs.WithMetadata("description_length", len(updatedFeature.Description)))

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
	h.logger.Info("Delete feature request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warning("Invalid feature ID for deletion",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest),
			logs.WithMetadata("provided_id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Delete feature attempt without authentication",
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Info("Processing feature deletion request",
		logs.WithUserID(userID),
		logs.WithFeatureID(id),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	// Check if feature exists and user is the creator
	feature, err := h.featureRepo.GetByID(id, nil)
	if err != nil {
		if err.Error() == "feature not found" {
			h.logger.Info("Delete attempt on non-existent feature",
				logs.WithUserID(userID),
				logs.WithFeatureID(id),
				logs.WithMethod(c.Request.Method),
				logs.WithPath(c.Request.URL.Path),
				logs.WithStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
			return
		}
		h.logger.Error("Failed to get feature for deletion validation", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feature"})
		return
	}

	if feature.CreatedBy != userID {
		h.logger.Warning("Unauthorized feature deletion attempt",
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusForbidden),
			logs.WithMetadata("feature_owner_id", feature.CreatedBy),
			logs.WithMetadata("feature_title", feature.Title))
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own features"})
		return
	}

	// Delete feature
	if err := h.featureRepo.Delete(id); err != nil {
		h.logger.Error("Failed to delete feature from database", err,
			logs.WithUserID(userID),
			logs.WithFeatureID(id),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError),
			logs.WithMetadata("feature_title", feature.Title))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feature"})
		return
	}

	h.logger.Info("Feature deleted successfully",
		logs.WithUserID(userID),
		logs.WithFeatureID(id),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("deleted_title", feature.Title),
		logs.WithMetadata("deleted_vote_count", feature.VoteCount))

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
	h.logger.Info("Get my features request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	userID, exists := getUserID(c)
	if !exists {
		h.logger.Warning("Get my features attempt without authentication",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	h.logger.Debug("Fetching user's created features",
		logs.WithUserID(userID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	featuresList, err := h.featureRepo.GetByCreatedBy(userID)
	if err != nil {
		h.logger.Error("Failed to get user features from database", err,
			logs.WithUserID(userID),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user features"})
		return
	}

	h.logger.Info("User features retrieved successfully",
		logs.WithUserID(userID),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK),
		logs.WithMetadata("feature_count", len(featuresList)))

	c.JSON(http.StatusOK, gin.H{
		"features": featuresList,
		"count":    len(featuresList),
	})
}

// Helper functions
func getUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

func getOptionalUserID(c *gin.Context) *int {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil
	}
	uid := userID.(int)
	return &uid
}
package handlers

import (
	"net/http"
	"strings"

	"github.com/feature-voting-platform/backend/internal/models"
	"github.com/feature-voting-platform/backend/internal/repository"
	"github.com/feature-voting-platform/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "User login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with token"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	utils.LogInfo("Login attempt started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Login request validation failed", err,
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	utils.LogInfo("User login attempt",
		utils.WithEmail(email),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	// Get user by email
	user, err := h.userRepo.GetByEmail(email)
	if err != nil {
		utils.LogWarning("Login attempt with non-existent email",
			utils.WithEmail(email),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		utils.LogWarning("Login attempt with invalid password",
			utils.WithEmail(email),
			utils.WithUserID(user.ID),
			utils.WithUsername(user.Username),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username, user.Email, h.jwtSecret)
	if err != nil {
		utils.LogError("Failed to generate JWT token", err,
			utils.WithUserID(user.ID),
			utils.WithUsername(user.Username),
			utils.WithEmail(email),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	utils.LogInfo("User login successful",
		utils.WithUserID(user.ID),
		utils.WithUsername(user.Username),
		utils.WithEmail(email),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK))

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user.ToResponse(),
		"token":   token,
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	utils.LogInfo("Get user profile request started",
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogWarning("Profile request without authentication",
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt := userID.(int)
	utils.LogDebug("Fetching user profile",
		utils.WithUserID(userIDInt),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path))

	user, err := h.userRepo.GetByID(userIDInt)
	if err != nil {
		utils.LogError("Failed to get user profile from database", err,
			utils.WithUserID(userIDInt),
			utils.WithMethod(c.Request.Method),
			utils.WithPath(c.Request.URL.Path),
			utils.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	utils.LogInfo("User profile retrieved successfully",
		utils.WithUserID(user.ID),
		utils.WithUsername(user.Username),
		utils.WithEmail(user.Email),
		utils.WithMethod(c.Request.Method),
		utils.WithPath(c.Request.URL.Path),
		utils.WithStatusCode(http.StatusOK))

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToResponse(),
	})
}
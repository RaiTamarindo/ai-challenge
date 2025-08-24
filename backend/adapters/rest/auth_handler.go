package rest

import (
	"net/http"
	"strings"

	"github.com/feature-voting-platform/backend/adapters/auth"
	"github.com/feature-voting-platform/backend/adapters/logs"
	"github.com/feature-voting-platform/backend/domain/users"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRepo        users.Repository
	tokenService    auth.TokenService
	passwordService auth.PasswordService
	logger          logs.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo users.Repository,
	tokenService auth.TokenService,
	passwordService auth.PasswordService,
	logger logs.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		tokenService:    tokenService,
		passwordService: passwordService,
		logger:          logger,
	}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body users.LoginRequest true "User login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with token"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	h.logger.Info("Login attempt started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	var req users.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Login request validation failed", err,
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	h.logger.Info("User login attempt",
		logs.WithEmail(email),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	// Get user by email
	user, err := h.userRepo.GetByEmail(email)
	if err != nil {
		h.logger.Warning("Login attempt with non-existent email",
			logs.WithEmail(email),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !h.passwordService.CheckPasswordHash(req.Password, user.PasswordHash) {
		h.logger.Warning("Login attempt with invalid password",
			logs.WithEmail(email),
			logs.WithUserID(user.ID),
			logs.WithUsername(user.Username),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.tokenService.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", err,
			logs.WithUserID(user.ID),
			logs.WithUsername(user.Username),
			logs.WithEmail(email),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("User login successful",
		logs.WithUserID(user.ID),
		logs.WithUsername(user.Username),
		logs.WithEmail(email),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK))

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
// @Success 200 {object} users.UserResponse "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	h.logger.Info("Get user profile request started",
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warning("Profile request without authentication",
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusUnauthorized))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDInt := userID.(int)
	h.logger.Debug("Fetching user profile",
		logs.WithUserID(userIDInt),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path))

	user, err := h.userRepo.GetByID(userIDInt)
	if err != nil {
		h.logger.Error("Failed to get user profile from database", err,
			logs.WithUserID(userIDInt),
			logs.WithMethod(c.Request.Method),
			logs.WithPath(c.Request.URL.Path),
			logs.WithStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	h.logger.Info("User profile retrieved successfully",
		logs.WithUserID(user.ID),
		logs.WithUsername(user.Username),
		logs.WithEmail(user.Email),
		logs.WithMethod(c.Request.Method),
		logs.WithPath(c.Request.URL.Path),
		logs.WithStatusCode(http.StatusOK))

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToResponse(),
	})
}
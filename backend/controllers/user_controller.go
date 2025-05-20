package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ginchat/models"
	"github.com/ginchat/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserController handles user-related requests
type UserController struct {
	DB *gorm.DB
}

// NewUserController creates a new UserController
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User Registration Data"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 409 {object} map[string]interface{} "User already exists"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /auth/register [post]
func (uc *UserController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password strength
	if err := validatePasswordStrength(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if result := uc.DB.Where("email = ?", req.Email).Or("username = ?", req.Username).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
		return
	}

	// Create new user
	now := time.Now()
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password, // Will be hashed by BeforeSave hook
		Role:      "member",
		Status:    "offline",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save user to database
	if result := uc.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.UserID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Log the registration
	logUserActivity(c, user.UserID, "User registered")

	// Return user data and token
	c.JSON(http.StatusCreated, gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}

// Login godoc
// @Summary Login a user
// @Description Login with email and password to get authentication token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "User Login Data"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /auth/login [post]
func (uc *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if result := uc.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		// Use the same error message for security (don't reveal if email exists)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password using bcrypt's secure comparison
	if !user.CheckPassword(req.Password) {
		// Add a small delay to prevent timing attacks
		time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Update user status
	user.IsLogin = true
	user.Status = "online"
	now := time.Now()
	user.LastLoginAt = &now
	user.Heartbeat = &now
	uc.DB.Save(&user)

	// Generate JWT token
	token, err := utils.GenerateJWT(user.UserID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Log the login
	logUserActivity(c, user.UserID, "User logged in")

	// Return user data and token
	c.JSON(http.StatusOK, gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}

// Logout godoc
// @Summary Logout a user
// @Description Logout the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 401 {object} map[string]interface{} "User not authenticated"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /auth/logout [post]
func (uc *UserController) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Find user
	var user models.User
	if result := uc.DB.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user status
	user.IsLogin = false
	user.Status = "offline"
	uc.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// validatePasswordStrength checks if a password meets the minimum security requirements
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common passwords
	commonPasswords := []string{"password", "123456", "qwerty", "admin", "welcome"}
	passwordLower := strings.ToLower(password)
	for _, common := range commonPasswords {
		if strings.Contains(passwordLower, common) {
			return fmt.Errorf("password contains a common pattern that is easily guessable")
		}
	}

	return nil
}

// logUserActivity logs user activities for auditing purposes
func logUserActivity(c *gin.Context, userID uint, activity string) {
	// Get client IP
	clientIP := c.ClientIP()

	// Get user agent
	userAgent := c.Request.UserAgent()

	// Log the activity
	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"ip_address": clientIP,
		"user_agent": userAgent,
		"activity":   activity,
		"timestamp":  time.Now().Format(time.RFC3339),
	}).Info("User activity")
}

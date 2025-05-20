package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	UserID      uint       `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username    string     `gorm:"size:50;not null;unique" json:"username"`
	Email       string     `gorm:"size:100;not null;unique" json:"email"`
	Password    string     `gorm:"size:255;not null" json:"-"` // Password is not exposed in JSON
	Role        string     `gorm:"size:50;default:member" json:"role"`
	IsLogin     bool       `gorm:"default:false" json:"is_login"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Heartbeat   *time.Time `json:"heartbeat"`
	Status      string     `gorm:"type:enum('online','offline','away');default:'offline'" json:"status"`
	AvatarURL   string     `gorm:"size:255" json:"avatar_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BeforeSave is a GORM hook that hashes the password before saving
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" && !isHashedPassword(u.Password) {
		// Use a higher cost factor for better security (12 is a good balance between security and performance)
		cost := 12

		// Generate a salt and hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), cost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
		// bcrypt already includes the salt in the hash
	}
	return nil
}

// isHashedPassword checks if a password is already hashed with bcrypt
// bcrypt hashes start with $2a$, $2b$, or $2y$
func isHashedPassword(password string) bool {
	return len(password) > 4 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2y$")
}

// BeforeCreate is a GORM hook that sets the timestamps before creating a record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate is a GORM hook that sets the updated_at timestamp before updating a record
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// UserResponse is a struct for returning user data without sensitive information
type UserResponse struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		UserID:    u.UserID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}

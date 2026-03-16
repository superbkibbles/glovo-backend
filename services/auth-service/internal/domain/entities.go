package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the auth system (Postgres)
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:text"` // nullable for OTP/Google users
	Salt         string    `gorm:"type:text"` // nullable for OTP/Google users
	Name         string
	Email        string `gorm:"index"`
	PhoneNumber  string `gorm:"index"`
	GoogleID     string `gorm:"index;column:google_id"`
	ProfilePhoto string
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	UserType     string    `gorm:"type:varchar(50)"` // customer, restaurant_owner, driver, admin
	IsSuperadmin bool      `gorm:"default:false"`
	Active       bool      `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Icon        string
	Permissions StringArray `gorm:"type:text"` // JSON array stored as text
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StringArray for PostgreSQL compatibility
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for StringArray")
	}
	return json.Unmarshal(b, s)
}

// Session represents a user session (stored in Redis)
type Session struct {
	UserID       string    `json:"user_id"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetDisplayName returns the display name
func (u *User) GetDisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Username
}

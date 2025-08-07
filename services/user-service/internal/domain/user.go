package domain

import (
	"time"

	"glovo-backend/shared/auth"
)

// User represents the domain entity
type User struct {
	ID          string        `json:"id" gorm:"primaryKey"`
	PhoneNumber string        `json:"phone_number" gorm:"uniqueIndex"`
	Email       string        `json:"email,omitempty"`
	Role        auth.UserRole `json:"role"`
	Status      UserStatus    `json:"status"`
	Profile     UserProfile   `json:"profile" gorm:"embedded"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type UserProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar,omitempty"`
}

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusPending   UserStatus = "pending"
)

// OTP represents the OTP entity for phone verification
type OTP struct {
	PhoneNumber string    `json:"phone_number" gorm:"primaryKey"`
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expires_at"`
	Verified    bool      `json:"verified"`
	CreatedAt   time.Time `json:"created_at"`
}

// LoginRequest represents login request data
type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// VerifyOTPRequest represents OTP verification request
type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTPCode     string `json:"otp_code" binding:"required"`
}

// AdminLoginRequest represents admin login via email/password
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Repository interfaces (ports)
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByPhoneNumber(phoneNumber string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(limit, offset int) ([]User, error)
}

type OTPRepository interface {
	Store(otp *OTP) error
	GetByPhoneNumber(phoneNumber string) (*OTP, error)
	Delete(phoneNumber string) error
}

// Service interfaces (ports)
type UserService interface {
	SendOTP(phoneNumber string) error
	VerifyOTP(phoneNumber, otpCode string) (*LoginResponse, error)
	AdminLogin(email, password string) (*LoginResponse, error)
	GetProfile(userID string) (*User, error)
	UpdateProfile(userID string, profile UserProfile) (*User, error)
	ListUsers(limit, offset int) ([]User, error)
	SuspendUser(userID string) error
	ReactivateUser(userID string) error
}

type SMSService interface {
	SendSMS(phoneNumber, message string) error
}

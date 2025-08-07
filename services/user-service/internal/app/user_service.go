package app

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"glovo-backend/services/user-service/internal/domain"
	"glovo-backend/shared/auth"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo   domain.UserRepository
	otpRepo    domain.OTPRepository
	smsService domain.SMSService
}

func NewUserService(userRepo domain.UserRepository, otpRepo domain.OTPRepository, smsService domain.SMSService) domain.UserService {
	return &userService{
		userRepo:   userRepo,
		otpRepo:    otpRepo,
		smsService: smsService,
	}
}

func (s *userService) SendOTP(phoneNumber string) error {
	// Generate 6-digit OTP
	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Create OTP record
	otp := &domain.OTP{
		PhoneNumber: phoneNumber,
		Code:        otpCode,
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5 minutes expiry
		Verified:    false,
		CreatedAt:   time.Now(),
	}

	// Store OTP
	if err := s.otpRepo.Store(otp); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send SMS
	message := fmt.Sprintf("Your Glovo verification code is: %s", otpCode)
	if err := s.smsService.SendSMS(phoneNumber, message); err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}

func (s *userService) VerifyOTP(phoneNumber, otpCode string) (*domain.LoginResponse, error) {
	// Get OTP from storage
	storedOTP, err := s.otpRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return nil, errors.New("invalid or expired OTP")
	}

	// Check if OTP is expired
	if time.Now().After(storedOTP.ExpiresAt) {
		s.otpRepo.Delete(phoneNumber)
		return nil, errors.New("OTP has expired")
	}

	// Verify OTP code
	if storedOTP.Code != otpCode {
		return nil, errors.New("invalid OTP code")
	}

	// Delete OTP after successful verification
	s.otpRepo.Delete(phoneNumber)

	// Get or create user
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		// User doesn't exist, create new user
		user = &domain.User{
			ID:          uuid.New().String(),
			PhoneNumber: phoneNumber,
			Role:        auth.RoleCustomer, // Default role
			Status:      domain.StatusActive,
			Profile:     domain.UserProfile{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *userService) AdminLogin(email, password string) (*domain.LoginResponse, error) {
	// Get admin user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is admin
	if user.Role != auth.RoleAdmin {
		return nil, errors.New("access denied")
	}

	// Note: In a real implementation, you'd store hashed passwords
	// For now, we'll use a simple check
	expectedPassword := "admin123" // This should be hashed and stored in DB
	if password != expectedPassword {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *userService) GetProfile(userID string) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *userService) UpdateProfile(userID string, profile domain.UserProfile) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	user.Profile = profile
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ListUsers(limit, offset int) ([]domain.User, error) {
	return s.userRepo.List(limit, offset)
}

func (s *userService) SuspendUser(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.Status = domain.StatusSuspended
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

func (s *userService) ReactivateUser(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.Status = domain.StatusActive
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// Helper function to hash passwords (for future use)
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Helper function to check password hash (for future use)
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

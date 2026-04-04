package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/middleware"
	"github.com/mendmzury/food-delivery/pkg/utils"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/repository"
)

type AuthService struct {
	config      *config.Config
	userRepo    domain.UserRepository
	roleRepo    domain.RoleRepository
	sessionRepo domain.SessionRepository
	otpRepo     OTPRepository
	jwtManager  *utils.JWTManager
	pwHasher    *utils.PasswordHasher
}

func NewAuthService(
	cfg *config.Config,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	sessionRepo domain.SessionRepository,
	otpRepo OTPRepository,
) *AuthService {
	return &AuthService{
		config:      cfg,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		sessionRepo: sessionRepo,
		otpRepo:     otpRepo,
		jwtManager:  utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry),
		pwHasher:    utils.NewPasswordHasher(cfg.PasswordPepper),
	}
}

type LoginResult struct {
	Token        string
	RefreshToken string
	ExpiresIn    int64
	User         *UserInfo
	IsNewUser    bool
}

type UserInfo struct {
	ID           string
	Username     string
	Name         string
	CompanyID    string
	CompanyName  string
	RoleID       string
	RoleName     string
	IsSuperadmin bool
	Permissions  []string
	UserType     string
	Email        string
	PhoneNumber  string
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !user.Active {
		return nil, fmt.Errorf("account is deactivated")
	}

	if user.PasswordHash == "" || user.Salt == "" {
		return nil, fmt.Errorf("this account uses OTP signin, please use phone or email")
	}

	valid, err := s.pwHasher.VerifyPassword(password, user.Salt, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid username or password")
	}

	var permissions []string
	var roleName string
	if role, err := s.roleRepo.GetByID(ctx, user.RoleID); err == nil && role != nil {
		roleName = role.Name
		permissions = []string(role.Permissions)
	}

	token, err := s.jwtManager.GenerateToken(
		user.ID.String(),
		user.Username,
		"",
		user.RoleID.String(),
		user.IsSuperadmin,
		permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshExpiry := s.config.JWTRefreshExpiry
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}
	refreshToken, err := s.jwtManager.GenerateTokenWithExpiry(
		user.ID.String(),
		user.Username,
		"",
		user.RoleID.String(),
		user.IsSuperadmin,
		permissions,
		refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	sessionTTL := refreshExpiry
	session := &domain.Session{
		UserID:       user.ID.String(),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sessionTTL),
		CreatedAt:    time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session, sessionTTL); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResult{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTExpiry.Seconds()),
		User: &UserInfo{
			ID:           user.ID.String(),
			Username:     user.Username,
			Name:         user.GetDisplayName(),
			RoleID:       user.RoleID.String(),
			RoleName:     roleName,
			IsSuperadmin: user.IsSuperadmin,
			Permissions:  permissions,
			UserType:     user.UserType,
		},
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	exists, err := s.sessionRepo.Exists(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("session not found or expired")
	}

	return &UserInfo{
		ID:           claims.UserID,
		Username:     claims.Username,
		CompanyID:    claims.CompanyID,
		RoleID:       claims.RoleID,
		IsSuperadmin: claims.IsSuperadmin,
		Permissions:  claims.Permissions,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	valid, err := s.pwHasher.VerifyPassword(oldPassword, user.Salt, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return fmt.Errorf("current password is incorrect")
	}

	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	salt, err := s.pwHasher.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	hash, err := s.pwHasher.HashPassword(newPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, id, hash, salt); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	_ = s.sessionRepo.DeleteByUserID(ctx, userID)
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, userID, newPassword string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	salt, err := s.pwHasher.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	hash, err := s.pwHasher.HashPassword(newPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, id, hash, salt); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	_ = s.sessionRepo.DeleteByUserID(ctx, userID)
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, username string) (string, error) {
	_ = username
	return s.getSuperadminPhone(ctx)
}

func (s *AuthService) getSuperadminPhone(ctx context.Context) (string, error) {
	superadmin, err := s.userRepo.GetSuperadmin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get superadmin: %w", err)
	}
	if superadmin == nil {
		return "", fmt.Errorf("superadmin not found")
	}
	return superadmin.PhoneNumber, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResult, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.Active {
		return nil, fmt.Errorf("account is deactivated")
	}

	var permissions []string
	var roleName string
	if role, err := s.roleRepo.GetByID(ctx, user.RoleID); err == nil && role != nil {
		roleName = role.Name
		permissions = []string(role.Permissions)
	}

	token, err := s.jwtManager.GenerateToken(
		claims.UserID,
		claims.Username,
		"",
		claims.RoleID,
		claims.IsSuperadmin,
		claims.Permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshExpiry := s.config.JWTRefreshExpiry
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}
	newRefreshToken, err := s.jwtManager.GenerateTokenWithExpiry(
		claims.UserID,
		claims.Username,
		"",
		claims.RoleID,
		claims.IsSuperadmin,
		claims.Permissions,
		refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	sessionTTL := refreshExpiry
	session := &domain.Session{
		UserID:       claims.UserID,
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(sessionTTL),
		CreatedAt:    time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session, sessionTTL); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResult{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.JWTExpiry.Seconds()),
		User: &UserInfo{
			ID:           user.ID.String(),
			Username:     user.Username,
			Name:         user.GetDisplayName(),
			RoleID:       user.RoleID.String(),
			RoleName:     roleName,
			IsSuperadmin: user.IsSuperadmin,
			Permissions:  permissions,
			UserType:     user.UserType,
		},
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.Delete(ctx, token)
}

// buildLoginResult creates LoginResult from user
func (s *AuthService) buildLoginResult(ctx context.Context, user *domain.User) (*LoginResult, error) {
	var permissions []string
	var roleName string
	if role, err := s.roleRepo.GetByID(ctx, user.RoleID); err == nil && role != nil {
		roleName = role.Name
		permissions = []string(role.Permissions)
	}

	token, err := s.jwtManager.GenerateToken(
		user.ID.String(),
		user.Username,
		"",
		user.RoleID.String(),
		user.IsSuperadmin,
		permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshExpiry := s.config.JWTRefreshExpiry
	if refreshExpiry == 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}
	refreshToken, err := s.jwtManager.GenerateTokenWithExpiry(
		user.ID.String(),
		user.Username,
		"",
		user.RoleID.String(),
		user.IsSuperadmin,
		permissions,
		refreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	session := &domain.Session{
		UserID:       user.ID.String(),
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(refreshExpiry),
		CreatedAt:    time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session, refreshExpiry); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResult{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTExpiry.Seconds()),
		User: &UserInfo{
			ID:           user.ID.String(),
			Username:     user.Username,
			Name:         user.GetDisplayName(),
			RoleID:       user.RoleID.String(),
			RoleName:     roleName,
			IsSuperadmin: user.IsSuperadmin,
			Permissions:  permissions,
			UserType:     user.UserType,
			Email:        user.Email,
			PhoneNumber:  user.PhoneNumber,
		},
	}, nil
}

// getCustomerRole returns the Customer role, creating it if needed
func (s *AuthService) getCustomerRole(ctx context.Context) (*domain.Role, error) {
	role, err := s.roleRepo.GetByName(ctx, "Customer")
	if err == nil && role != nil {
		return role, nil
	}
	role = &domain.Role{
		ID:          uuid.New(),
		Name:        "Customer",
		Icon:        "user",
		Permissions: domain.StringArray(middleware.CustomerPermissions()),
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create Customer role: %w", err)
	}
	return role, nil
}

// SendOTP sends OTP to phone or email
func (s *AuthService) SendOTP(ctx context.Context, phone, email string) error {
	if phone == "" && email == "" {
		return fmt.Errorf("phone_number or email is required")
	}
	if phone != "" && email != "" {
		return fmt.Errorf("provide either phone_number or email, not both")
	}

	code, err := repository.GenerateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if phone != "" {
		if err := s.otpRepo.StorePhoneOTP(ctx, phone, code); err != nil {
			return fmt.Errorf("failed to store OTP: %w", err)
		}
		return s.otpRepo.SendPhoneOTP(ctx, phone, code)
	}

	if err := s.otpRepo.StoreEmailOTP(ctx, email, code); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}
	return s.otpRepo.SendEmailOTP(ctx, email, code)
}

// SignUpWithPhone creates or finds user by phone after OTP verification
func (s *AuthService) SignUpWithPhone(ctx context.Context, phone, otp string) (*LoginResult, error) {
	if phone == "" || otp == "" {
		return nil, fmt.Errorf("phone_number and otp are required")
	}

	valid, err := s.otpRepo.VerifyAndConsumePhoneOTP(ctx, phone, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	isNewUser := false
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		isNewUser = true
		customerRole, err := s.getCustomerRole(ctx)
		if err != nil {
			return nil, err
		}
		username := "phone:" + phone
		user = &domain.User{
			ID:          uuid.New(),
			Username:    username,
			PhoneNumber: phone,
			UserType:    "customer",
			RoleID:      customerRole.ID,
			Active:      true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	if !user.Active {
		return nil, fmt.Errorf("account is deactivated")
	}

	result, err := s.buildLoginResult(ctx, user)
	if err != nil {
		return nil, err
	}
	result.IsNewUser = isNewUser
	return result, nil
}

// SignInWithPhone signs in user by phone after OTP verification
func (s *AuthService) SignInWithPhone(ctx context.Context, phone, otp string) (*LoginResult, error) {
	return s.SignUpWithPhone(ctx, phone, otp)
}

// SignUpWithEmail creates or finds user by email after OTP verification
func (s *AuthService) SignUpWithEmail(ctx context.Context, email, otp, name string) (*LoginResult, error) {
	if email == "" || otp == "" {
		return nil, fmt.Errorf("email and otp are required")
	}

	valid, err := s.otpRepo.VerifyAndConsumeEmailOTP(ctx, email, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	isNewUser := false
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		isNewUser = true
		customerRole, err := s.getCustomerRole(ctx)
		if err != nil {
			return nil, err
		}
		username := "email:" + email
		user = &domain.User{
			ID:          uuid.New(),
			Username:    username,
			Email:       email,
			Name:        name,
			UserType:    "customer",
			RoleID:      customerRole.ID,
			Active:      true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	if !user.Active {
		return nil, fmt.Errorf("account is deactivated")
	}

	result, err := s.buildLoginResult(ctx, user)
	if err != nil {
		return nil, err
	}
	result.IsNewUser = isNewUser
	return result, nil
}

// SignInWithEmail signs in user by email after OTP verification
func (s *AuthService) SignInWithEmail(ctx context.Context, email, otp string) (*LoginResult, error) {
	return s.SignUpWithEmail(ctx, email, otp, "")
}

// SignUpWithGoogle creates or finds user by Google ID token
func (s *AuthService) SignUpWithGoogle(ctx context.Context, idToken string) (*LoginResult, error) {
	if idToken == "" {
		return nil, fmt.Errorf("id_token is required")
	}

	// Verify Firebase ID token and extract claims
	claims, err := verifyFirebaseIDToken(ctx, s.config.FirebaseCredentialsPath, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid Google token: %w", err)
	}

	googleID := claims.Subject
	if googleID == "" {
		return nil, fmt.Errorf("invalid token: missing subject")
	}

	user, err := s.userRepo.GetByGoogleID(ctx, googleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	isNewUser := false
	if user == nil {
		isNewUser = true
		customerRole, err := s.getCustomerRole(ctx)
		if err != nil {
			return nil, err
		}
		username := "google:" + googleID
		user = &domain.User{
			ID:       uuid.New(),
			Username: username,
			GoogleID: googleID,
			Name:     claims.Name,
			Email:    claims.Email,
			UserType: "customer",
			RoleID:   customerRole.ID,
			Active:   true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	if !user.Active {
		return nil, fmt.Errorf("account is deactivated")
	}

	result, err := s.buildLoginResult(ctx, user)
	if err != nil {
		return nil, err
	}
	result.IsNewUser = isNewUser
	return result, nil
}

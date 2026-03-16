package middleware

import (
	"context"
	"strings"

	"github.com/mendmzury/food-delivery/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is a server interceptor for authentication
type AuthInterceptor struct {
	jwtManager *utils.JWTManager
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(jwtSecret string) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: utils.NewJWTManager(jwtSecret, 0),
	}
}

func (i *AuthInterceptor) validateToken(tokenString string) (*Claims, error) {
	jwtClaims, err := i.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
	}

	return &Claims{
		UserID:       jwtClaims.UserID,
		Username:     jwtClaims.Username,
		CompanyID:    jwtClaims.CompanyID,
		RoleID:       jwtClaims.RoleID,
		IsSuperadmin: jwtClaims.IsSuperadmin,
		Permissions:  jwtClaims.Permissions,
	}, nil
}

type ContextKey string

const (
	UserIDKey       ContextKey = "user_id"
	UsernameKey     ContextKey = "username"
	CompanyIDKey    ContextKey = "company_id"
	RoleIDKey       ContextKey = "role_id"
	IsSuperadminKey ContextKey = "is_superadmin"
	PermissionsKey  ContextKey = "permissions"
)

// Claims represents JWT claims
type Claims struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	CompanyID    string   `json:"company_id,omitempty"`
	RoleID       string   `json:"role_id"`
	IsSuperadmin bool     `json:"is_superadmin"`
	Permissions  []string `json:"permissions"`
}

// UnaryServerInterceptor returns a unary server interceptor for authentication
func (i *AuthInterceptor) UnaryServerInterceptor(publicMethods []string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		for _, method := range publicMethods {
			if strings.HasSuffix(info.FullMethod, method) {
				return handler(ctx, req)
			}
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		claims, err := i.validateToken(token)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, CompanyIDKey, claims.CompanyID)
		ctx = context.WithValue(ctx, RoleIDKey, claims.RoleID)
		ctx = context.WithValue(ctx, IsSuperadminKey, claims.IsSuperadmin)
		ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a stream server interceptor for authentication
func (i *AuthInterceptor) StreamServerInterceptor(publicMethods []string) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		for _, method := range publicMethods {
			if strings.HasSuffix(info.FullMethod, method) {
				return handler(srv, ss)
			}
		}

		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			return status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		_, err := i.validateToken(token)
		if err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

// GetUserFromContext retrieves user info from context
func GetUserFromContext(ctx context.Context) (*Claims, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	claims := &Claims{UserID: userID}
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		claims.Username = username
	}
	if companyID, ok := ctx.Value(CompanyIDKey).(string); ok {
		claims.CompanyID = companyID
	}
	if roleID, ok := ctx.Value(RoleIDKey).(string); ok {
		claims.RoleID = roleID
	}
	if isSuperadmin, ok := ctx.Value(IsSuperadminKey).(bool); ok {
		claims.IsSuperadmin = isSuperadmin
	}
	if permissions, ok := ctx.Value(PermissionsKey).([]string); ok {
		claims.Permissions = permissions
	}

	return claims, nil
}

// SetUserInContext sets user info in context
func SetUserInContext(ctx context.Context, claims *Claims) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	ctx = context.WithValue(ctx, UsernameKey, claims.Username)
	ctx = context.WithValue(ctx, CompanyIDKey, claims.CompanyID)
	ctx = context.WithValue(ctx, RoleIDKey, claims.RoleID)
	ctx = context.WithValue(ctx, IsSuperadminKey, claims.IsSuperadmin)
	ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)
	return ctx
}

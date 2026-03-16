package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeAlreadyExists  = "ALREADY_EXISTS"
)

// AppError represents an application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code, message, details string) *AppError {
	return &AppError{Code: code, Message: message, Details: details}
}

// ToGRPCError converts an AppError to a gRPC status error
func (e *AppError) ToGRPCError() error {
	var code codes.Code
	switch e.Code {
	case ErrCodeNotFound:
		code = codes.NotFound
	case ErrCodeInvalidInput, ErrCodeValidation:
		code = codes.InvalidArgument
	case ErrCodeUnauthorized:
		code = codes.Unauthenticated
	case ErrCodeForbidden:
		code = codes.PermissionDenied
	case ErrCodeConflict, ErrCodeAlreadyExists:
		code = codes.AlreadyExists
	default:
		code = codes.Internal
	}
	return status.Error(code, e.Message)
}

func GRPCErrorNotFound(resource string) error {
	return status.Errorf(codes.NotFound, "%s not found", resource)
}

func GRPCErrorInvalidInput(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

func GRPCErrorUnauthorized(message string) error {
	return status.Error(codes.Unauthenticated, message)
}

func GRPCErrorForbidden(message string) error {
	return status.Error(codes.PermissionDenied, message)
}

func GRPCErrorAlreadyExists(resource string) error {
	return status.Errorf(codes.AlreadyExists, "%s already exists", resource)
}

func GRPCErrorInternal(message string) error {
	return status.Error(codes.Internal, message)
}

func IsGRPCError(err error, code codes.Code) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == code
}

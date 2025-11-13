package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	authv1 "github.com/raibid-labs/mop/examples/02-grpc-service/proto/auth/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthService implements the gRPC AuthService
type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	sessions *SessionStore
	tokens   *TokenManager
	logger   *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(logger *zap.Logger) *AuthService {
	return &AuthService{
		sessions: NewSessionStore(),
		tokens:   NewTokenManager(),
		logger:   logger,
	}
}

// Login handles user authentication
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	s.logger.Info("login attempt", zap.String("username", req.Username))

	// Validate request
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}

	// Simple authentication - in production, check against a database
	// For demo purposes, we accept any username with password "password"
	if req.Password != "password" {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Generate user ID and tokens
	userID := uuid.New().String()
	token, tokenExpiry := s.tokens.GenerateToken(userID)
	refreshToken, _ := s.tokens.GenerateRefreshToken(userID)

	// Create session
	email := fmt.Sprintf("%s@example.com", req.Username)
	roles := []string{"user"}
	s.sessions.CreateSession(userID, req.Username, email, roles)

	// Build response
	user := &authv1.User{
		Id:       userID,
		Username: req.Username,
		Email:    email,
		Roles:    roles,
	}

	s.logger.Info("login successful", zap.String("user_id", userID))

	return &authv1.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    timestamppb.New(tokenExpiry),
		User:         user,
	}, nil
}

// Logout handles user logout
func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	s.logger.Info("logout attempt")

	// Validate token
	userID, valid := s.tokens.ValidateToken(req.Token)
	if !valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Revoke token and delete session
	s.tokens.RevokeToken(req.Token)
	s.sessions.DeleteSession(userID)

	s.logger.Info("logout successful", zap.String("user_id", userID))

	return &authv1.LogoutResponse{
		Success: true,
	}, nil
}

// ValidateToken checks if a token is valid
func (s *AuthService) ValidateToken(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	s.logger.Info("validate token attempt")

	// Validate token
	userID, valid := s.tokens.ValidateToken(req.Token)
	if !valid {
		return &authv1.ValidateResponse{
			Valid: false,
		}, nil
	}

	// Get session
	session, exists := s.sessions.GetSession(userID)
	if !exists {
		return &authv1.ValidateResponse{
			Valid: false,
		}, nil
	}

	// Build user
	user := &authv1.User{
		Id:       session.UserID,
		Username: session.Username,
		Email:    session.Email,
		Roles:    session.Roles,
	}

	s.logger.Info("token valid", zap.String("user_id", userID))

	return &authv1.ValidateResponse{
		Valid:     true,
		User:      user,
		ExpiresAt: timestamppb.New(time.Now().Add(1 * time.Hour)),
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	s.logger.Info("refresh token attempt")

	// Validate refresh token
	userID, valid := s.tokens.ValidateToken(req.RefreshToken)
	if !valid {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	if !s.tokens.IsRefreshToken(req.RefreshToken) {
		return nil, status.Error(codes.InvalidArgument, "not a refresh token")
	}

	// Generate new access token
	token, tokenExpiry := s.tokens.GenerateToken(userID)

	s.logger.Info("token refreshed", zap.String("user_id", userID))

	return &authv1.RefreshResponse{
		Token:     token,
		ExpiresAt: timestamppb.New(tokenExpiry),
	}, nil
}

// StreamEvents sends authentication events to the client (server streaming)
func (s *AuthService) StreamEvents(req *authv1.EventsRequest, stream authv1.AuthService_StreamEventsServer) error {
	s.logger.Info("stream events started", zap.Strings("event_types", req.EventTypes))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	eventCount := 0
	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("stream events ended", zap.Int("events_sent", eventCount))
			return nil
		case <-ticker.C:
			event := &authv1.Event{
				EventType: "user_activity",
				UserId:    uuid.New().String(),
				Timestamp: timestamppb.Now(),
				Metadata: map[string]string{
					"action": "login_attempt",
					"source": "web",
				},
			}

			if err := stream.Send(event); err != nil {
				s.logger.Error("failed to send event", zap.Error(err))
				return status.Error(codes.Internal, "failed to send event")
			}

			eventCount++
			s.logger.Debug("event sent", zap.Int("count", eventCount))
		}
	}
}

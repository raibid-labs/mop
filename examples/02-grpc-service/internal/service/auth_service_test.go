package service

import (
	"context"
	"testing"
	"time"

	authv1 "github.com/raibid-labs/mop/examples/02-grpc-service/proto/auth/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthService_Login(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewAuthService(logger)

	tests := []struct {
		name        string
		username    string
		password    string
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password",
			wantErr:  false,
		},
		{
			name:        "empty username",
			username:    "",
			password:    "password",
			wantErr:     true,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "empty password",
			username:    "testuser",
			password:    "",
			wantErr:     true,
			expectedErr: codes.InvalidArgument,
		},
		{
			name:        "wrong password",
			username:    "testuser",
			password:    "wrongpassword",
			wantErr:     true,
			expectedErr: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Login(context.Background(), &authv1.LoginRequest{
				Username: tt.username,
				Password: tt.password,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.expectedErr {
						t.Errorf("expected error code %v, got %v", tt.expectedErr, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.Token == "" {
				t.Error("expected token, got empty string")
			}
			if resp.RefreshToken == "" {
				t.Error("expected refresh token, got empty string")
			}
			if resp.User == nil {
				t.Error("expected user, got nil")
			}
			if resp.User.Username != tt.username {
				t.Errorf("expected username %s, got %s", tt.username, resp.User.Username)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewAuthService(logger)

	// First login to get a token
	loginResp, err := service.Login(context.Background(), &authv1.LoginRequest{
		Username: "testuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:    "successful logout",
			token:   loginResp.Token,
			wantErr: false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			wantErr:     true,
			expectedErr: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Logout(context.Background(), &authv1.LogoutRequest{
				Token: tt.token,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.expectedErr {
						t.Errorf("expected error code %v, got %v", tt.expectedErr, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !resp.Success {
				t.Error("expected success=true, got false")
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewAuthService(logger)

	// First login to get a token
	loginResp, err := service.Login(context.Background(), &authv1.LoginRequest{
		Username: "testuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{
			name:  "valid token",
			token: loginResp.Token,
			valid: true,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.ValidateToken(context.Background(), &authv1.ValidateRequest{
				Token: tt.token,
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.Valid != tt.valid {
				t.Errorf("expected valid=%v, got %v", tt.valid, resp.Valid)
			}

			if tt.valid && resp.User == nil {
				t.Error("expected user for valid token, got nil")
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewAuthService(logger)

	// First login to get tokens
	loginResp, err := service.Login(context.Background(), &authv1.LoginRequest{
		Username: "testuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		expectedErr codes.Code
	}{
		{
			name:    "successful refresh",
			token:   loginResp.RefreshToken,
			wantErr: false,
		},
		{
			name:        "invalid refresh token",
			token:       "invalid-token",
			wantErr:     true,
			expectedErr: codes.Unauthenticated,
		},
		{
			name:        "using access token instead of refresh token",
			token:       loginResp.Token,
			wantErr:     true,
			expectedErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.RefreshToken(context.Background(), &authv1.RefreshRequest{
				RefreshToken: tt.token,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if st, ok := status.FromError(err); ok {
					if st.Code() != tt.expectedErr {
						t.Errorf("expected error code %v, got %v", tt.expectedErr, st.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.Token == "" {
				t.Error("expected new token, got empty string")
			}
		})
	}
}

func TestAuthService_StreamEvents(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewAuthService(logger)

	// Create a mock stream
	mockStream := &mockStreamEventsServer{
		ctx:    context.Background(),
		events: make([]*authv1.Event, 0),
	}

	// Start streaming in a goroutine
	done := make(chan error)
	go func() {
		err := service.StreamEvents(&authv1.EventsRequest{
			EventTypes: []string{"user_activity"},
		}, mockStream)
		done <- err
	}()

	// Wait for a few events to be sent
	time.Sleep(6 * time.Second)

	// Cancel the stream
	mockStream.cancel()

	// Wait for stream to finish
	err := <-done
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify we received at least one event
	if len(mockStream.events) == 0 {
		t.Error("expected at least one event, got none")
	}

	// Verify event structure
	if len(mockStream.events) > 0 {
		event := mockStream.events[0]
		if event.EventType == "" {
			t.Error("expected event type, got empty string")
		}
		if event.UserId == "" {
			t.Error("expected user ID, got empty string")
		}
	}
}

// Mock stream for testing
type mockStreamEventsServer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	events   []*authv1.Event
	authv1.AuthService_StreamEventsServer
}

func (m *mockStreamEventsServer) Send(event *authv1.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockStreamEventsServer) Context() context.Context {
	if m.cancel == nil {
		m.ctx, m.cancel = context.WithCancel(m.ctx)
	}
	return m.ctx
}

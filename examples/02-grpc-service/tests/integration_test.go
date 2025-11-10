package tests

import (
	"context"
	"net"
	"testing"
	"time"

	authv1 "github.com/raibid-labs/mop/examples/02-grpc-service/proto/auth/v1"
	"github.com/raibid-labs/mop/examples/02-grpc-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	logger, _ := zap.NewDevelopment()

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, service.NewAuthService(logger))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func getTestClient(t *testing.T) (authv1.AuthServiceClient, func()) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	client := authv1.NewAuthServiceClient(conn)
	cleanup := func() { conn.Close() }

	return client, cleanup
}

func TestIntegration_LoginLogoutFlow(t *testing.T) {
	client, cleanup := getTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Login
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: "testuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if loginResp.Token == "" {
		t.Error("expected token, got empty string")
	}
	if loginResp.User.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", loginResp.User.Username)
	}

	// Logout
	logoutResp, err := client.Logout(ctx, &authv1.LogoutRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	if !logoutResp.Success {
		t.Error("expected logout success")
	}
}

func TestIntegration_TokenValidation(t *testing.T) {
	client, cleanup := getTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Login
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: "testuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Validate token
	validateResp, err := client.ValidateToken(ctx, &authv1.ValidateRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	if !validateResp.Valid {
		t.Error("expected token to be valid")
	}
	if validateResp.User.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", validateResp.User.Username)
	}
}

func TestIntegration_TokenRefresh(t *testing.T) {
	client, cleanup := getTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Login
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: "testuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	originalToken := loginResp.Token

	// Refresh token
	refreshResp, err := client.RefreshToken(ctx, &authv1.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if refreshResp.Token == "" {
		t.Error("expected new token, got empty string")
	}
	if refreshResp.Token == originalToken {
		t.Error("expected different token after refresh")
	}
}

func TestIntegration_StreamEvents(t *testing.T) {
	client, cleanup := getTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Start streaming
	stream, err := client.StreamEvents(ctx, &authv1.EventsRequest{
		EventTypes: []string{"user_activity"},
	})
	if err != nil {
		t.Fatalf("stream failed: %v", err)
	}

	// Receive at least one event
	event, err := stream.Recv()
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}

	if event.EventType == "" {
		t.Error("expected event type, got empty string")
	}
	if event.UserId == "" {
		t.Error("expected user ID, got empty string")
	}
}

func TestIntegration_FullWorkflow(t *testing.T) {
	client, cleanup := getTestClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Step 1: Login
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: "integrationuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Step 2: Validate token
	validateResp, err := client.ValidateToken(ctx, &authv1.ValidateRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !validateResp.Valid {
		t.Error("token should be valid")
	}

	// Step 3: Refresh token
	refreshResp, err := client.RefreshToken(ctx, &authv1.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshResp.Token == "" {
		t.Error("expected refreshed token")
	}

	// Step 4: Start event stream in goroutine
	streamCtx, streamCancel := context.WithTimeout(ctx, 6*time.Second)
	defer streamCancel()

	eventReceived := make(chan bool, 1)
	go func() {
		stream, err := client.StreamEvents(streamCtx, &authv1.EventsRequest{
			EventTypes: []string{"user_activity"},
		})
		if err != nil {
			return
		}

		_, err = stream.Recv()
		if err == nil {
			eventReceived <- true
		}
	}()

	// Wait for event or timeout
	select {
	case <-eventReceived:
		t.Log("Event stream working")
	case <-time.After(7 * time.Second):
		t.Log("Event stream timeout (expected in fast tests)")
	}

	// Step 5: Logout
	logoutResp, err := client.Logout(ctx, &authv1.LogoutRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if !logoutResp.Success {
		t.Error("logout should succeed")
	}

	// Step 6: Verify token is invalid after logout
	validateResp2, err := client.ValidateToken(ctx, &authv1.ValidateRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		t.Fatalf("validate after logout failed: %v", err)
	}
	if validateResp2.Valid {
		t.Error("token should be invalid after logout")
	}
}

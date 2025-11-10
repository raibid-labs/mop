package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	authv1 "github.com/raibid-labs/mop/examples/02-grpc-service/proto/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Parse command line flags
	address := flag.String("addr", "localhost:9090", "gRPC server address")
	username := flag.String("username", "testuser", "username for login")
	password := flag.String("password", "password", "password for login")
	action := flag.String("action", "full-flow", "action to perform: login, logout, validate, refresh, stream, full-flow")
	flag.Parse()

	// Connect to gRPC server
	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := authv1.NewAuthServiceClient(conn)

	switch *action {
	case "login":
		testLogin(client, *username, *password)
	case "logout":
		testLogout(client, *username, *password)
	case "validate":
		testValidate(client, *username, *password)
	case "refresh":
		testRefresh(client, *username, *password)
	case "stream":
		testStream(client)
	case "full-flow":
		testFullFlow(client, *username, *password)
	default:
		log.Fatalf("unknown action: %s", *action)
	}
}

func testLogin(client authv1.AuthServiceClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("Attempting login with username=%s\n", username)

	resp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Printf("Login successful!\n")
	fmt.Printf("Token: %s\n", resp.Token)
	fmt.Printf("User ID: %s\n", resp.User.Id)
	fmt.Printf("Email: %s\n", resp.User.Email)
	fmt.Printf("Roles: %v\n", resp.User.Roles)
	fmt.Printf("Expires At: %s\n", resp.ExpiresAt.AsTime())
}

func testLogout(client authv1.AuthServiceClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First login to get a token
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Printf("Logged in, now logging out...\n")

	// Logout
	logoutResp, err := client.Logout(ctx, &authv1.LogoutRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		log.Fatalf("logout failed: %v", err)
	}

	fmt.Printf("Logout successful: %v\n", logoutResp.Success)
}

func testValidate(client authv1.AuthServiceClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First login to get a token
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Printf("Validating token...\n")

	// Validate token
	validateResp, err := client.ValidateToken(ctx, &authv1.ValidateRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		log.Fatalf("validate failed: %v", err)
	}

	if validateResp.Valid {
		fmt.Printf("Token is valid!\n")
		fmt.Printf("User: %s (%s)\n", validateResp.User.Username, validateResp.User.Email)
	} else {
		fmt.Printf("Token is invalid\n")
	}
}

func testRefresh(client authv1.AuthServiceClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First login to get a refresh token
	loginResp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}

	fmt.Printf("Refreshing token...\n")

	// Refresh token
	refreshResp, err := client.RefreshToken(ctx, &authv1.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		log.Fatalf("refresh failed: %v", err)
	}

	fmt.Printf("New token: %s\n", refreshResp.Token)
	fmt.Printf("Expires At: %s\n", refreshResp.ExpiresAt.AsTime())
}

func testStream(client authv1.AuthServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Printf("Starting event stream...\n")

	stream, err := client.StreamEvents(ctx, &authv1.EventsRequest{
		EventTypes: []string{"user_activity", "login"},
	})
	if err != nil {
		log.Fatalf("stream failed: %v", err)
	}

	eventCount := 0
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Stream ended\n")
			break
		}
		if err != nil {
			log.Fatalf("receive failed: %v", err)
		}

		eventCount++
		fmt.Printf("Event %d: %s (user=%s, time=%s)\n",
			eventCount,
			event.EventType,
			event.UserId,
			event.Timestamp.AsTime(),
		)
	}
}

func testFullFlow(client authv1.AuthServiceClient, username, password string) {
	ctx := context.Background()

	fmt.Println("=== Full Flow Test ===")

	// Step 1: Login
	fmt.Println("1. Testing Login...")
	loginCtx, loginCancel := context.WithTimeout(ctx, 5*time.Second)
	defer loginCancel()

	loginResp, err := client.Login(loginCtx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("   ✓ Login successful (user_id=%s)\n\n", loginResp.User.Id)

	// Step 2: Validate Token
	fmt.Println("2. Testing Token Validation...")
	validateCtx, validateCancel := context.WithTimeout(ctx, 5*time.Second)
	defer validateCancel()

	validateResp, err := client.ValidateToken(validateCtx, &authv1.ValidateRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		log.Fatalf("validate failed: %v", err)
	}
	fmt.Printf("   ✓ Token valid=%v\n\n", validateResp.Valid)

	// Step 3: Refresh Token
	fmt.Println("3. Testing Token Refresh...")
	refreshCtx, refreshCancel := context.WithTimeout(ctx, 5*time.Second)
	defer refreshCancel()

	refreshResp, err := client.RefreshToken(refreshCtx, &authv1.RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		log.Fatalf("refresh failed: %v", err)
	}
	fmt.Printf("   ✓ Token refreshed (new_token=%s...)\n\n", refreshResp.Token[:8])

	// Step 4: Stream Events (for 10 seconds)
	fmt.Println("4. Testing Event Streaming (10 seconds)...")
	streamCtx, streamCancel := context.WithTimeout(ctx, 10*time.Second)
	defer streamCancel()

	stream, err := client.StreamEvents(streamCtx, &authv1.EventsRequest{
		EventTypes: []string{"user_activity"},
	})
	if err != nil {
		log.Fatalf("stream failed: %v", err)
	}

	eventCount := 0
	for {
		_, err := stream.Recv()
		if err == io.EOF || err != nil {
			break
		}
		eventCount++
	}
	fmt.Printf("   ✓ Received %d events\n\n", eventCount)

	// Step 5: Logout
	fmt.Println("5. Testing Logout...")
	logoutCtx, logoutCancel := context.WithTimeout(ctx, 5*time.Second)
	defer logoutCancel()

	logoutResp, err := client.Logout(logoutCtx, &authv1.LogoutRequest{
		Token: loginResp.Token,
	})
	if err != nil {
		log.Fatalf("logout failed: %v", err)
	}
	fmt.Printf("   ✓ Logout successful=%v\n\n", logoutResp.Success)

	fmt.Println("=== All Tests Passed ===")
}

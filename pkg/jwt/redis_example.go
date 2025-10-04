package jwt

import (
	"context"
	"fmt"
	"log"
	"time"

	"monorepo/pkg/redis"
)

// ExampleRedisUsage demonstrates how to use the existing pkg/redis package with JWT
func ExampleRedisUsage() {
	// Create Redis client using the existing pkg/redis package
	redisClient, err := redis.New(
		redis.WithAddrs([]string{"localhost:6379"}),
		redis.WithUsername("demo"),
		redis.WithPassword("demo"),
		redis.WithDB(0),
		redis.WithPoolSize(10),
	)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}

	// Create JWT store using the existing Redis client
	jwtStore := NewRedisStore(redisClient)

	// Create JWT manager with the existing Redis store
	jwtManager, err := NewStatefulWithRedis(redisClient,
		WithAccessTokenSecret("access-secret-key-existing"),
		WithRefreshTokenSecret("refresh-secret-key-existing"),
		WithAccessTokenExpiry(time.Minute*15),
		WithRefreshTokenExpiry(time.Hour*24*7),
		WithStateful(true),
	)
	if err != nil {
		log.Fatalf("Error creating JWT manager: %v", err)
	}

	// Generate tokens with session tracking
	accessToken, refreshToken, sessionID, err := jwtManager.GenerateTokensWithSession(
		context.TODO(), "user123", "agent123", "sub_agent",
		"Chrome on Windows 10", "103.23.141.22",
	)
	if err != nil {
		log.Fatalf("Error generating tokens with session: %v", err)
	}

	fmt.Printf("Session ID: %s\n", sessionID)

	fmt.Printf("Access Token: %s\n", accessToken)
	fmt.Printf("Refresh Token: %s\n", refreshToken)

	// Validate tokens
	claims, err := jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Fatalf("Error validating refresh token: %v", err)
	}

	fmt.Printf("Validated user ID: %s\n", claims.UserID)
	fmt.Printf("Token ID: %s\n", claims.ID)

	// Refresh access token
	newAccessToken, err := jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		log.Fatalf("Error refreshing access token: %v", err)
	}

	fmt.Printf("New Access Token: %s\n", newAccessToken)

	// Try to use the old refresh token again (should fail)
	_, err = jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		fmt.Printf("As expected, refresh token is no longer valid: %v\n", err)
	}

	// Demonstrate token revocation
	fmt.Println("\n=== Token Revocation Example ===")

	// Generate a new refresh token
	refreshToken2, err := jwtManager.GenerateRefreshToken("user456", "agent123", "user")
	if err != nil {
		log.Fatalf("Error generating refresh token: %v", err)
	}

	// Validate it first
	claims2, err := jwtManager.ValidateRefreshToken(refreshToken2)
	if err != nil {
		log.Fatalf("Error validating refresh token: %v", err)
	}

	fmt.Printf("Generated token for user: %s with token ID: %s\n", claims2.UserID, claims2.ID)

	// Revoke the token
	err = jwtManager.RevokeRefreshToken(claims2.UserID, claims2.ID)
	if err != nil {
		log.Fatalf("Error revoking refresh token: %v", err)
	}

	// Try to validate the revoked token
	_, err = jwtManager.ValidateRefreshToken(refreshToken2)
	if err != nil {
		fmt.Printf("As expected, revoked token is no longer valid: %v\n", err)
	}

	// Demonstrate revoking all tokens for a user
	fmt.Println("\n=== Revoke All Tokens Example ===")

	// Generate multiple tokens for the same user
	token1, err := jwtManager.GenerateRefreshToken("user789", "agent123", "user")
	if err != nil {
		log.Fatalf("Error generating refresh token 1: %v", err)
	}

	token2, err := jwtManager.GenerateRefreshToken("user789", "agent123", "user")
	if err != nil {
		log.Fatalf("Error generating refresh token 2: %v", err)
	}

	// Validate both tokens
	_, err = jwtManager.ValidateRefreshToken(token1)
	if err != nil {
		log.Fatalf("Error validating refresh token 1: %v", err)
	}

	_, err = jwtManager.ValidateRefreshToken(token2)
	if err != nil {
		log.Fatalf("Error validating refresh token 2: %v", err)
	}

	fmt.Println("Both tokens are valid")

	// Revoke all tokens for the user
	err = jwtManager.RevokeAllRefreshTokens("user789")
	if err != nil {
		log.Fatalf("Error revoking all refresh tokens: %v", err)
	}

	// Try to validate both tokens
	_, err = jwtManager.ValidateRefreshToken(token1)
	if err != nil {
		fmt.Printf("Token 1 is no longer valid: %v\n", err)
	}

	_, err = jwtManager.ValidateRefreshToken(token2)
	if err != nil {
		fmt.Printf("Token 2 is no longer valid: %v\n", err)
	}

	// Close Redis connection
	err = jwtStore.Close()
	if err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	fmt.Println("\nExisting Redis JWT example completed successfully!")
}

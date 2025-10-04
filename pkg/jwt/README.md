# JWT Package

This package provides JWT token functionality for access tokens and refresh tokens.

## Features

- Generate access tokens with configurable expiry
- Generate refresh tokens with configurable expiry
- Validate both access and refresh tokens
- Refresh access tokens using refresh tokens
- Type safety to prevent using wrong token types

## Usage

```go
package main

import (
    "fmt"
    "time"
    
    "monorepo/pkg/jwt"
)

func main() {
    // Initialize JWT manager with configuration
    jwtManager, err := jwt.New(
        jwt.WithAccessTokenSecret("your-access-secret-key"),
        jwt.WithRefreshTokenSecret("your-refresh-secret-key"),
        jwt.WithAccessTokenExpiry(time.Minute * 15),  // 15 minutes
        jwt.WithRefreshTokenExpiry(time.Hour * 24 * 7), // 7 days
    )
    if err != nil {
        panic(err)
    }

    // Generate access token
    accessToken, err := jwtManager.GenerateAccessToken("user123", "agent123", "user")
    if err != nil {
        fmt.Printf("Error generating access token: %v\n", err)
        return
    }

    // Generate refresh token
    refreshToken, err := jwtManager.GenerateRefreshToken("user123", "agent123", "user")
    if err != nil {
        fmt.Printf("Error generating refresh token: %v\n", err)
        return
    }

    fmt.Printf("Access Token: %s\n", accessToken)
    fmt.Printf("Refresh Token: %s\n", refreshToken)

    // Validate access token
    claims, err := jwtManager.ValidateAccessToken(accessToken)
    if err != nil {
        fmt.Printf("Error validating access token: %v\n", err)
        return
    }

    fmt.Printf("User ID: %s\n", claims.UserID)
    fmt.Printf("Agent ID: %s\n", claims.AgentID)
    fmt.Printf("Agent Type: %s\n", claims.AgentType)

    // Refresh access token using refresh token
    newAccessToken, err := jwtManager.RefreshAccessToken(refreshToken)
    if err != nil {
        fmt.Printf("Error refreshing access token: %v\n", err)
        return
    }

    fmt.Printf("New Access Token: %s\n", newAccessToken)

    // Check token expiration
    expiry, err := jwtManager.GetTokenExpiration(accessToken)
    if err != nil {
        fmt.Printf("Error getting token expiration: %v\n", err)
        return
    }

    remaining, err := jwtManager.GetTokenRemainingTime(accessToken)
    if err != nil {
        fmt.Printf("Error getting remaining time: %v\n", err)
        return
    }

    fmt.Printf("Token expires at: %s\n", expiry.Format(time.RFC3339))
    fmt.Printf("Time remaining: %v\n", remaining)
}
```

## Configuration

The `TokenConfig` struct contains:

- `AccessTokenSecret`: Secret key for signing access tokens
- `RefreshTokenSecret`: Secret key for signing refresh tokens
- `AccessTokenExpiry`: Duration for access token expiry
- `RefreshTokenExpiry`: Duration for refresh token expiry

## Token Claims

The `TokenClaims` struct contains:

- `UserID`: User identifier
- `AgentID`: Agent identifier
- `AgentType`: Type of agent
- `TokenType`: Either "access" or "refresh"
- Standard JWT registered claims

## Token Expiration Utilities

The JWT client provides utility methods for working with token expiration:

- `GetTokenExpiration(tokenString)` - Returns the expiration time of a token
- `GetTokenRemainingTime(tokenString)` - Returns remaining time until expiration
- `IsTokenExpired(tokenString)` - Checks if a token is expired
- `GetAccessTokenExpiry()` - Returns configured access token expiry duration
- `GetRefreshTokenExpiry()` - Returns configured refresh token expiry duration

```go
// Get expiration info
expiry, _ := jwtManager.GetTokenExpiration(accessToken)
remaining, _ := jwtManager.GetTokenRemainingTime(accessToken)
expired, _ := jwtManager.IsTokenExpired(accessToken)

// Get configured durations
accessExpiry := jwtManager.GetAccessTokenExpiry()
refreshExpiry := jwtManager.GetRefreshTokenExpiry()
```

## Security Considerations

- Use strong, unique secrets for access and refresh tokens
- Keep access token expiry short (e.g., 15 minutes)
- Keep refresh token expiry longer but not indefinite (e.g., 7 days)
- Store refresh tokens securely on the client side
- Implement proper token rotation in production

## Stateful vs Stateless Mode

The JWT package supports both stateful and stateless token management:

- **Stateless Mode (Stateful: false)**:
    - Traditional JWT approach where tokens are self-contained
    - No server-side storage required
    - Cannot revoke individual tokens before expiry
    - More scalable but less control over token lifecycle

- **Stateful Mode (Stateful: true)**:
    - Refresh tokens are stored and tracked in a store
    - Can revoke individual tokens or all tokens for a user
    - More secure as tokens can be immediately invalidated
    - Requires storage backend (like Redis) for production use
    - Refresh tokens are invalidated after single use

## Using Redis Store

For production use with stateful mode, use the existing `pkg/redis` package:

```go
package main

import (
	"monorepo/pkg/jwt"
	"monorepo/pkg/redis"
	"time"
)

func main() {
	// Create Redis client using existing pkg/redis package
	redisClient, err := redis.New(
		redis.WithAddrs([]string{"localhost:6379"}),
		redis.WithUsername("demo"),
		redis.WithPassword("demo"),
		redis.WithDB(0),
		redis.WithPoolSize(10),
	)
	if err != nil {
		panic(err)
	}

	// Create JWT manager with Redis for stateful mode
	jwtManager, err := jwt.NewStatefulWithRedis(redisClient,
		jwt.WithAccessTokenSecret("your-access-secret"),
		jwt.WithRefreshTokenSecret("your-refresh-secret"),
		jwt.WithAccessTokenExpiry(time.Minute*15),
		jwt.WithRefreshTokenExpiry(time.Hour*24*7),
		jwt.WithStateful(true),
	)
	if err != nil {
		panic(err)
	}

	// Use jwtManager for token operations
	// ... rest of your code
}
```

### Redis Store Features

- **Automatic expiry**: Tokens are stored with TTL (Time To Live) matching their expiry time
- **Connection reuse**: Leverages existing Redis connection management
- **Error handling**: Proper error handling for Redis connection issues
- **Key pattern**: Uses `refresh_token:{userID}:{tokenID}` pattern for easy management
- **Cleanup**: Automatic cleanup of expired tokens via Redis TTL

### Redis Configuration Options

The Redis store uses the existing `pkg/redis` configuration options:

- `Addrs`: Redis server addresses (e.g., `[]string{"localhost:6379"}`)
- `Password`: Redis password (empty string for no password)
- `DB`: Redis database number (0-15)
- `PoolSize`: Connection pool size (default: 10)
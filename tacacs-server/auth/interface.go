package auth

import "context"

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// AuthenticateUser authenticates a user with username and password
	AuthenticateUser(ctx context.Context, username, password string) (*UserInfo, error)
	
	// GetPrivilegeLevel returns the privilege level for the given roles
	GetPrivilegeLevel(roles []string) int
	
	// IsAuthorized checks if the user with given roles is authorized to execute the command
	IsAuthorized(roles []string, command string) bool
	
	// CleanupCache cleans up expired tokens from cache
	CleanupCache()
}

// UserInfo contains information about an authenticated user
type UserInfo struct {
	Username string
	Roles    []string
	Groups   []string
}
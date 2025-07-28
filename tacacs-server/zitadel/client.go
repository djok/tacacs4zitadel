package zitadel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"tacacs-zitadel-server/auth"
	"tacacs-zitadel-server/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type Client struct {
	httpClient   *http.Client
	config       *config.Config
	logger       *logrus.Logger
	tokenCache   map[string]*CachedToken
	cacheMutex   sync.RWMutex
	clientToken  *TokenResponse
	tokenExpiry  time.Time
	tokenMutex   sync.RWMutex
}

type CachedToken struct {
	Token  *TokenResponse
	Expiry time.Time
	Roles  []string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// UserInfo type removed - using auth.UserInfo instead

type ZitadelUserInfo struct {
	Sub               string   `json:"sub"`
	PreferredUsername string   `json:"preferred_username"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Roles             []string `json:"urn:zitadel:iam:org:project:roles"`
	Groups            []string `json:"groups"`
}

func NewClient(cfg *config.Config, logger *logrus.Logger) (*Client, error) {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:     cfg,
		logger:     logger,
		tokenCache: make(map[string]*CachedToken),
	}, nil
}

func (c *Client) getClientToken(ctx context.Context) (*TokenResponse, error) {
	c.tokenMutex.RLock()
	if c.clientToken != nil && time.Now().Before(c.tokenExpiry.Add(-30*time.Second)) {
		token := c.clientToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	if c.clientToken != nil && time.Now().Before(c.tokenExpiry.Add(-30*time.Second)) {
		return c.clientToken, nil
	}

	tokenURL := fmt.Sprintf("%s/oauth/v2/token", c.config.ZitadelURL)
	
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.config.ZitadelClientID)
	data.Set("client_secret", c.config.ZitadelClientSecret)
	data.Set("scope", "openid profile email urn:zitadel:iam:org:project:id:zitadel:aud")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get client token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	c.clientToken = &token
	c.tokenExpiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	
	return &token, nil
}

func (c *Client) AuthenticateUser(ctx context.Context, username, password string) (*auth.UserInfo, error) {
	cacheKey := fmt.Sprintf("%s:%s", username, password)
	
	c.cacheMutex.RLock()
	if cached, exists := c.tokenCache[cacheKey]; exists && time.Now().Before(cached.Expiry) {
		c.cacheMutex.RUnlock()
		return &auth.UserInfo{
			Username: username,
			Roles:    cached.Roles,
		}, nil
	}
	c.cacheMutex.RUnlock()

	// Get user token using Resource Owner Password Credentials flow
	token, err := c.authenticateWithPassword(ctx, username, password)
	if err != nil {
		c.logger.WithError(err).WithField("username", username).Debug("Authentication failed")
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get user info and extract roles
	userInfo, err := c.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		c.logger.WithError(err).WithField("username", username).Warn("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	roles := c.extractRolesFromToken(token.AccessToken)
	if len(roles) == 0 {
		roles = userInfo.Roles
	}

	result := &auth.UserInfo{
		Username: userInfo.PreferredUsername,
		Roles:    roles,
		Groups:   userInfo.Groups,
	}

	c.cacheMutex.Lock()
	c.tokenCache[cacheKey] = &CachedToken{
		Token:  token,
		Expiry: time.Now().Add(time.Duration(c.config.TokenCacheTimeout) * time.Second),
		Roles:  roles,
	}
	c.cacheMutex.Unlock()

	c.logger.WithFields(logrus.Fields{
		"username": username,
		"roles":    roles,
	}).Info("User authenticated successfully")

	return result, nil
}

func (c *Client) authenticateWithPassword(ctx context.Context, username, password string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", c.config.ZitadelURL)
	
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.config.ZitadelClientID)
	data.Set("client_secret", c.config.ZitadelClientSecret)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", "openid profile email urn:zitadel:iam:org:project:id:zitadel:aud")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}

func (c *Client) getUserInfo(ctx context.Context, accessToken string) (*ZitadelUserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/oidc/v1/userinfo", c.config.ZitadelURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo ZitadelUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &userInfo, nil
}

func (c *Client) extractRolesFromToken(accessToken string) []string {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		c.logger.WithError(err).Warn("Failed to parse access token")
		return []string{}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return []string{}
	}

	var roles []string

	// Zitadel roles are in a specific claim
	if rolesClaim, ok := claims["urn:zitadel:iam:org:project:roles"]; ok {
		if rolesMap, ok := rolesClaim.(map[string]interface{}); ok {
			for role := range rolesMap {
				roles = append(roles, role)
			}
		}
	}

	// Also check standard roles claim
	if rolesClaim, ok := claims["roles"]; ok {
		if rolesArray, ok := rolesClaim.([]interface{}); ok {
			for _, role := range rolesArray {
				if roleStr, ok := role.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	return roles
}

func (c *Client) GetPrivilegeLevel(roles []string) int {
	for _, role := range roles {
		switch strings.ToLower(role) {
		case "network-admin", "admin", "zitadel.admin":
			return 15
		case "network-user", "user", "zitadel.user":
			return 1
		case "network-readonly", "readonly", "viewer":
			return 0
		}
	}
	return 0
}

func (c *Client) IsAuthorized(roles []string, command string) bool {
	privilegeLevel := c.GetPrivilegeLevel(roles)
	
	if privilegeLevel >= 15 {
		return true
	}
	
	if privilegeLevel >= 1 {
		return true
	}
	
	if privilegeLevel >= 0 {
		readOnlyCommands := []string{
			"show", "ping", "traceroute", "telnet", "ssh",
		}
		
		for _, cmd := range readOnlyCommands {
			if strings.HasPrefix(strings.ToLower(command), cmd) {
				return true
			}
		}
	}
	
	return false
}

func (c *Client) CleanupCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	
	now := time.Now()
	for key, cached := range c.tokenCache {
		if now.After(cached.Expiry) {
			delete(c.tokenCache, key)
		}
	}
}
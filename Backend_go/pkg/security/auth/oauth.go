package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/config"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// OAuthStateStore manages OAuth state tokens to prevent CSRF
type OAuthStateStore struct {
	states map[string]stateData
	mu     sync.RWMutex
}

type stateData struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
	Provider  string
}

// TokenInfo represents the OAuth2 token info from the provider
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// UserInfo represents user information from OAuth providers
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	Provider      string
	Raw           map[string]interface{}
}

var (
	stateStore     *OAuthStateStore
	stateStoreOnce sync.Once
)

// GetStateStore returns the singleton OAuthStateStore
func GetStateStore() *OAuthStateStore {
	stateStoreOnce.Do(func() {
		stateStore = &OAuthStateStore{
			states: make(map[string]stateData),
		}
	})
	return stateStore
}

// GenerateState creates a new state token and stores it
func (s *OAuthStateStore) GenerateState(provider string) (string, error) {
	// Generate random state token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state with expiry (default 10 minutes)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = stateData{
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Provider:  provider,
	}

	return state, nil
}

// ValidateState checks if a state token is valid and removes it
func (s *OAuthStateStore) ValidateState(state, provider string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.states[state]
	if !exists {
		return false
	}

	// Check if state is expired
	if time.Now().After(data.ExpiresAt) {
		delete(s.states, state)
		return false
	}

	// Check if provider matches
	if data.Provider != provider {
		return false
	}

	// Valid state, remove it so it can't be reused
	delete(s.states, state)
	return true
}

// GetProviderFromState returns the provider associated with a state token
// without removing the state from the store
func (s *OAuthStateStore) GetProviderFromState(state string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.states[state]
	if !exists {
		return "", false
	}

	// Check if state is expired
	if time.Now().After(data.ExpiresAt) {
		return "", false
	}

	return data.Provider, true
}

// CleanupExpiredStates removes expired state tokens
func (s *OAuthStateStore) CleanupExpiredStates() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for state, data := range s.states {
		if now.After(data.ExpiresAt) {
			delete(s.states, state)
		}
	}
}

// OAuthService handles OAuth2 authentication
type OAuthService struct {
	providers map[string]*oauth2.Config
	cfg       *config.Config
}

// NewOAuthService creates a new OAuth service with configured providers
func NewOAuthService(cfg *config.Config) *OAuthService {
	service := &OAuthService{
		providers: make(map[string]*oauth2.Config),
		cfg:       cfg,
	}

	// Configure providers from config
	for name, providerCfg := range cfg.Auth.OAuth2Providers {
		service.providers[name] = &oauth2.Config{
			ClientID:     providerCfg.ClientID,
			ClientSecret: providerCfg.ClientSecret,
			RedirectURL:  providerCfg.RedirectURL,
			Scopes:       providerCfg.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  providerCfg.AuthURL,
				TokenURL: providerCfg.TokenURL,
			},
		}
	}

	// Start a goroutine to periodically clean up expired states
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			GetStateStore().CleanupExpiredStates()
		}
	}()

	return service
}

// GetAuthURL returns an OAuth2 authorization URL for the specified provider
func (s *OAuthService) GetAuthURL(provider string) (string, string, error) {
	config, exists := s.providers[provider]
	if !exists {
		return "", "", fmt.Errorf("unknown OAuth provider: %s", provider)
	}

	state, err := GetStateStore().GenerateState(provider)
	if err != nil {
		return "", "", err
	}

	return config.AuthCodeURL(state), state, nil
}

// Exchange exchanges an authorization code for a token
func (s *OAuthService) Exchange(ctx context.Context, provider, code string) (*oauth2.Token, error) {
	config, exists := s.providers[provider]
	if !exists {
		return nil, fmt.Errorf("unknown OAuth provider: %s", provider)
	}

	return config.Exchange(ctx, code)
}

// GetUserInfo fetches user information from the OAuth provider
func (s *OAuthService) GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*UserInfo, error) {
	config, exists := s.providers[provider]
	if !exists {
		return nil, fmt.Errorf("unknown OAuth provider: %s", provider)
	}

	client := config.Client(ctx, token)

	// Get provider config for UserInfo URL
	providerCfg, exists := s.cfg.Auth.OAuth2Providers[provider]
	if !exists || providerCfg.UserInfoURL == "" {
		return nil, errors.New("missing user info URL for provider")
	}

	resp, err := client.Get(providerCfg.UserInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s - %s", resp.Status, string(body))
	}

	// Parse the response
	var rawData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	// Extract common fields (this may vary by provider)
	userInfo := &UserInfo{
		Provider: provider,
		Raw:      rawData,
	}

	// Map common fields (providers might use different field names)
	if id, ok := rawData["id"].(string); ok {
		userInfo.ID = id
	} else if id, ok := rawData["sub"].(string); ok {
		userInfo.ID = id
	}

	if email, ok := rawData["email"].(string); ok {
		userInfo.Email = email
	}

	if verified, ok := rawData["verified_email"].(bool); ok {
		userInfo.VerifiedEmail = verified
	} else if verified, ok := rawData["email_verified"].(bool); ok {
		userInfo.VerifiedEmail = verified
	}

	if name, ok := rawData["name"].(string); ok {
		userInfo.Name = name
	}

	if givenName, ok := rawData["given_name"].(string); ok {
		userInfo.GivenName = givenName
	} else if givenName, ok := rawData["first_name"].(string); ok {
		userInfo.GivenName = givenName
	}

	if familyName, ok := rawData["family_name"].(string); ok {
		userInfo.FamilyName = familyName
	} else if familyName, ok := rawData["last_name"].(string); ok {
		userInfo.FamilyName = familyName
	}

	if picture, ok := rawData["picture"].(string); ok {
		userInfo.Picture = picture
	} else if picture, ok := rawData["avatar_url"].(string); ok {
		userInfo.Picture = picture
	}

	if locale, ok := rawData["locale"].(string); ok {
		userInfo.Locale = locale
	}

	return userInfo, nil
}

// GetProviders returns all configured OAuth2 providers
func (s *OAuthService) GetProviders() map[string]*oauth2.Config {
	return s.providers
}

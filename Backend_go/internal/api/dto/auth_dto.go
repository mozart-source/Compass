package dto

// OAuth2LoginRequest represents a request to initiate OAuth2 login
type OAuth2LoginRequest struct {
	Provider string `json:"provider" binding:"required"`
}

// OAuth2LoginResponse contains the URL to redirect the user to for OAuth2 authentication
type OAuth2LoginResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// OAuth2CallbackRequest represents data from an OAuth2 callback
type OAuth2CallbackRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// OAuth2CallbackResponse contains the token and user information after successful OAuth2 authentication
type OAuth2CallbackResponse struct {
	Token     string       `json:"token"`
	ExpiresAt int64        `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// ProviderInfo contains information about an OAuth2 provider
type ProviderInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	AuthURL     string   `json:"auth_url,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
}

// OAuth2ProvidersResponse contains information about available OAuth2 providers
type OAuth2ProvidersResponse struct {
	Providers []ProviderInfo `json:"providers"`
}

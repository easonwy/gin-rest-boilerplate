package auth

// LoginRequest defines the user login request structure
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse defines the user login response structure
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry time in seconds
}

// RefreshTokenRequest defines the refresh token request structure
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

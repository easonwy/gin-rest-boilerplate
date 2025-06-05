package auth

// LoginRequest defines the user login request structure
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse defines the user login response structure
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // Access token expiry time in seconds
}

// RefreshTokenRequest defines the refresh token request structure
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

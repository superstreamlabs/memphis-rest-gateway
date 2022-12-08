package models

type AuthSchema struct {
	Username               string `json:"username" validate:"required"`
	ConnectionToken        string `json:"connection_token" validate:"required"`
	TokenExpiryMins        int    `json:"token_expiry_in_minutes"`
	RefreshTokenExpiryMins int    `json:"refresh_token_expiry_in_minutes"`
}

type RefreshTokenSchema struct {
	TokenExpiryMins        int `json:"token_expiry_in_minutes"`
	RefreshTokenExpiryMins int `json:"refresh_token_expiry_in_minutes"`
}

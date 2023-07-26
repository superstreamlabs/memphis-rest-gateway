package models

type AuthSchema struct {
	Username               string `json:"username" validate:"required"`
	ConnectionToken        string `json:"connection_token"`
	Password               string `json:"password"`
	TokenExpiryMins        int    `json:"token_expiry_in_minutes"`
	RefreshTokenExpiryMins int    `json:"refresh_token_expiry_in_minutes"`
	AccountId              int    `json:"account_id"`
}

type RefreshTokenSchema struct {
	JwtRefreshToken        string `json:"jwt_refresh_token"`
	TokenExpiryMins        int    `json:"token_expiry_in_minutes"`
	RefreshTokenExpiryMins int    `json:"refresh_token_expiry_in_minutes"`
}

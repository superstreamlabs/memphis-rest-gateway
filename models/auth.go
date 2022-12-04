package models

type AuthSchema struct {
	Username        string `json:"username" validate:"required"`
	ConnectionToken string `json:"connection_token" validate:"required"`
}

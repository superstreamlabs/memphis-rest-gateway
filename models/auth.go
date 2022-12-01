package models

type AuthSchema struct {
	Host            string `json:"host" validate:"required"`
	Username        string `json:"username" validate:"required"`
	ConnectionToken string `json:"connection_token" validate:"required"`
}

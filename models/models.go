package models

type RestGwUpdate struct {
	Type   string                 `json:"type"`
	Update map[string]interface{} `json:"update"`
}

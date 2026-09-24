package models

type Guest struct {
	CardID  string `json:"card_id"`
	Balance int64  `json:"balance"`
}

type Member struct {
	CardID   string `json:"card_id"`
	Username string `json:"username"`
	Balance  int64  `json:"balance"`
	Points   int64  `json:"points"`
}

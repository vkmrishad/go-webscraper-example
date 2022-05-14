package models

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorMessage struct {
	Error Error `json:"errors"`
}

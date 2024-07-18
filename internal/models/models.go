package models

type Message struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type Stats struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
	Pending   int `json:"pending"`
}

type Request struct {
	Content string `json:"content"`
}

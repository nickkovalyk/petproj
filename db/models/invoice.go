package models

type Invoice struct {
	ID          int    `json:"id"`
	Body        string `json:"body" db:"body"`
	CreatedDate int64  `json:"createdDate" db:"created_date"`
}
